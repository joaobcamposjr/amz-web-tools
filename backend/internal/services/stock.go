package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"amz-web-tools/backend/internal/config"
	"amz-web-tools/backend/internal/models"

	_ "github.com/sijms/go-ora/v2"
)

type StockService struct {
	oracleDB *sql.DB
	config   *config.Config
}

func NewStockService(cfg *config.Config) (*StockService, error) {
	log.Printf("🔧 Starting StockService initialization...")

	// Check if Oracle configuration is available
	log.Printf("🔧 Oracle config check: Host=%s, User=%s, Password=%s, Service=%s",
		cfg.OracleHost, cfg.OracleUser, cfg.OraclePassword, cfg.OracleService)

	if cfg.OracleHost == "" || cfg.OracleUser == "" || cfg.OraclePassword == "" || cfg.OracleService == "" {
		log.Printf("⚠️ Oracle configuration not available, using mock data")
		return &StockService{
			oracleDB: nil,
			config:   cfg,
		}, nil
	}

	log.Printf("🔧 Oracle configuration is available, attempting connection...")

	// Oracle connection string
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
		cfg.OracleUser,
		cfg.OraclePassword,
		cfg.OracleHost,
		cfg.OraclePort,
		cfg.OracleService,
	)

	log.Printf("🔧 Oracle DSN: %s", dsn)

	log.Printf("🔗 Connecting to Oracle: %s:%s/%s", cfg.OracleHost, cfg.OraclePort, cfg.OracleService)

	oracleDB, err := sql.Open("oracle", dsn)
	if err != nil {
		log.Printf("❌ Failed to open Oracle connection: %v", err)
		return &StockService{
			oracleDB: nil,
			config:   cfg,
		}, nil
	}

	log.Printf("🔧 Oracle connection opened, testing ping...")

	// Test connection with aggressive timeout using goroutine to ensure it doesn't hang
	pingDone := make(chan error, 1)
	var pingErr error

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pingErr = oracleDB.PingContext(ctx)
		pingDone <- pingErr
	}()

	// Wait for ping with timeout
	select {
	case err := <-pingDone:
		if err != nil {
			log.Printf("❌ Failed to ping Oracle database (timeout or error): %v", err)
			oracleDB.Close()
			return &StockService{
				oracleDB: nil,
				config:   cfg,
			}, nil
		}
	case <-time.After(6 * time.Second):
		log.Printf("❌ Oracle ping timeout after 6 seconds, closing connection")
		oracleDB.Close()
		return &StockService{
			oracleDB: nil,
			config:   cfg,
		}, nil
	}

	log.Println("✅ Oracle database connected successfully")

	service := &StockService{
		oracleDB: oracleDB,
		config:   cfg,
	}

	log.Printf("🔧 StockService created with Oracle connection: %v", service.oracleDB != nil)
	return service, nil
}

// SearchStock searches for stock by SKU
func (s *StockService) SearchStock(sku string) ([]models.StockItem, error) {
	// Clean and format SKU (remove 'LC' prefix and convert to uppercase)
	cleanSKU := strings.ToUpper(strings.Replace(sku, "LC", "", -1))

	log.Printf("🔍 Searching stock for SKU: %s (cleaned: %s)", sku, cleanSKU)

	// Check if Oracle connection is available
	if s.oracleDB == nil {
		log.Printf("⚠️ Oracle connection not available (oracleDB is nil), returning empty results")
		log.Printf("🔧 Debug: Config values - Host: %s, User: %s, Service: %s", s.config.OracleHost, s.config.OracleUser, s.config.OracleService)
		return []models.StockItem{}, nil
	}

	query := `
		SELECT DISTINCT
			e.cod_empresa,
			em.nome nom_empresa,
			e.cod_fornecedor, 
			fe.NOME_FORNECEDOR,
			e.cod_item as cod_item,
			e.valor_reposicao,
			e.CUSTO_CONTABIL,
			e.VALOR_VENDA,
			e.ESTOQUE,
			e.RESERVADO,
			e.ESTOQUE - e.RESERVADO as ESTOQUE_DISPONIVEL
		FROM nbs.CRANI_PECAS_ITENS e 
		LEFT JOIN nbs.FORNECEDOR_ESTOQUE fe 
		ON e.cod_fornecedor = fe.cod_fornecedor
		LEFT JOIN nbs.EMPRESAS em
		ON e.COD_EMPRESA = em.COD_EMPRESA 
		WHERE
			e.cod_empresa IN (17,34,140,144,147)
			AND e.cod_fornecedor IN (7,9,1,13,16,17)
			AND e.cod_item = :1`

	rows, err := s.oracleDB.Query(query, cleanSKU)
	if err != nil {
		log.Printf("❌ Error querying Oracle: %v", err)
		return nil, fmt.Errorf("failed to query stock: %w", err)
	}
	defer rows.Close()

	var items []models.StockItem
	for rows.Next() {
		var item models.StockItem
		err := rows.Scan(
			&item.CodEmpresa,
			&item.NomeEmpresa,
			&item.CodFornecedor,
			&item.NomeFornecedor,
			&item.CodItem,
			&item.ValorReposicao,
			&item.CustoContabil,
			&item.ValorVenda,
			&item.Estoque,
			&item.Reservado,
			&item.EstoqueDisponivel,
		)
		if err != nil {
			log.Printf("❌ Error scanning row: %v", err)
			return nil, fmt.Errorf("failed to scan stock row: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Printf("❌ Error iterating rows: %v", err)
		return nil, fmt.Errorf("error iterating stock rows: %w", err)
	}

	log.Printf("✅ Found %d stock items for SKU: %s", len(items), cleanSKU)
	return items, nil
}

// GetOracleDB returns the Oracle database connection
func (s *StockService) GetOracleDB() *sql.DB {
	return s.oracleDB
}

// Close closes the Oracle connection
func (s *StockService) Close() error {
	if s.oracleDB != nil {
		return s.oracleDB.Close()
	}
	return nil
}
