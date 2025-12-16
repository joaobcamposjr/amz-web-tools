package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"amz-web-tools/backend/internal/config"
	"amz-web-tools/backend/internal/models"
	"amz-web-tools/backend/internal/websocket"

	_ "github.com/lib/pq"          // PostgreSQL driver
	_ "github.com/sijms/go-ora/v2" // Oracle driver
)

type XMLIntegratorService struct {
	oracleDB *sql.DB
	pgDB     *sql.DB
	config   *config.Config
	wsHub    *websocket.Hub
	logs     map[string][]map[string]interface{} // Armazenar logs por process_id
}

type PedidoInfo struct {
	NumPedido       string `json:"num_pedido"`
	NumPrenota      string `json:"num_prenota"`
	NumEnvio        string `json:"num_envio"`
	NumNotaFiscal   string `json:"num_notafiscal"`
	NomTipoEnvio    string `json:"nom_tipoenvio"`
	FlgStatusPedido int    `json:"flg_statuspedido"`
	NomEmpresa      string `json:"nom_empresa"`
}

type OracleXMLData struct {
	Controle      string `json:"controle"`
	Emissao       string `json:"emissao"`
	CodOrcMapa    string `json:"cod_orc_mapa"`
	PedidoExterno string `json:"pedido_externo"`
	XMLNota       string `json:"xml_nota"`
	Status        string `json:"status"`
}

type MLResponse struct {
	Status    string `json:"status"`
	Substatus string `json:"substatus"`
	LeadTime  struct {
		Buffering struct {
			Date string `json:"date"`
		} `json:"buffering"`
	} `json:"lead_time"`
}

type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func NewXMLIntegratorService(cfg *config.Config, wsHub *websocket.Hub) (*XMLIntegratorService, error) {
	log.Println("🔧 Iniciando XMLIntegratorService...")

	var oracleDB *sql.DB
	var pgDB *sql.DB
	var err error

	// Tentar conectar ao Oracle (opcional)
	if cfg.OracleHost != "" && cfg.OracleUser != "" {
		oracleDSN := fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
			cfg.OracleUser, cfg.OraclePassword, cfg.OracleHost, cfg.OraclePort, cfg.OracleService)

		oracleDB, err = sql.Open("oracle", oracleDSN)
		if err != nil {
			log.Printf("⚠️ Erro ao conectar Oracle: %v", err)
			oracleDB = nil
		} else {
			// Test connection with timeout to avoid hanging
			pingDone := make(chan error, 1)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				pingDone <- oracleDB.PingContext(ctx)
			}()

			select {
			case err := <-pingDone:
				if err != nil {
					log.Printf("⚠️ Erro ao ping Oracle: %v", err)
					oracleDB.Close()
					oracleDB = nil
				} else {
					log.Println("✅ Conexão Oracle estabelecida")
				}
			case <-time.After(6 * time.Second):
				log.Printf("⚠️ Timeout ao ping Oracle após 6 segundos")
				oracleDB.Close()
				oracleDB = nil
			}
		}
	} else {
		log.Println("⚠️ Configuração Oracle não disponível")
	}

	// Tentar conectar ao PostgreSQL (opcional para testes)
	if cfg.PGHost != "" && cfg.PGUser != "" {
		pgDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDatabase, cfg.PGSSLMode)

		pgDB, err = sql.Open("postgres", pgDSN)
		if err != nil {
			log.Printf("⚠️ Erro ao conectar PostgreSQL: %v", err)
			pgDB = nil
		} else {
			// Test connection with timeout to avoid hanging
			pingDone := make(chan error, 1)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				pingDone <- pgDB.PingContext(ctx)
			}()

			select {
			case err := <-pingDone:
				if err != nil {
					log.Printf("⚠️ Erro ao ping PostgreSQL: %v", err)
					pgDB.Close()
					pgDB = nil
				} else {
					log.Println("✅ Conexão PostgreSQL estabelecida")
				}
			case <-time.After(6 * time.Second):
				log.Printf("⚠️ Timeout ao ping PostgreSQL após 6 segundos")
				pgDB.Close()
				pgDB = nil
			}
		}
	} else {
		log.Println("⚠️ Configuração PostgreSQL não disponível")
	}

	return &XMLIntegratorService{
		oracleDB: oracleDB,
		pgDB:     pgDB,
		config:   cfg,
		wsHub:    wsHub,
		logs:     make(map[string][]map[string]interface{}),
	}, nil
}

func (s *XMLIntegratorService) getMLToken(nomEmpresa string) (string, error) {
	log.Printf("🔑 Obtendo token ML REAL para empresa: %s", nomEmpresa)

	// Converter nome da empresa para formato da conta
	empresaLower := strings.ToLower(nomEmpresa)
	// Aplicar replace: principal -> amz
	empresaLower = strings.Replace(empresaLower, "principal", "amz", -1)
	tkConta := fmt.Sprintf("tk%s", empresaLower)
	urlToken := fmt.Sprintf("https://imgs-amz.s3.us-east-1.amazonaws.com/tk/%s.html", tkConta)

	log.Printf("🌐 URL Token: %s", urlToken)

	resp, err := http.Get(urlToken)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar token: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %w", err)
	}

	contentStr := string(content)
	tokenIndex := strings.Index(contentStr, "y>") + 2
	if tokenIndex < 2 {
		return "", fmt.Errorf("formato de token inválido")
	}

	tokenMid := contentStr[tokenIndex:]
	tokenIndexEnd := strings.Index(tokenMid, "</")
	if tokenIndexEnd == -1 {
		return "", fmt.Errorf("formato de token inválido")
	}

	tokenML := "APP_USR-" + tokenMid[:tokenIndexEnd]
	log.Printf("✅ Token obtido: %s...", tokenML[:30])

	return tokenML, nil
}

func (s *XMLIntegratorService) sendTelegramMessage(message string) {
	tokenTelegram := "7055914019:AAEJL0zVCZsaGv-RLKVf4VGcFY4-haShUx4"
	chatID := "-4282027613"

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tokenTelegram)

	msg := TelegramMessage{
		ChatID: chatID,
		Text:   message,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ Erro ao serializar mensagem Telegram: %v", err)
		return
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("❌ Erro ao enviar Telegram: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("📱 Mensagem Telegram enviada: %s", message)
}

func (s *XMLIntegratorService) getPedidosFromPostgres(numPedido string) ([]PedidoInfo, error) {
	log.Printf("🔍 Buscando pedido real no PostgreSQL: %s", numPedido)

	query := `
		SELECT 
			num_pedido,
			num_prenota,
			num_envio,
			COALESCE(num_notafiscal, '') as num_notafiscal,
			COALESCE(nom_tipoenvio, 'Mercado Envios') as nom_tipo_envio,
			flg_statuspedido,
			nom_empresa
		FROM integrator.fato_statusvenda 
		WHERE num_pedido = $1
	`

	rows, err := s.pgDB.Query(query, numPedido)
	if err != nil {
		log.Printf("❌ Erro na consulta PostgreSQL: %v", err)
		return nil, fmt.Errorf("erro ao consultar PostgreSQL: %v", err)
	}
	defer rows.Close()

	var pedidos []PedidoInfo
	for rows.Next() {
		var p PedidoInfo
		err := rows.Scan(
			&p.NumPedido,
			&p.NumPrenota,
			&p.NumEnvio,
			&p.NumNotaFiscal,
			&p.NomTipoEnvio,
			&p.FlgStatusPedido,
			&p.NomEmpresa,
		)
		if err != nil {
			log.Printf("❌ Erro ao escanear linha: %v", err)
			continue
		}
		pedidos = append(pedidos, p)
		log.Printf("✅ Pedido encontrado: %s - Prenota: %s - Envio: %s - Nota: %s",
			p.NumPedido, p.NumPrenota, p.NumEnvio, p.NumNotaFiscal)
	}

	if err = rows.Err(); err != nil {
		log.Printf("❌ Erro nas linhas: %v", err)
		return nil, fmt.Errorf("erro ao processar linhas: %v", err)
	}

	log.Printf("📊 Total de pedidos encontrados: %d", len(pedidos))
	return pedidos, nil
}

// checkMercadoLivreStatus verifica o status do pedido no Mercado Livre
func (s *XMLIntegratorService) checkMercadoLivreStatus(numEnvio, token string) (map[string]interface{}, error) {
	if numEnvio == "" {
		return nil, fmt.Errorf("num_envio não pode estar vazio")
	}

	url := fmt.Sprintf("https://api.mercadolibre.com/shipments/%s", numEnvio)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"x-format-new":  "true",
	}

	resp, err := s.makeHTTPRequest("GET", url, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar status ML: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao parsear resposta ML: %v", err)
	}

	return result, nil
}

// makeHTTPRequest faz uma requisição HTTP
func (s *XMLIntegratorService) makeHTTPRequest(method, url string, headers map[string]string, body []byte) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *XMLIntegratorService) getXMLDataFromOracle(codOrcMapa string) (*OracleXMLData, error) {
	if s.oracleDB == nil {
		log.Printf("❌ Oracle não disponível para cod_orc_mapa: %s", codOrcMapa)
		return nil, fmt.Errorf("Oracle não disponível")
	}

	log.Printf("🔍 Buscando dados XML no Oracle para cod_orc_mapa: %s", codOrcMapa)

	query := `
		SELECT 	
			CONTROLE,
			EMISSAO,
			o.COD_ORC_MAPA,
			o.PEDIDO_EXTERNO,
			n.XML_NOTA,
			o.STATUS
		FROM nbs.orc_mapa o
		LEFT JOIN nbs.vendas v 
		ON v.cod_orc_mapa = o.cod_orc_mapa
		LEFT JOIN nbs.nfe_movimento n
		ON v.cod_empresa = n.id_empresa
		AND v.controle = n.numr_nfe
		AND v.serie = n.serie_nbs
		WHERE o.cod_orc_mapa = :1
	`

	var xmlData OracleXMLData
	err := s.oracleDB.QueryRow(query, codOrcMapa).Scan(
		&xmlData.Controle,
		&xmlData.Emissao,
		&xmlData.CodOrcMapa,
		&xmlData.PedidoExterno,
		&xmlData.XMLNota,
		&xmlData.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("⚠️ Nenhum XML encontrado para cod_orc_mapa: %s", codOrcMapa)
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar XML: %w", err)
	}

	log.Printf("✅ XML encontrado - Controle: %s, Status: %s", xmlData.Controle, xmlData.Status)
	return &xmlData, nil
}

func (s *XMLIntegratorService) updatePostgresStatus(numPrenota, numNotaFiscal, xmlData, status string) error {
	if s.pgDB == nil {
		log.Printf("❌ PostgreSQL não disponível para atualização da prenota: %s", numPrenota)
		return fmt.Errorf("PostgreSQL não disponível")
	}

	log.Printf("📝 Atualizando status PostgreSQL para prenota: %s", numPrenota)

	query := `
		UPDATE integrator.fato_statusvenda 
		SET
			dat_notafiscal = NOW(),
			num_xml = $1,
			num_notafiscal = $2,
			flg_statuspedido = $3
		WHERE num_prenota = $4
	`

	_, err := s.pgDB.Exec(query, xmlData, numNotaFiscal, status, numPrenota)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}

	log.Printf("✅ Status PostgreSQL atualizado")
	return nil
}

func (s *XMLIntegratorService) getMLShipmentStatus(numEnvio, token string) (*MLResponse, error) {
	log.Printf("🚚 Verificando status REAL do envio ML: %s", numEnvio)

	url := fmt.Sprintf("https://api.mercadolibre.com/shipments/%s", numEnvio)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"x-format-new":  "true",
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var mlResponse MLResponse
	if err := json.NewDecoder(resp.Body).Decode(&mlResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	log.Printf("📊 Status ML - Status: %s, Substatus: %s", mlResponse.Status, mlResponse.Substatus)
	return &mlResponse, nil
}

func (s *XMLIntegratorService) sendXMLToML(numEnvio, xmlData, token string) error {
	log.Printf("📤 Enviando XML REAL para ML - Envio: %s", numEnvio)

	url := fmt.Sprintf("https://api.mercadolibre.com/shipments/%s/invoice_data?siteId=MLB", numEnvio)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/xml",
		"Accept":        "application/xml",
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(xmlData))
	if err != nil {
		return fmt.Errorf("erro ao criar request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao fazer request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	log.Printf("📤 Resposta ML - Status: %d, Body: %s", resp.StatusCode, string(responseBody))

	if resp.StatusCode == 406 {
		log.Println("✅ XML enviado com sucesso!")
		return nil
	}

	return fmt.Errorf("erro ao enviar XML - Status: %d, Body: %s", resp.StatusCode, string(responseBody))
}

func (s *XMLIntegratorService) ProcessXMLIntegration(numPedido string) (*models.APIResponse, error) {
	log.Printf("🚀 Iniciando processamento XML para pedido: %s", numPedido)

	// Buscar pedidos primeiro para obter a empresa
	pedidos, err := s.getPedidosFromPostgres(numPedido)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao buscar pedidos: %v", err)
		s.sendTelegramMessage(errorMsg)
		return nil, fmt.Errorf(errorMsg)
	}

	if len(pedidos) == 0 {
		msg := fmt.Sprintf("⚠️ Nenhum pedido encontrado para: %s", numPedido)
		log.Println(msg)

		// Criar logs detalhados para armazenamento
		processLogs := []map[string]interface{}{
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "info",
				"step":      "Início",
				"message":   fmt.Sprintf("Iniciando processamento XML para pedido: %s", numPedido),
			},
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "warning",
				"step":      "Busca Pedidos",
				"message":   fmt.Sprintf("Nenhum pedido encontrado para: %s - Verifique se o pedido existe na tabela integrator.fato_statusvenda", numPedido),
			},
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "error",
				"step":      "Validação",
				"message":   "Motivo: Pedido não encontrado na base de dados PostgreSQL ou não está com status 1 ou 2",
			},
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "info",
				"step":      "Conclusão",
				"message":   "Processamento concluído - Total: 0 | Sucessos: 0 | Erros: 1",
			},
		}

		// Armazenar logs
		s.logs[numPedido] = processLogs

		// Enviar logs via WebSocket
		if s.wsHub != nil {
			for _, logEntry := range processLogs {
				s.wsHub.BroadcastLog(websocket.LogMessage{
					Type:      "log",
					Timestamp: logEntry["timestamp"].(string),
					Level:     logEntry["level"].(string),
					Step:      logEntry["step"].(string),
					Message:   logEntry["message"].(string),
					ProcessID: numPedido,
				})
			}
		}

		return &models.APIResponse{
			Success: true,
			Message: "Processamento XML concluído",
			Data: map[string]interface{}{
				"total_processed": 0,
				"success_count":   0,
				"error_count":     1,
				"results":         []map[string]interface{}{},
				"logs":            processLogs,
			},
		}, nil
	}

	// Obter token ML usando a empresa do primeiro pedido
	nomEmpresa := pedidos[0].NomEmpresa
	token, err := s.getMLToken(nomEmpresa)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao obter token ML para empresa %s: %v", nomEmpresa, err)
		s.sendTelegramMessage(errorMsg)
		return nil, fmt.Errorf(errorMsg)
	}

	var results []map[string]interface{}
	var logs []map[string]interface{}
	successCount := 0
	errorCount := 0

	// Adicionar log inicial
	logs = append(logs, map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     "info",
		"step":      "Início",
		"message":   fmt.Sprintf("Iniciando processamento XML para pedido: %s", numPedido),
	})

	for _, pedido := range pedidos {
		log.Printf("📦 Processando pedido: %s", pedido.NumPedido)

		// Log início do processamento do pedido
		logs = append(logs, map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "info",
			"step":      "Processamento",
			"message":   fmt.Sprintf("Processando pedido: %s", pedido.NumPedido),
		})

		// Log busca no Oracle
		logs = append(logs, map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "info",
			"step":      "Oracle",
			"message":   fmt.Sprintf("Buscando dados XML no Oracle para prenota: %s", pedido.NumPrenota),
		})

		// Buscar dados XML
		xmlData, err := s.getXMLDataFromOracle(pedido.NumPrenota)
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Erro ao buscar XML para pedido %s: %v", pedido.NumPedido, err)
			log.Println(errorMsg)
			s.sendTelegramMessage(errorMsg)
			errorCount++

			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "error",
				"step":      "Oracle",
				"message":   fmt.Sprintf("Erro ao buscar XML para prenota %s: %v", pedido.NumPrenota, err),
			})

			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "error",
				"step":      "Validação",
				"message":   "Motivo: Erro de conexão com Oracle ou dados não encontrados na tabela nbs.orc_mapa",
			})
			continue
		}

		if xmlData == nil {
			msg := fmt.Sprintf("⚠️ Pedido %s ainda não possui XML", pedido.NumPedido)
			log.Println(msg)
			s.sendTelegramMessage(msg)
			errorCount++

			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "warning",
				"step":      "Oracle",
				"message":   fmt.Sprintf("XML não encontrado para prenota %s", pedido.NumPrenota),
			})

			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "error",
				"step":      "Validação",
				"message":   "Motivo: Pedido ainda não possui XML no Oracle ou dados não encontrados na tabela nbs.nfe_movimento",
			})
			continue
		}

		// Verificar se XML está vazio
		if xmlData.XMLNota == "" {
			msg := fmt.Sprintf("⚠️ Pedido %s possui XML vazio", pedido.NumPedido)
			log.Println(msg)
			s.sendTelegramMessage(msg)
			errorCount++

			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "warning",
				"step":      "Validação XML",
				"message":   fmt.Sprintf("XML encontrado mas está vazio para prenota %s", pedido.NumPrenota),
			})

			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "error",
				"step":      "Validação",
				"message":   "Motivo: XML não encontrado ou está vazio na tabela nbs.nfe_movimento",
			})
			continue
		}

		logs = append(logs, map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "success",
			"step":      "Oracle",
			"message":   fmt.Sprintf("XML encontrado - Controle: %s, Status: %s", xmlData.Controle, xmlData.Status),
		})

		// Log atualização PostgreSQL
		logs = append(logs, map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "info",
			"step":      "PostgreSQL",
			"message":   fmt.Sprintf("Atualizando status PostgreSQL para prenota: %s", pedido.NumPrenota),
		})

		// Atualizar status no PostgreSQL
		err = s.updatePostgresStatus(pedido.NumPrenota, xmlData.Controle, xmlData.XMLNota, "2")
		if err != nil {
			log.Printf("❌ Erro ao atualizar status PostgreSQL: %v", err)

			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "error",
				"step":      "PostgreSQL",
				"message":   fmt.Sprintf("Erro ao atualizar status: %v", err),
			})
			continue
		}

		logs = append(logs, map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "success",
			"step":      "PostgreSQL",
			"message":   "Status PostgreSQL atualizado com sucesso",
		})

		// Log verificação status ML
		logs = append(logs, map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "info",
			"step":      "Mercado Livre",
			"message":   fmt.Sprintf("Verificando status do envio ML: %s", pedido.NumEnvio),
		})

		// Verificar status do envio ML
		mlStatus, err := s.getMLShipmentStatus(pedido.NumEnvio, token)
		if err != nil {
			log.Printf("❌ Erro ao verificar status ML: %v", err)

			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "error",
				"step":      "Mercado Livre",
				"message":   fmt.Sprintf("Erro ao verificar status ML: %v", err),
			})
			continue
		}

		logs = append(logs, map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "success",
			"step":      "Mercado Livre",
			"message":   fmt.Sprintf("Status ML obtido - Status: %s, Substatus: %s", mlStatus.Status, mlStatus.Substatus),
		})

		// Processar baseado no status
		if mlStatus.Status == "ready_to_ship" && mlStatus.Substatus == "invoice_pending" {
			log.Printf("📤 Enviando XML para ML - Pedido: %s", pedido.NumPedido)

			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "info",
				"step":      "Envio XML",
				"message":   fmt.Sprintf("Enviando XML para ML - Pedido: %s", pedido.NumPedido),
			})

			err = s.sendXMLToML(pedido.NumEnvio, xmlData.XMLNota, token)
			if err != nil {
				log.Printf("❌ Falha ao enviar XML do Pedido %s: %v", pedido.NumPedido, err)
				errorCount++

				logs = append(logs, map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"level":     "error",
					"step":      "Envio XML",
					"message":   fmt.Sprintf("Falha ao enviar XML: %v", err),
				})
			} else {
				log.Printf("✅ %s (MeLi) - Pedido %s, NF: %s - XML enviado com sucesso!",
					strings.ToUpper(nomEmpresa), pedido.NumPedido, xmlData.Controle)

				logs = append(logs, map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"level":     "success",
					"step":      "Envio XML",
					"message":   "XML enviado com sucesso para o Mercado Livre",
				})

				// Atualizar status para 3 (concluído)
				s.updatePostgresStatus(pedido.NumPrenota, xmlData.Controle, xmlData.XMLNota, "3")
				successCount++
			}
		} else if mlStatus.Status == "pending" && mlStatus.Substatus == "buffered" {
			log.Printf("⏰ %s (MeLi) - Pedido %s, NF: %s - Entrega Agendada para: %s | Status: %s | Substatus: %s",
				strings.ToUpper(nomEmpresa), pedido.NumPedido, xmlData.Controle,
				mlStatus.LeadTime.Buffering.Date, mlStatus.Status, mlStatus.Substatus)
		} else {
			// Envio Flex ou outros casos
			if pedido.NumNotaFiscal != "" {
				log.Printf("🚚 %s (MeLi) - Pedido %s, NF: %s - Envio Flex!",
					strings.ToUpper(nomEmpresa), pedido.NumPedido, xmlData.Controle)
				s.updatePostgresStatus(pedido.NumPrenota, xmlData.Controle, xmlData.XMLNota, "3")
				successCount++
			} else {
				log.Printf("⏳ %s (MeLi) - Pedido %s - Aguardando Nota Fiscal - Envio Flex!",
					strings.ToUpper(nomEmpresa), pedido.NumPedido)
			}
		}

		results = append(results, map[string]interface{}{
			"pedido":      pedido.NumPedido,
			"prenota":     pedido.NumPrenota,
			"envio":       pedido.NumEnvio,
			"nota_fiscal": xmlData.Controle,
			"status":      mlStatus.Status,
			"substatus":   mlStatus.Substatus,
		})
	}

	// Resumo final unificado
	totalProcessed := len(pedidos)

	// Criar mensagem detalhada com informações dos pedidos processados
	var details []string
	for _, result := range results {
		status := result["status"].(string)
		substatus := result["substatus"].(string)
		pedido := result["pedido"].(string)
		notaFiscal := result["nota_fiscal"].(string)

		if status == "ready_to_ship" {
			details = append(details, fmt.Sprintf("✅ Pedido %s (NF: %s) - XML enviado", pedido, notaFiscal))
		} else if status == "pending" && substatus == "buffered" {
			details = append(details, fmt.Sprintf("⏰ Pedido %s (NF: %s) - Entrega agendada", pedido, notaFiscal))
		} else if status == "shipped" {
			details = append(details, fmt.Sprintf("🚚 Pedido %s (NF: %s) - Envio Flex", pedido, notaFiscal))
		} else {
			details = append(details, fmt.Sprintf("❌ Pedido %s (NF: %s) - Erro no processamento", pedido, notaFiscal))
		}
	}

	// Mensagem unificada
	finalMsg := fmt.Sprintf("📊 %s (MeLi) - Processamento XML Concluído\n\n📈 Resumo:\n• Total processado: %d\n• Sucessos: %d\n• Erros: %d\n\n📋 Detalhes:\n%s",
		strings.ToUpper(nomEmpresa), totalProcessed, successCount, errorCount, strings.Join(details, "\n"))

	log.Println("📊 Enviando resumo unificado para Telegram")
	s.sendTelegramMessage(finalMsg)

	log.Printf("🔍 Debug - results length: %d, results: %v", len(results), results)
	log.Printf("🔍 Debug - logs length: %d, logs: %v", len(logs), logs)

	// Log final
	logs = append(logs, map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     "info",
		"step":      "Conclusão",
		"message":   fmt.Sprintf("Processamento concluído - Total: %d | Sucessos: %d | Erros: %d", totalProcessed, successCount, errorCount),
	})

	return &models.APIResponse{
		Success: true,
		Message: "Processamento XML concluído",
		Data: map[string]interface{}{
			"total_processed": totalProcessed,
			"success_count":   successCount,
			"error_count":     errorCount,
			"results":         results,
			"logs":            logs,
		},
	}, nil
}

// GetLogs returns logs for a specific process
func (s *XMLIntegratorService) GetLogs(processID string) ([]map[string]interface{}, error) {
	if logs, exists := s.logs[processID]; exists {
		return logs, nil
	}
	return []map[string]interface{}{}, nil
}

// GetPostgresDB returns the PostgreSQL database connection
func (s *XMLIntegratorService) GetPostgresDB() *sql.DB {
	return s.pgDB
}

func (s *XMLIntegratorService) Close() error {
	if s.oracleDB != nil {
		s.oracleDB.Close()
	}
	if s.pgDB != nil {
		s.pgDB.Close()
	}
	return nil
}
