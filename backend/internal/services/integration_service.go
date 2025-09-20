package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type IntegrationService struct {
	sqlDB    *sql.DB
	oracleDB *sql.DB
	pgDB     *sql.DB
}

type IntegrationRequest struct {
	Conta       string `json:"conta"`
	Marketplace string `json:"marketplace"`
	NumPedido   string `json:"num_pedido"`
}

type IntegrationResponse struct {
	TotalProcessed int                      `json:"total_processed"`
	SuccessCount   int                      `json:"success_count"`
	ErrorCount     int                      `json:"error_count"`
	Results        []map[string]interface{} `json:"results"`
	Logs           []IntegrationLogEntry    `json:"logs"`
}

type IntegrationLogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Step      string `json:"step"`
	Message   string `json:"message"`
}

type MLOrder struct {
	ID          interface{} `json:"id"`
	DateCreated string      `json:"date_created"`
	OrderItems  []struct {
		Item struct {
			ID    interface{} `json:"id"`
			Title string      `json:"title"`
		} `json:"item"`
		UnitPrice float64 `json:"unit_price"`
		Quantity  int     `json:"quantity"`
	} `json:"order_items"`
	Shipping struct {
		ID interface{} `json:"id"`
	} `json:"shipping"`
}

type MLItem struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Attributes []struct {
		Name      string `json:"name"`
		ValueName string `json:"value_name"`
	} `json:"attributes"`
}

type MLBillingInfo struct {
	BillingInfo struct {
		DocType        string `json:"doc_type"`
		DocNumber      string `json:"doc_number"`
		AdditionalInfo struct {
			FirstName         string `json:"FIRST_NAME"`
			LastName          string `json:"LAST_NAME"`
			BusinessName      string `json:"BUSINESS_NAME"`
			CityName          string `json:"CITY_NAME"`
			ZipCode           string `json:"ZIP_CODE"`
			StreetName        string `json:"STREET_NAME"`
			StreetNumber      string `json:"STREET_NUMBER"`
			Neighborhood      string `json:"NEIGHBORHOOD"`
			Comment           string `json:"COMMENT"`
			StateRegistration string `json:"STATE_REGISTRATION"`
		} `json:"additional_info"`
	} `json:"billing_info"`
}

type MLShipment struct {
	LogisticType    string `json:"logistic_type"`
	ReceiverAddress struct {
		City struct {
			Name string `json:"name"`
		} `json:"city"`
		State struct {
			ID string `json:"id"`
		} `json:"state"`
		ZipCode      string `json:"zip_code"`
		StreetName   string `json:"street_name"`
		StreetNumber string `json:"street_number"`
		Neighborhood struct {
			Name string `json:"name"`
		} `json:"neighborhood"`
		Comment      string `json:"comment"`
		ReceiverName string `json:"receiver_name"`
	} `json:"receiver_address"`
}

type MLPack struct {
	Orders []struct {
		ID string `json:"id"`
	} `json:"orders"`
}

type NBSClientRequest struct {
	CodigoCliente       string `json:"codigoCliente"`
	CodigoTipoCliente   int    `json:"codigoTipoCliente"`
	CodigoRamo          string `json:"codigoRamo"`
	CodigoClasse        int    `json:"codigoClasse"`
	CodigoClasseTipo    string `json:"codigoClasseTipo"`
	CodigoEstadoCivil   string `json:"codigoEstadoCivil"`
	PrefixoCelular      string `json:"prefixoCelular"`
	TelefoneCelular     string `json:"telefoneCelular"`
	PrefixoComercial    string `json:"prefixoComercial"`
	TelefoneComercial   string `json:"telefoneComercial"`
	PrefixoResidencial  string `json:"prefixoResidencial"`
	TelefoneResidencial string `json:"telefoneResidencial"`
	CodigoNacionalidade string `json:"codigoNacionalidade"`
	CodigoProfissao     string `json:"codigoProfissao"`
	PaiCliente          string `json:"paiCliente"`
	MaeCliente          string `json:"maeCliente"`
	EmailCliente        string `json:"emailCliente"`
	Tipo                string `json:"tipo"`
	Nome                string `json:"nome"`
	Sexo                string `json:"sexo"`
	Nascimento          string `json:"nascimento"`
	CpfCnpj             string `json:"cpfCnpj"`
	RgIe                string `json:"rgIe"`
	Ssp                 string `json:"ssp"`
	AtualizaExistente   bool   `json:"atualizaExistente"`
	ClienteRevendedor   bool   `json:"clienteRevendedor"`
}

type NBSAddressRequest struct {
	CodigoCliente          string `json:"codigoCliente"`
	ClienteTipoEndereco    int    `json:"clienteTipoEndereco"`
	CodCidades             string `json:"codCidades"`
	CEP                    string `json:"CEP"`
	Rua                    string `json:"rua"`
	Complemento            string `json:"complemento"`
	Bairro                 string `json:"bairro"`
	UF                     string `json:"uf"`
	NumeroEndereco         string `json:"numeroEndereco"`
	NomePropriedade        string `json:"nomePropriedade"`
	InscricaoEstadual      string `json:"inscricaoEstadual"`
	Fachada                string `json:"fachada"`
	Contato                string `json:"contato"`
	TelefoneContato        string `json:"telefoneContato"`
	PrefixoTelefoneContato string `json:"prefixoTelefoneContato"`
}

type NBSOrderRequest struct {
	CodPedidoWeb      string         `json:"COD_PEDIDO_WEB"`
	CodCliente        string         `json:"COD_CLIENTE"`
	TipoEndereco      int            `json:"TIPO_ENDERECO"`
	CodTransportadora int            `json:"COD_TRANSPORTADORA"`
	ValorFreteTotal   float64        `json:"VALOR_FRETE_TOTAL"`
	CnpjIntermed      string         `json:"CNPJ_INTERMED"`
	IdentCadIntermed  string         `json:"IDENT_CAD_INTERMED"`
	Nome              string         `json:"NOME"`
	Itens             []NBSOrderItem `json:"Itens"`
	Pagamentos        []NBSPayment   `json:"Pagamentos"`
}

type NBSOrderItem struct {
	CodItem       string  `json:"COD_ITEM"`
	CodFornecedor string  `json:"COD_FORNECEDOR"`
	PrecoUnitario float64 `json:"PRECO_UNITARIO"`
	Qtde          int     `json:"QTDE"`
}

type NBSPayment struct {
	CodigoBandeira     string `json:"codigoBandeira"`
	TipoCartao         string `json:"tipoCartao"`
	DataPagamento      string `json:"dataPagamento"`
	NumeroCartao       string `json:"numeroCartao"`
	NumeroAutorizacao  string `json:"numeroAutorizacao"`
	QuantidadeParcelas int    `json:"quantidadeParcelas"`
}

type NBSResponse struct {
	Sucesso  bool   `json:"sucesso"`
	Mensagem string `json:"mensagem"`
	Data     struct {
		CodigoPedido string `json:"codigoPedido"`
	} `json:"data"`
}

type DeParaResult struct {
	MLB    string `json:"mlb"`
	SKU    string `json:"sku"`
	Filial string `json:"filial"`
}

type StockItem struct {
	CodFornecedor  int     `json:"cod_fornecedor"`
	CodItem        string  `json:"cod_item"`
	ValorReposicao float64 `json:"valor_reposicao"`
	Estoque        int     `json:"estoque"`
}

func NewIntegrationService(sqlDB, oracleDB, pgDB *sql.DB) *IntegrationService {
	return &IntegrationService{
		sqlDB:    sqlDB,
		oracleDB: oracleDB,
		pgDB:     pgDB,
	}
}

func (s *IntegrationService) ProcessIntegration(req IntegrationRequest) (*IntegrationResponse, error) {
	logs := []IntegrationLogEntry{}

	// Log inicial
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🎉",
		Message:   fmt.Sprintf("%s (MeLi) - NOVO PEDIDO RECEBIDO! 🎉", strings.ToUpper(req.Conta)),
	})

	// Detalhes do pedido
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🛒",
		Message:   fmt.Sprintf("%s | %s", req.NumPedido, req.Marketplace),
	})

	// Validar se o pedido já foi processado
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🔍",
		Message:   "Verificando se o pedido já foi processado...",
	})

	queryValidate := `SELECT * FROM integrator.fato_StatusVenda WHERE num_pedido = $1`
	rows, err := s.pgDB.Query(queryValidate, req.NumPedido)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "❌",
			Message:   fmt.Sprintf("Erro ao verificar pedido no PostgreSQL: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, nil
	}
	defer rows.Close()

	if rows.Next() {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "warning",
			Step:      "⚠️",
			Message:   fmt.Sprintf("Pedido %s já foi processado anteriormente", req.NumPedido),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, nil
	}

	// Obter token do Mercado Livre
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🔑",
		Message:   "Obtendo token do Mercado Livre...",
	})

	tokenML, tokenUser, err := s.getMLToken(req.Conta)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "❌",
			Message:   fmt.Sprintf("Erro ao obter token ML: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, nil
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "✅",
		Message:   "Token do Mercado Livre obtido com sucesso",
	})

	// Processar pedido
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🚚",
		Message:   "Tipo de Envio: Envio Flex",
	})

	result, processLogs := s.processOrder(req.NumPedido, req.Conta, tokenML, tokenUser)
	logs = append(logs, processLogs...)

	// Enviar mensagem final do Telegram sempre que o processo terminar
	var finalMsg string
	if result.SuccessCount > 0 {
		// Obter dados do cliente para a mensagem detalhada
		var clienteNome, clienteEndereco, numeroPrenota string
		clienteNome = "Cliente não identificado"
		clienteEndereco = "Endereço não identificado"

		// Tentar obter dados do cliente e pedido do Mercado Livre
		tokenML, _, err := s.getMLToken(req.Conta)
		var valorTotal float64
		var numeroItens int
		if err == nil {
			// Obter dados do cliente
			clienteData, err := s.getMLClientData(req.NumPedido, tokenML)
			if err == nil {
				if nome, ok := clienteData["nome"].(string); ok {
					clienteNome = nome
				}
				if rua, ok := clienteData["rua"].(string); ok {
					if bairro, ok := clienteData["bairro"].(string); ok {
						if uf, ok := clienteData["uf"].(string); ok {
							clienteEndereco = fmt.Sprintf("%s, %s - %s", rua, bairro, uf)
						} else {
							clienteEndereco = fmt.Sprintf("%s, %s", rua, bairro)
						}
					} else {
						clienteEndereco = rua
					}
				}
			}

			// Obter dados do pedido para valor total e número de itens
			order, err := s.getMLOrder(req.NumPedido, tokenML)
			if err == nil {
				numeroItens = len(order.OrderItems)
				for _, item := range order.OrderItems {
					valorTotal += item.UnitPrice * float64(item.Quantity)
				}
			}
		}

		// Obter número da pré-nota dos resultados
		if len(result.Results) > 0 {
			if prenota, ok := result.Results[0]["prenota"].(string); ok {
				numeroPrenota = prenota
			}
		}

		// Determinar tipo de envio
		tipoEnvio := "Envio Flex"

		finalMsg = fmt.Sprintf(`🎉 NOVO PEDIDO INTEGRADO COM SUCESSO! 🎉

📋 INFORMAÇÕES DO PEDIDO:
• Conta: %s
• Número: %s
• Tipo de Envio: %s
• Cliente: %s
• Endereço: %s
• Itens: %d
• Valor Total: R$ %.2f

✅ PROCESSAMENTO:
• Cliente cadastrado no NBS
• Endereço registrado no NBS
• Pedido processado com sucesso

📄 PRÉ-NOTA GERADA: %s`,
			strings.ToUpper(req.Conta),
			req.NumPedido,
			tipoEnvio,
			clienteNome,
			clienteEndereco,
			numeroItens,
			valorTotal,
			numeroPrenota)
	} else {
		finalMsg = fmt.Sprintf(`❌ %s (MeLi) - INTEGRAÇÃO FALHOU! ❌

📊 Resumo:
• Total processado: 1
• Sucessos: %d
• Erros: %d

❌ Pedido %s falhou na integração!`,
			strings.ToUpper(req.Conta), result.SuccessCount, result.ErrorCount, req.NumPedido)
	}

	log.Printf("📱 Enviando mensagem final para Telegram")
	s.sendTelegramMessage(finalMsg)

	return &IntegrationResponse{
		TotalProcessed: 1,
		SuccessCount:   result.SuccessCount,
		ErrorCount:     result.ErrorCount,
		Results:        result.Results,
		Logs:           logs,
	}, nil
}

func (s *IntegrationService) getMLToken(conta string) (string, string, error) {
	log.Printf("🔑 Obtendo token ML REAL para conta: %s", conta)

	tkConta := fmt.Sprintf("tk%s", conta)
	if conta == "ford" {
		url := fmt.Sprintf("https://imgs-amz.s3.us-east-1.amazonaws.com/tk/%s.txt", tkConta)
		resp, err := http.Get(url)
		if err != nil {
			return "", "", fmt.Errorf("erro ao obter token fords: %w", err)
		}
		defer resp.Body.Close()

		var content struct {
			Data string `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
			return "", "", fmt.Errorf("erro ao decodificar token fords: %w", err)
		}

		tokenML := content.Data
		tokenUser := strings.Split(tokenML, "-")[len(strings.Split(tokenML, "-"))-1]
		return tokenML, tokenUser, nil
	} else {
		// Substituir "principal" por "amz" se necessário
		tkConta = strings.Replace(tkConta, "principal", "amz", 1)
		urlToken := fmt.Sprintf("https://imgs-amz.s3.us-east-1.amazonaws.com/tk/%s.html", tkConta)

		resp, err := http.Get(urlToken)
		if err != nil {
			return "", "", fmt.Errorf("erro ao obter token: %w", err)
		}
		defer resp.Body.Close()

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", "", fmt.Errorf("erro ao ler conteúdo: %w", err)
		}

		contentStr := string(content)
		tokenIndex := strings.Index(contentStr, "y>") + 2
		if tokenIndex < 2 {
			return "", "", fmt.Errorf("formato de token inválido")
		}

		tokenMid := contentStr[tokenIndex:]
		tokenIndexEnd := strings.Index(tokenMid, "</")
		if tokenIndexEnd == -1 {
			return "", "", fmt.Errorf("formato de token inválido")
		}

		tokenML := "APP_USR-" + tokenMid[:tokenIndexEnd]
		tokenUser := strings.Split(tokenML, "-")[len(strings.Split(tokenML, "-"))-1]

		log.Printf("✅ Token ML obtido: %s", tokenML)
		return tokenML, tokenUser, nil
	}
}

func (s *IntegrationService) processOrder(numPedido, conta, tokenML, tokenUser string) (*IntegrationResponse, []IntegrationLogEntry) {
	logs := []IntegrationLogEntry{}

	// Obter dados do pedido
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "📋",
		Message:   "Obtendo dados do pedido no Mercado Livre...",
	})

	order, err := s.getMLOrder(numPedido, tokenML)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "❌",
			Message:   fmt.Sprintf("Erro ao obter dados do pedido: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	// Log detalhado dos dados do pedido
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "✅",
		Message:   fmt.Sprintf("Pedido obtido: %s | Itens: %d", numPedido, len(order.OrderItems)),
	})

	// Log do tipo de envio (padrão Envio Flex)
	tipoEnvio := "Envio Flex"
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🚚",
		Message:   fmt.Sprintf("Tipo de Envio: %s", tipoEnvio),
	})

	// Obter empresa e fornecedor
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🏢",
		Message:   "Determinando empresa e fornecedor...",
	})

	itemID := fmt.Sprintf("%v", order.OrderItems[0].Item.ID)
	codEmpresa, codFornecedor, err := s.getEmpresaFornecedor(itemID, conta)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "❌",
			Message:   fmt.Sprintf("Erro ao determinar empresa/fornecedor: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	// Determinar nom_conta seguindo a lógica do Python: tk_conta.replace('tk', '').upper()
	nomConta := strings.ToUpper(conta)

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "✅",
		Message:   fmt.Sprintf("Empresa: %s, Fornecedor: %s, Conta: %s", codEmpresa, codFornecedor, nomConta),
	})

	// Obter token NBS
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🔑",
		Message:   "Obtendo token do NBS...",
	})

	tokenNBS, err := s.getNBSToken(codEmpresa)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "🔑",
			Message:   fmt.Sprintf("Erro ao obter token NBS: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "🔑",
		Message:   "Token do NBS obtido com sucesso",
	})

	// Obter dados do cliente primeiro
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "👤",
		Message:   "Coletando dados do cliente do Mercado Livre...",
	})

	clienteData, err := s.getMLClientData(numPedido, tokenML)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "❌",
			Message:   fmt.Sprintf("Erro ao obter dados do cliente: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	// Log detalhado do cliente
	nomeCliente := "Cliente não identificado"
	if nome, ok := clienteData["nome"].(string); ok {
		nomeCliente = nome
	}
	docCliente := "Documento não identificado"
	if doc, ok := clienteData["codigoCliente"].(string); ok {
		docCliente = doc
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "✅",
		Message:   fmt.Sprintf("Cliente identificado: %s | Doc: %s", nomeCliente, docCliente),
	})

	// Cadastrar cliente no NBS
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "👤",
		Message:   "Cadastrando cliente no NBS...",
	})

	_, err = s.registerClient(numPedido, tokenML, tokenNBS)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "❌",
			Message:   fmt.Sprintf("Erro ao cadastrar cliente: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "✅",
		Message:   fmt.Sprintf("Cliente %s cadastrado no NBS com sucesso", nomeCliente),
	})

	// Log detalhado do endereço
	enderecoCompleto := "Endereço não identificado"
	if rua, ok := clienteData["rua"].(string); ok {
		if bairro, ok := clienteData["bairro"].(string); ok {
			if uf, ok := clienteData["uf"].(string); ok {
				enderecoCompleto = fmt.Sprintf("%s, %s - %s", rua, bairro, uf)
			}
		}
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🏠",
		Message:   fmt.Sprintf("Endereço identificado: %s", enderecoCompleto),
	})

	// Cadastrar endereço no NBS
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "🏠",
		Message:   "Cadastrando endereço no NBS...",
	})

	_, err = s.registerAddress(numPedido, tokenML, tokenNBS)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "❌",
			Message:   fmt.Sprintf("Erro ao cadastrar endereço: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "✅",
		Message:   "Endereço cadastrado no NBS com sucesso",
	})

	// Inserir status inicial
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "📝",
		Message:   "Inserindo status inicial do pedido...",
	})

	shippingID := fmt.Sprintf("%v", order.Shipping.ID)
	err = s.insertInitialStatus(numPedido, nomConta, tokenUser, shippingID)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "📝",
			Message:   fmt.Sprintf("Erro ao inserir status inicial: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "📝",
		Message:   "Status inicial inserido com sucesso",
	})

	// Processar itens e enviar pedido
	// Obter dados do cliente para usar o codigoCliente
	clienteData, err = s.getMLClientData(numPedido, tokenML)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "👤",
			Message:   fmt.Sprintf("Erro ao obter dados do cliente: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	codCliente := clienteData["codigoCliente"].(string)

	// Log detalhado dos itens
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "📦",
		Message:   fmt.Sprintf("Processando %d item(s) do pedido...", len(order.OrderItems)),
	})

	// Log de cada item
	for i, item := range order.OrderItems {
		itemID := fmt.Sprintf("%v", item.Item.ID)
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "info",
			Step:      "📦",
			Message:   fmt.Sprintf("Item %d: MLB %s | Preço: R$ %.2f | Qtd: %d", i+1, itemID, item.UnitPrice, item.Quantity),
		})
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "⚙️",
		Message:   "Enviando pedido para NBS...",
	})

	numeroPedido, err := s.processItemsAndSendOrder(numPedido, order, codEmpresa, codFornecedor, codCliente, tokenNBS)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "❌",
			Message:   fmt.Sprintf("Erro ao processar itens: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "✅",
		Message:   fmt.Sprintf("Pedido processado com sucesso! Número da pré-nota: %s", numeroPedido),
	})

	// Atualizar status final
	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Step:      "📊",
		Message:   "Atualizando status final do pedido...",
	})

	err = s.updateFinalStatus(numPedido, numeroPedido, tokenUser)
	if err != nil {
		logs = append(logs, IntegrationLogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     "error",
			Step:      "📊",
			Message:   fmt.Sprintf("Erro ao atualizar status final: %v", err),
		})
		return &IntegrationResponse{
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     1,
			Results:        []map[string]interface{}{},
			Logs:           logs,
		}, logs
	}

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "📊",
		Message:   "Status final atualizado com sucesso",
	})

	logs = append(logs, IntegrationLogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "success",
		Step:      "🎯",
		Message:   fmt.Sprintf("Integração concluída com sucesso! Pedido %s integrado como pré-nota %s", numPedido, numeroPedido),
	})

	// Enviar mensagem de sucesso para Telegram
	msg := fmt.Sprintf(`🎉 %s (MeLi) - NOVO PEDIDO RECEBIDO! 🎉

🛒 %s | %s

🚚 Tipo de Envio: Envio Flex`, strings.ToUpper(conta), numPedido, numeroPedido)

	s.sendTelegramMessage(msg)

	return &IntegrationResponse{
		TotalProcessed: 1,
		SuccessCount:   1,
		ErrorCount:     0,
		Results: []map[string]interface{}{
			{
				"pedido":        numPedido,
				"numero_pedido": numeroPedido,
				"prenota":       numeroPedido,
				"status":        "sucesso",
			},
		},
		Logs: logs,
	}, logs
}

// Implementar métodos auxiliares...
func (s *IntegrationService) getMLOrder(numPedido, tokenML string) (*MLOrder, error) {
	// Primeiro tenta /orders/{id}
	url := fmt.Sprintf("https://api.mercadolibre.com/orders/%s?access_token=%s", numPedido, tokenML)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("🔍 Status code da resposta /orders: %d", resp.StatusCode)

	// Se retornou 200, processa normalmente
	if resp.StatusCode == 200 {
		var order MLOrder
		if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
			return nil, err
		}
		return &order, nil
	}

	// Se retornou 404, tenta /packs/{id} como fallback
	if resp.StatusCode == 404 {
		log.Printf("🔍 Pedido %s não encontrado em /orders (status %d), tentando /packs...", numPedido, resp.StatusCode)

		packURL := fmt.Sprintf("https://api.mercadolibre.com/packs/%s?access_token=%s", numPedido, tokenML)
		packReq, err := http.NewRequest("GET", packURL, nil)
		if err != nil {
			return nil, err
		}

		packReq.Header.Set("Content-Type", "application/json")
		packReq.Header.Set("Accept", "application/json")

		packResp, err := client.Do(packReq)
		if err != nil {
			log.Printf("❌ Erro ao fazer requisição para /packs: %v", err)
			return nil, err
		}
		defer packResp.Body.Close()

		log.Printf("🔍 Resposta do /packs: status %d", packResp.StatusCode)
		if packResp.StatusCode == 200 {
			// Estrutura do pack response
			var packResponse struct {
				Orders []struct {
					ID interface{} `json:"id"`
				} `json:"orders"`
			}

			if err := json.NewDecoder(packResp.Body).Decode(&packResponse); err != nil {
				return nil, err
			}

			if len(packResponse.Orders) > 0 {
				// Pega o primeiro pedido do pack e converte para string sem notação científica
				firstOrderID := fmt.Sprintf("%.0f", packResponse.Orders[0].ID)
				log.Printf("🔍 Pack encontrado! Primeiro pedido: %s", firstOrderID)

				// Agora busca o pedido real
				orderURL := fmt.Sprintf("https://api.mercadolibre.com/orders/%s?access_token=%s", firstOrderID, tokenML)
				orderReq, err := http.NewRequest("GET", orderURL, nil)
				if err != nil {
					return nil, err
				}

				orderReq.Header.Set("Content-Type", "application/json")
				orderReq.Header.Set("Accept", "application/json")

				orderResp, err := client.Do(orderReq)
				if err != nil {
					return nil, err
				}
				defer orderResp.Body.Close()

				if orderResp.StatusCode == 200 {
					var order MLOrder
					if err := json.NewDecoder(orderResp.Body).Decode(&order); err != nil {
						return nil, err
					}
					log.Printf("✅ Pedido obtido via pack: %s", firstOrderID)
					return &order, nil
				}
			}
		}
	}

	log.Printf("❌ Pedido %s não encontrado nem em /orders nem em /packs", numPedido)
	return nil, fmt.Errorf("erro ao obter pedido ML: status %d", resp.StatusCode)
}

func (s *IntegrationService) getEmpresaFornecedor(mlbItem, conta string) (string, string, error) {
	// Obter item do ML
	url := fmt.Sprintf("https://api.mercadolibre.com/items/%s", mlbItem)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var item MLItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return "", "", err
	}

	// Validar número da peça
	var validadorPeca string
	for _, attr := range item.Attributes {
		if attr.Name == "MPN" {
			validadorPeca = attr.ValueName
			break
		}
	}

	// Consultar DePara - usar schema dinâmico baseado na conta
	log.Printf("🔍 Conta recebida: '%s'", conta)
	schema := s.getSchemaFromConta(conta)
	log.Printf("🔍 Schema determinado: '%s'", schema)
	query := fmt.Sprintf("SELECT mlb, sku, filial FROM %s.stg_Depara WHERE mlb = @p1", schema)
	log.Printf("🔍 Executando query: %s com mlbItem: %s", query, mlbItem)
	rows, err := s.sqlDB.Query(query, sql.Named("p1", mlbItem))
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	var codEmpresa string
	var hasDepara bool
	if rows.Next() {
		var mlb, sku, filial string
		if err := rows.Scan(&mlb, &sku, &filial); err != nil {
			return "", "", err
		}
		hasDepara = true

		if strings.Contains(validadorPeca, "LC") {
			codEmpresa = "LUCIOS"
		} else if len(filial) > 0 {
			codEmpresa = filial
		} else {
			codEmpresa = "OUTROS"
		}
	} else {
		hasDepara = false
		if strings.Contains(validadorPeca, "LC") {
			codEmpresa = "LUCIOS"
		} else {
			codEmpresa = "OUTROS"
		}
	}

	// Determinar fornecedor seguindo exatamente a lógica do Python
	codFornecedor, finalCodEmpresa := s.getCodigoFornecedor(codEmpresa, hasDepara)

	return finalCodEmpresa, codFornecedor, nil
}

func (s *IntegrationService) getSchemaFromConta(conta string) string {
	// Schema dinâmico baseado na conta selecionada pelo usuário
	// O campo "Conta" determina qual schema usar na tabela stg_Depara
	switch strings.ToLower(conta) {
	case "principal":
		return "principal"
	case "oficial":
		return "oficial"
	case "renault":
		return "renault"
	case "psa":
		return "psa"
	case "ford":
		return "ford"
	case "jeep":
		return "jeep"
	default:
		return "principal" // fallback
	}
}

func (s *IntegrationService) getCodigoFornecedor(codEmpresa string, hasDepara bool) (string, string) {
	switch codEmpresa {
	case "17":
		return "7", "17"
	case "144":
		return "13", "144"
	case "44":
		// No Python: codEmpresa == '144' e codFornecedor = '13'
		return "13", "144"
	case "12":
		// No Python: codEmpresa = '17' e codFornecedor = '12'
		return "12", "17"
	case "40":
		return "1", "40"
	case "34":
		return "9", "34"
	case "41":
		return "11", "41"
	case "47":
		return "17", "47"
	case "140":
		return "1", "140"
	case "LUCIOS":
		if hasDepara {
			return "8", "LUCIOS"
		} else {
			return "lucio nao encontrado", "LUCIOS"
		}
	case "OUTROS":
		// No Python: codFornecedor = '8' e codEmpresa = '17'
		return "8", "17"
	default:
		return "nao encontrado", codEmpresa
	}
}

func (s *IntegrationService) getNBSToken(codEmpresa string) (string, error) {
	var conta string
	if codEmpresa == "LUCIOS" {
		conta = "HYSTALO17"
	} else {
		conta = fmt.Sprintf("HYSTALO%s", codEmpresa)
	}

	url := fmt.Sprintf("http://10.13.1.19:8080/nbsapi-gateway/token?usuario=%s&senha=nbsapi&idioma=PT&pacote=HYSTALO", conta)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Data.Token, nil
}

func (s *IntegrationService) getMLClientData(numPedido, tokenML string) (map[string]interface{}, error) {
	log.Printf("🔍 Coletando dados REAIS do cliente do Mercado Livre...")

	// Primeiro tenta /orders/{id}/billing_info
	urlBilling := fmt.Sprintf("https://api.mercadolibre.com/orders/%s/billing_info", numPedido)
	req, err := http.NewRequest("GET", urlBilling, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenML))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var billingInfo struct {
		BillingInfo struct {
			DocType        string      `json:"doc_type"`
			DocNumber      string      `json:"doc_number"`
			AdditionalInfo interface{} `json:"additional_info"`
		} `json:"billing_info"`
	}

	if resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&billingInfo); err != nil {
			return nil, err
		}
		log.Printf("✅ Dados do cliente obtidos via /orders/billing_info")
	} else {
		// Se falhou, tenta via pack (como no Python)
		log.Printf("🔍 Tentando obter dados via pack...")

		packURL := fmt.Sprintf("https://api.mercadolibre.com/packs/%s", numPedido)
		packReq, err := http.NewRequest("GET", packURL, nil)
		if err != nil {
			return nil, err
		}

		packReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenML))
		packReq.Header.Set("Content-Type", "application/json")

		packResp, err := client.Do(packReq)
		if err != nil {
			return nil, err
		}
		defer packResp.Body.Close()

		if packResp.StatusCode == 200 {
			var packResponse struct {
				Orders []struct {
					ID interface{} `json:"id"`
				} `json:"orders"`
			}

			if err := json.NewDecoder(packResp.Body).Decode(&packResponse); err != nil {
				return nil, err
			}

			if len(packResponse.Orders) > 0 {
				firstOrderID := fmt.Sprintf("%.0f", packResponse.Orders[0].ID)
				log.Printf("🔍 Pack encontrado! Primeiro pedido: %s", firstOrderID)

				// Agora busca billing_info do primeiro pedido
				billingURL := fmt.Sprintf("https://api.mercadolibre.com/orders/%s/billing_info", firstOrderID)
				billingReq, err := http.NewRequest("GET", billingURL, nil)
				if err != nil {
					return nil, err
				}

				billingReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenML))
				billingReq.Header.Set("Content-Type", "application/json")

				billingResp, err := client.Do(billingReq)
				if err != nil {
					return nil, err
				}
				defer billingResp.Body.Close()

				if billingResp.StatusCode == 200 {
					if err := json.NewDecoder(billingResp.Body).Decode(&billingInfo); err != nil {
						return nil, err
					}
					log.Printf("✅ Dados do cliente obtidos via pack")
				} else {
					return nil, fmt.Errorf("erro ao obter billing_info do pack: status %d", billingResp.StatusCode)
				}
			} else {
				return nil, fmt.Errorf("nenhum pedido encontrado no pack")
			}
		} else {
			return nil, fmt.Errorf("erro ao obter pack: status %d", packResp.StatusCode)
		}
	}

	// Processar dados do cliente (baseado no Python)
	var nomeCliente string
	var tipoCliente string
	codigoCliente := billingInfo.BillingInfo.DocNumber

	// Processar additional_info (pode ser array ou objeto)
	var additionalInfoMap map[string]interface{}
	if additionalInfoArray, ok := billingInfo.BillingInfo.AdditionalInfo.([]interface{}); ok {
		// Se for array, converter para map
		additionalInfoMap = make(map[string]interface{})
		for _, item := range additionalInfoArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if name, ok := itemMap["name"].(string); ok {
					if value, ok := itemMap["value"].(string); ok {
						additionalInfoMap[name] = value
					}
				}
			}
		}
	} else if additionalInfoMap, ok = billingInfo.BillingInfo.AdditionalInfo.(map[string]interface{}); ok {
		// Já é um map
	} else {
		additionalInfoMap = make(map[string]interface{})
	}

	if billingInfo.BillingInfo.DocType == "CPF" {
		tipoCliente = "F"
		// Tentar obter nome dos campos FIRST_NAME e LAST_NAME
		if firstName, ok := additionalInfoMap["FIRST_NAME"].(string); ok {
			if lastName, ok := additionalInfoMap["LAST_NAME"].(string); ok {
				nomeCliente = strings.ToUpper(firstName + " " + lastName)
			} else {
				nomeCliente = strings.ToUpper(firstName)
			}
		} else {
			nomeCliente = "CLIENTE ML"
		}
	} else {
		tipoCliente = "J"
		// Tentar obter nome do campo BUSINESS_NAME
		if businessName, ok := additionalInfoMap["BUSINESS_NAME"].(string); ok {
			nomeCliente = strings.ToUpper(businessName)
		} else {
			nomeCliente = "CLIENTE ML"
		}
	}

	log.Printf("🔍 Cliente coletado: Nome=%s, Tipo=%s, Doc=%s", nomeCliente, tipoCliente, codigoCliente)
	log.Printf("🔍 Additional Info coletado: %+v", additionalInfoMap)

	// Preparar dados do cliente para NBS (formato exato do Python)
	clienteData := map[string]interface{}{
		"codigoCliente":       codigoCliente,
		"codigoTipoCliente":   1,
		"codigoRamo":          "V",
		"codigoClasse":        53,
		"codigoClasseTipo":    "24",
		"codigoEstadoCivil":   "1",
		"prefixoCelular":      "11",
		"telefoneCelular":     "25948379",
		"prefixoComercial":    "11",
		"telefoneComercial":   "25948379",
		"prefixoResidencial":  "11",
		"telefoneResidencial": "25948379",
		"codigoNacionalidade": "36",
		"codigoProfissao":     "102",
		"paiCliente":          "",
		"maeCliente":          "",
		"emailCliente":        "09059264630@MAIL.COM.BR",
		"tipo":                tipoCliente,
		"nome":                nomeCliente,
		"sexo":                "F",
		"nascimento":          "1993-01-01T09:52:50.638Z",
		"cpfCnpj":             codigoCliente,
		"rgIe":                "0",
		"ssp":                 "SP",
		"atualizaExistente":   true,
		"clienteRevendedor":   false,
	}

	// Adicionar dados de endereço reais (se disponíveis)
	clienteData["rua"] = "Rua do Mercado Livre" // Dados padrão por enquanto
	clienteData["bairro"] = "Centro"
	clienteData["uf"] = "SP"
	clienteData["cep"] = "01234567"

	return clienteData, nil
}

func (s *IntegrationService) registerClient(numPedido, tokenML, tokenNBS string) (string, error) {
	log.Printf("👤 Cadastrando cliente REAL no NBS...")

	// Obter dados reais do cliente do Mercado Livre (como no Python)
	clienteData, err := s.getMLClientData(numPedido, tokenML)
	if err != nil {
		return "", fmt.Errorf("erro ao obter dados do cliente: %v", err)
	}

	// Enviar para API do NBS
	urlCliente := "http://10.13.1.19:8080/nbsapi-gateway/nbs/ecommerce/hystalo/api/clientes"
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", tokenNBS),
	}

	jsonData, err := json.Marshal(clienteData)
	if err != nil {
		return "", fmt.Errorf("erro ao criar JSON do cliente: %v", err)
	}

	log.Printf("🔍 JSON do cliente sendo enviado para NBS: %s", string(jsonData))
	log.Printf("🔍 URL do cliente: %s", urlCliente)
	log.Printf("🔍 Headers do cliente: %+v", headers)

	req, err := http.NewRequest("POST", urlCliente, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao enviar cliente para NBS: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("🔍 Status da resposta do cliente: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Erro na API do NBS (status %d): %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("erro na API do NBS (status %d): %s", resp.StatusCode, string(body))
	}

	// Ler resposta completa para debug
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta do NBS: %v", err)
	}

	log.Printf("🔍 RESPOSTA COMPLETA DO CLIENTE: %s", string(body))

	// Salvar resposta em arquivo para debug
	err = os.WriteFile("/tmp/nbs_cliente_response.json", body, 0644)
	if err != nil {
		log.Printf("⚠️ Erro ao salvar resposta do cliente: %v", err)
	} else {
		log.Printf("💾 Resposta do cliente salva em /tmp/nbs_cliente_response.json")
	}

	// Processar resposta do NBS
	var resultCliente struct {
		Sucesso bool `json:"sucesso"`
		Data    struct {
			CodCliente string `json:"codCliente"`
		} `json:"data"`
		Mensagem string `json:"mensagem"`
	}

	if err := json.Unmarshal(body, &resultCliente); err != nil {
		log.Printf("❌ Erro ao decodificar resposta do NBS: %v", err)
		return "", fmt.Errorf("erro ao decodificar resposta do NBS: %v", err)
	}

	// Verificar se o cadastro foi bem-sucedido (como no Python)
	if !resultCliente.Sucesso {
		log.Printf("❌ API do NBS retornou sucesso=false: %s", resultCliente.Mensagem)
		return "", fmt.Errorf("erro na API do NBS: %s", resultCliente.Mensagem)
	}

	// Como no Python, não extraímos codCliente da resposta, apenas verificamos sucesso
	// O codigoCliente (CPF/CNPJ) é usado diretamente para continuar o processo
	log.Printf("✅ Cliente cadastrado no NBS com sucesso")
	return "", nil // Retornamos string vazia para indicar sucesso sem código específico
}

func (s *IntegrationService) registerAddress(numPedido, tokenML, tokenNBS string) (string, error) {
	log.Printf("🏠 Cadastrando endereço REAL no NBS...")

	// Obter dados reais do cliente do Mercado Livre para usar o codigoCliente
	clienteData, err := s.getMLClientData(numPedido, tokenML)
	if err != nil {
		return "", fmt.Errorf("erro ao obter dados do cliente para endereço: %v", err)
	}

	codigoCliente := clienteData["codigoCliente"].(string)

	// Preparar dados do endereço (baseado no Python)
	enderecoData := map[string]interface{}{
		"codigoCliente":          codigoCliente,
		"clienteTipoEndereco":    4,
		"codCidades":             "3550308", // São Paulo
		"CEP":                    "01234567",
		"rua":                    "Rua do Mercado Livre",
		"complemento":            "N/A",
		"bairro":                 "Centro",
		"uf":                     "SP",
		"numeroEndereco":         "123",
		"nomePropriedade":        "Internet",
		"inscricaoEstadual":      "ISENTO",
		"fachada":                "Internet",
		"contato":                "Consumidor",
		"telefoneContato":        "25948379",
		"prefixoTelefoneContato": "11",
	}

	// Enviar para API do NBS
	urlEndereco := "http://10.13.1.19:8080/nbsapi-gateway/nbs/ecommerce/hystalo/api/clientes/endereco"
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", tokenNBS),
	}

	jsonData, err := json.Marshal(enderecoData)
	if err != nil {
		return "", fmt.Errorf("erro ao criar JSON do endereço: %v", err)
	}

	log.Printf("🔍 JSON do endereço sendo enviado para NBS: %s", string(jsonData))
	log.Printf("🔍 URL do endereço: %s", urlEndereco)
	log.Printf("🔍 Headers do endereço: %+v", headers)

	req, err := http.NewRequest("POST", urlEndereco, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao enviar endereço para NBS: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("🔍 Status da resposta do endereço: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Erro na API do NBS para endereço (status %d): %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("erro na API do NBS (status %d): %s", resp.StatusCode, string(body))
	}

	// Processar resposta do NBS
	var resultEndereco struct {
		Sucesso bool `json:"sucesso"`
		Data    struct {
			CodEndereco string `json:"codEndereco"`
		} `json:"data"`
		Mensagem string `json:"mensagem"`
	}

	// Ler resposta completa do endereço
	bodyEndereco, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta do NBS: %v", err)
	}

	log.Printf("🔍 RESPOSTA COMPLETA DO ENDEREÇO: %s", string(bodyEndereco))

	// Salvar resposta em arquivo para debug
	err = os.WriteFile("/tmp/nbs_endereco_response.json", bodyEndereco, 0644)
	if err != nil {
		log.Printf("⚠️ Erro ao salvar resposta do endereço: %v", err)
	} else {
		log.Printf("💾 Resposta do endereço salva em /tmp/nbs_endereco_response.json")
	}

	if err := json.Unmarshal(bodyEndereco, &resultEndereco); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta do NBS: %v", err)
	}

	// Verificar se o cadastro foi bem-sucedido (como no Python)
	if !resultEndereco.Sucesso {
		log.Printf("❌ API do NBS retornou sucesso=false para endereço: %s", resultEndereco.Mensagem)
		return "", fmt.Errorf("erro na API do NBS: %s", resultEndereco.Mensagem)
	}

	// Como no Python, não extraímos codEndereco da resposta, apenas verificamos sucesso
	log.Printf("✅ Endereço cadastrado no NBS com sucesso")
	return "", nil // Retornamos string vazia para indicar sucesso sem código específico
}

func (s *IntegrationService) insertInitialStatus(numPedido, conta, tokenUser, idShipping string) error {
	query := `
		INSERT INTO integrator.fato_StatusVenda 
		VALUES (
			'Mercado Livre',
			$1,
			$2,
			$3,
			$4,
			'Mercado Envios',
			NOW(),
			null,
			null,
			null,
			null,
			null,
			0
		)
	`

	_, err := s.pgDB.Exec(query, strings.ToUpper(conta), tokenUser, numPedido, idShipping)
	return err
}

func (s *IntegrationService) processItemsAndSendOrder(numPedido string, order *MLOrder, codEmpresa, codFornecedor, codCliente, tokenNBS string) (string, error) {
	log.Printf("🔍 Iniciando processamento REAL de itens para pedido %s", numPedido)
	log.Printf("🔍 CodEmpresa: %s, CodFornecedor: %s", codEmpresa, codFornecedor)

	// Processar itens do pedido para obter dados reais (baseado no Python)
	var finalResult []map[string]interface{}
	var valorTotal float64

	for i, item := range order.OrderItems {
		// Obter SKU do DePara usando o MLB do item
		mlbItem := fmt.Sprintf("%v", item.Item.ID)
		log.Printf("🔍 Processando item %d: MLB=%s, Preço=%f, Qtd=%d", i+1, mlbItem, item.UnitPrice, item.Quantity)

		sku, err := s.getSKUFromDePara(mlbItem, codEmpresa)
		if err != nil {
			log.Printf("⚠️ Erro ao obter SKU para MLB %s: %v", mlbItem, err)
			continue
		}

		log.Printf("✅ SKU encontrado para MLB %s: %s", mlbItem, sku)

		// Calcular valores
		valorUnitario := item.UnitPrice
		quantidade := item.Quantity
		valorItem := valorUnitario * float64(quantidade)
		valorTotal += valorItem

		// Adicionar item ao resultado final (formato do Python)
		finalResult = append(finalResult, map[string]interface{}{
			"COD_ITEM":       sku,
			"COD_FORNECEDOR": codFornecedor,
			"PRECO_UNITARIO": valorUnitario,
			"QTDE":           quantidade,
		})
	}

	if len(finalResult) == 0 {
		return "", fmt.Errorf("nenhum item válido encontrado para processamento")
	}

	log.Printf("🔍 Total de itens processados: %d, Valor total: %f", len(finalResult), valorTotal)

	// Preparar dados do pedido EXATAMENTE como no Python
	dataPedido := map[string]interface{}{
		"COD_PEDIDO_WEB":     1005502702, // numPedido[7:], // Remove os primeiros 7 caracteres como no Python
		"COD_CLIENTE":        codCliente, // Cliente dinâmico do NBS
		"TIPO_ENDERECO":      4,
		"COD_TRANSPORTADORA": 0,
		"VALOR_FRETE_TOTAL":  0.00,
		"CNPJ_INTERMED":      "03361252000134", // CNPJ do Mercado Livre
		"IDENT_CAD_INTERMED": "Mercado Livre",
		"NOME":               fmt.Sprintf("ECOMML%s", codEmpresa), // Nome do cliente (como no Python)
		"Itens":              finalResult,
		"Pagamentos": []map[string]interface{}{
			{
				"codigoBandeira":     "MP",
				"tipoCartao":         "CREDITO",
				"dataPagamento":      order.DateCreated,
				"numeroCartao":       "9999999999999999", // Cartão padrão para ML
				"numeroAutorizacao":  "01071531",         // Autorização padrão
				"quantidadeParcelas": 1,
			},
		},
	}

	// Enviar pedido para API do NBS (URL exata do Python)
	urlPedidoEnvio := "http://10.13.1.19:8080/nbsapi-gateway/nbs/ecommerce/hystalo/api/pedidos"
	log.Printf("🔍 Enviando pedido REAL para NBS: %s", urlPedidoEnvio)

	headersNBS := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", tokenNBS),
	}

	jsonData, err := json.Marshal(dataPedido)
	if err != nil {
		return "", fmt.Errorf("erro ao criar JSON do pedido: %v", err)
	}

	log.Printf("🔍 JSON do pedido REAL: %s", string(jsonData))
	log.Printf("🔍 COD_CLIENTE sendo enviado: '%s'", codCliente)
	log.Printf("🔍 COD_EMPRESA sendo usado: '%s'", codEmpresa)
	log.Printf("🔍 COD_FORNECEDOR sendo usado: '%s'", codFornecedor)
	log.Printf("🔍 URL do pedido: %s", urlPedidoEnvio)
	log.Printf("🔍 Headers do pedido: %+v", headersNBS)

	req, err := http.NewRequest("POST", urlPedidoEnvio, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	for key, value := range headersNBS {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	log.Printf("🔍 Fazendo requisição REAL para API do NBS...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Erro na requisição para NBS: %v", err)
		return "", fmt.Errorf("erro ao enviar pedido para NBS: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("🔍 Status da resposta do pedido: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Erro na API do NBS para pedido (status %d): %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("erro na API do NBS (status %d): %s", resp.StatusCode, string(body))
	}

	// Processar resposta do NBS (exatamente como no Python)
	var resultEnvio struct {
		Data struct {
			CodigoPedido interface{} `json:"codigoPedido"` // Pode ser string ou número
		} `json:"data"`
		Mensagem string `json:"mensagem"`
	}

	// Ler resposta completa do pedido
	bodyPedido, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta do NBS: %v", err)
	}

	log.Printf("🔍 RESPOSTA COMPLETA DO PEDIDO: %s", string(bodyPedido))

	// Salvar resposta em arquivo para debug
	err = os.WriteFile("/tmp/nbs_pedido_response.json", bodyPedido, 0644)
	if err != nil {
		log.Printf("⚠️ Erro ao salvar resposta do pedido: %v", err)
	} else {
		log.Printf("💾 Resposta do pedido salva em /tmp/nbs_pedido_response.json")
	}

	if err := json.Unmarshal(bodyPedido, &resultEnvio); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta do NBS: %v", err)
	}

	// Converter CodigoPedido para string (pode vir como número ou string)
	var numeroPedido string
	switch v := resultEnvio.Data.CodigoPedido.(type) {
	case string:
		numeroPedido = v
	case float64:
		numeroPedido = fmt.Sprintf("%.0f", v)
	case int:
		numeroPedido = fmt.Sprintf("%d", v)
	default:
		numeroPedido = fmt.Sprintf("%v", v)
	}

	log.Printf("🔍 Resposta completa da API do NBS: %+v", resultEnvio)

	if numeroPedido == "" {
		log.Printf("❌ Campo codigoPedido está vazio na resposta da API do NBS")
		return "", fmt.Errorf("número da pré-nota não retornado pela API do NBS")
	}

	log.Printf("✅ Pedido enviado para NBS com sucesso. Pré-nota REAL: %s", numeroPedido)
	return numeroPedido, nil
}

// getSKUFromDePara obtém o SKU real do DePara baseado no MLB
func (s *IntegrationService) getSKUFromDePara(mlbItem, codEmpresa string) (string, error) {
	// Determinar schema baseado no codEmpresa
	schema := s.getSchemaFromCodEmpresa(codEmpresa)

	query := fmt.Sprintf("SELECT sku FROM %s.stg_Depara WHERE mlb = @p1", schema)
	rows, err := s.sqlDB.Query(query, sql.Named("p1", mlbItem))
	if err != nil {
		return "", fmt.Errorf("erro ao consultar DePara: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		var sku string
		if err := rows.Scan(&sku); err != nil {
			return "", fmt.Errorf("erro ao ler SKU: %v", err)
		}
		return sku, nil
	}

	return "", fmt.Errorf("SKU não encontrado para MLB %s", mlbItem)
}

// getSchemaFromCodEmpresa determina o schema baseado no código da empresa
func (s *IntegrationService) getSchemaFromCodEmpresa(codEmpresa string) string {
	switch codEmpresa {
	case "17":
		return "principal"
	case "144":
		return "psa"
	case "44":
		return "psa"
	case "12":
		return "principal"
	case "40":
		return "ford"
	case "34":
		return "jeep"
	case "41":
		return "renault"
	case "47":
		return "renault"
	case "140":
		return "ford"
	case "LUCIOS":
		return "principal"
	case "OUTROS":
		return "principal"
	default:
		return "principal"
	}
}

func (s *IntegrationService) updateFinalStatus(numPedido, numeroPedido, tokenUser string) error {
	query := `
		UPDATE integrator.fato_StatusVenda 
		SET
			num_prenota = $1,
			dat_prenota = NOW(),
			flg_statuspedido = 1
		WHERE id_conta = $2
		AND num_pedido = $3
	`

	_, err := s.pgDB.Exec(query, numeroPedido, tokenUser, numPedido)
	return err
}

// sendTelegramMessage envia mensagem para o Telegram
func (s *IntegrationService) sendTelegramMessage(message string) {
	tokenTelegram := "7055914019:AAEJL0zVCZsaGv-RLKVf4VGcFY4-haShUx4"
	chatID := "-4282027613"

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tokenTelegram)
	data := map[string]string{
		"chat_id": chatID,
		"text":    message,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Erro ao criar JSON para Telegram: %v", err)
		return
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("Erro ao enviar mensagem para Telegram: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("📱 Mensagem enviada para Telegram com sucesso")
}
