package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"          // PostgreSQL driver
	_ "github.com/sijms/go-ora/v2" // Oracle driver

	"amz-web-tools/backend/internal/config"
	"amz-web-tools/backend/internal/models"
)

type CarPlateService struct {
	db     *sql.DB
	Config *config.Config
}

// PlateAPIResponse represents the raw JSON response from the API
type PlateAPIResponse map[string]interface{}

func NewCarPlateService(db *sql.DB, cfg *config.Config) *CarPlateService {
	return &CarPlateService{
		db:     db,
		Config: cfg,
	}
}

// GetDB returns the database connection
func (s *CarPlateService) GetDB() *sql.DB {
	return s.db
}

// PlateResult represents the result of a plate query with source information
type PlateResult struct {
	Data   *PlateAPIResponse `json:"data"`
	Source string            `json:"source"` // "cache" or "api"
}

// GetCarPlate retrieves car plate information with caching
func (s *CarPlateService) GetCarPlate(plate string, userID string) (*PlateResult, error) {
	// Normalize plate (remove spaces, convert to uppercase)
	plate = strings.ToUpper(strings.ReplaceAll(plate, " ", ""))

	log.Printf("Searching for plate: %s", plate)

	// First, check cache
	cachedData, err := s.getFromCache(plate)
	if err == nil && cachedData != nil {
		log.Printf("🚀 Plate %s found in cache - returning cached data", plate)
		// Save to history even if from cache
		s.saveToHistory(plate, cachedData, "success", "", userID)
		return &PlateResult{Data: cachedData, Source: "cache"}, nil
	}

	if err != nil {
		log.Printf("⚠️ Error checking cache for plate %s: %v", plate, err)
	} else {
		log.Printf("🔍 Plate %s not found in cache, fetching from API", plate)
	}

	// If not in cache, fetch from API
	apiData, err := s.fetchFromAPI(plate)
	if err != nil {
		log.Printf("❌ Failed to fetch plate %s from API: %v", plate, err)
		// Save error to history
		s.saveToHistory(plate, nil, "error", err.Error(), userID)
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}

	log.Printf("✅ Successfully fetched plate %s from API", plate)

	// Cache the result
	if err := s.saveToCache(plate, apiData); err != nil {
		log.Printf("⚠️ Warning: Failed to cache plate %s: %v", plate, err)
		// Don't return error, just log warning
	} else {
		log.Printf("✅ Plate %s cached successfully", plate)
	}

	// Save to history
	s.saveToHistory(plate, apiData, "success", "", userID)

	return &PlateResult{Data: apiData, Source: "api"}, nil
}

// GetCacheStats returns cache statistics
func (s *CarPlateService) GetCacheStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total cache entries
	var totalCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM plate_cache").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total_entries"] = totalCount

	// Valid cache entries
	var validCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM plate_cache WHERE expires_at > GETDATE()").Scan(&validCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid count: %w", err)
	}
	stats["valid_entries"] = validCount

	// Expired cache entries
	var expiredCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM plate_cache WHERE expires_at <= GETDATE()").Scan(&expiredCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired count: %w", err)
	}
	stats["expired_entries"] = expiredCount

	// Cache hit rate (approximate)
	if totalCount > 0 {
		stats["cache_hit_rate"] = float64(validCount) / float64(totalCount) * 100
	} else {
		stats["cache_hit_rate"] = 0.0
	}

	return stats, nil
}

// getFromCache retrieves plate data from database cache
func (s *CarPlateService) getFromCache(plate string) (*PlateAPIResponse, error) {
	query := `
		SELECT data FROM plate_cache 
		WHERE plate = @p1 AND expires_at > GETDATE()`

	var dataJSON string
	err := s.db.QueryRow(query, sql.Named("p1", plate)).Scan(&dataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found in cache
		}
		return nil, err
	}

	var plateData PlateAPIResponse
	if err := json.Unmarshal([]byte(dataJSON), &plateData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return &plateData, nil
}

// fetchFromAPI fetches plate data from external API
func (s *CarPlateService) fetchFromAPI(plate string) (*PlateAPIResponse, error) {
	// Replace PLACA placeholder in URL
	apiURL := strings.Replace(s.Config.PlateAPIURL, "PLACA", plate, 1)

	log.Printf("🌐 Fetching from API: %s", apiURL)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make HTTP request
	resp, err := client.Get(apiURL)
	if err != nil {
		log.Printf("❌ HTTP request failed: %v", err)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("📡 API Response Status: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		// Try to read error body
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ API Error Response: %s", string(bodyBytes))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var plateData PlateAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&plateData); err != nil {
		log.Printf("❌ Failed to decode JSON: %v", err)
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	log.Printf("✅ Successfully decoded API response")

	// Validate response - check if there's an error message
	if mensagem, exists := plateData["mensagemRetorno"]; exists {
		if mensagemStr, ok := mensagem.(string); ok && mensagemStr != "Sem erros." {
			log.Printf("❌ API returned error: %s", mensagemStr)
			return nil, fmt.Errorf("API error: %s", mensagemStr)
		}
	}

	log.Printf("✅ API response validated successfully")
	return &plateData, nil
}

// saveToCache saves plate data to database cache
func (s *CarPlateService) saveToCache(plate string, data *PlateAPIResponse) error {
	// Convert to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Cache for 7 days (168 hours) to reduce API calls
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	query := `
		MERGE plate_cache AS target
		USING (SELECT @p1 AS plate, @p2 AS data, @p3 AS expires_at) AS source
		ON target.plate = source.plate
		WHEN MATCHED THEN
			UPDATE SET data = source.data, expires_at = source.expires_at, created_at = GETDATE()
		WHEN NOT MATCHED THEN
			INSERT (plate, data, expires_at) VALUES (source.plate, source.data, source.expires_at);`

	_, err = s.db.Exec(query, plate, string(dataJSON), expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save to cache: %w", err)
	}

	log.Printf("✅ Plate %s cached until %s", plate, expiresAt.Format("2006-01-02 15:04:05"))
	return nil
}

// saveToHistory saves consultation to history table
func (s *CarPlateService) saveToHistory(plate string, data *PlateAPIResponse, status, errorMessage, userID string) {
	var responseData string
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err == nil {
			responseData = string(jsonData)
		}
	}

	query := `
		INSERT INTO car_plate_history (plate, response_data, status, error_message, user_id)
		VALUES (@p1, @p2, @p3, @p4, @p5)`

	_, err := s.db.Exec(query, plate, responseData, status, errorMessage, userID)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to save plate %s to history: %v", plate, err)
	} else {
		log.Printf("✅ Plate %s saved to history", plate)
	}
}

// GetPlateHistory retrieves search history for a user
func (s *CarPlateService) GetPlateHistory(userID string, limit int) ([]models.CarPlateHistory, error) {
	query := `
		SELECT TOP (@p1) id, plate, response_data, status, error_message, created_at, user_id
		FROM car_plate_history 
		WHERE user_id = @p2 OR user_id IS NULL
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query, sql.Named("p1", limit), sql.Named("p2", userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.CarPlateHistory
	for rows.Next() {
		var item models.CarPlateHistory
		err := rows.Scan(&item.ID, &item.Plate, &item.ResponseData, &item.Status, &item.ErrorMessage, &item.CreatedAt, &item.UserID)
		if err != nil {
			return nil, err
		}
		history = append(history, item)
	}

	return history, nil
}

// CleanExpiredCache removes expired cache entries
func (s *CarPlateService) CleanExpiredCache() error {
	query := `DELETE FROM plate_cache WHERE expires_at < GETDATE()`

	result, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Cleaned %d expired cache entries", rowsAffected)

	return nil
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	CarPlateQueries int `json:"car_plate_queries"`
	DeParaProducts  int `json:"depara_products"`
	XMLImports      int `json:"xml_imports"`
	StockItems      int `json:"stock_items"`
}

// GetDashboardStats retrieves dashboard statistics
func (s *CarPlateService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 1. Consultas de Placa - contagem da tabela plate_cache no SQL Server
	query1 := `SELECT COUNT(*) FROM portal.dbo.plate_cache`
	err := s.db.QueryRow(query1).Scan(&stats.CarPlateQueries)
	if err != nil {
		log.Printf("Error getting car plate queries count: %v", err)
		stats.CarPlateQueries = 0
	}

	// 2. Produtos DePara - contagem de todas as tabelas _base no schema integration
	// Lista das tabelas DePara para somar
	deparaTables := []string{
		"integration.amazonas_psa.mercadolivre_base",
		"integration.amazonas_renault.mercadolivre_base",
		"integration.amazonas_principal.mercadolivre_base",
		"integration.amazonas_oficial.mercadolivre_base",
		"integration.amazonas_jeep.mercadolivre_base",
		"integration.amazonas_ford.mercadolivre_base",
	}

	totalDeParaProducts := 0
	for _, table := range deparaTables {
		query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table)
		var count int
		err := s.db.QueryRow(query).Scan(&count)
		if err != nil {
			log.Printf("⚠️ Error getting count from %s: %v", table, err)
			// Continue with other tables even if one fails
			continue
		}
		totalDeParaProducts += count
		log.Printf("✅ Count from %s: %d", table, count)
	}

	stats.DeParaProducts = totalDeParaProducts
	log.Printf("✅ Total DePara products across all tables: %d", totalDeParaProducts)

	// 3. Importações XML - contagem da tabela codako_bi.integrator.fato_statusvenda no PostgreSQL
	// Usando a consulta correta fornecida pelo usuário
	log.Printf("🔍 Calling GetXMLImportsCount...")
	xmlCount := s.GetXMLImportsCount()
	log.Printf("🔍 XML imports result: %d", xmlCount)

	// Se PostgreSQL falhar, usar fallback do SQL Server
	if xmlCount == 0 {
		log.Printf("🔄 PostgreSQL failed, using SQL Server fallback for XML imports")
		query3 := `SELECT COUNT(*) FROM xml_integrator_logs`
		err = s.db.QueryRow(query3).Scan(&stats.XMLImports)
		if err != nil {
			log.Printf("Error getting XML imports fallback count: %v", err)
			stats.XMLImports = 0
		} else {
			log.Printf("✅ XML imports fallback count: %d", stats.XMLImports)
		}
	} else {
		stats.XMLImports = xmlCount
	}

	// 4. Itens em Estoque - contagem da tabela nbs.CRANI_PECAS_ITENS no Oracle
	// Usando a consulta correta fornecida pelo usuário
	log.Printf("🔍 Calling GetStockItemsCount...")
	stats.StockItems = s.GetStockItemsCount()
	log.Printf("🔍 Stock items result: %d", stats.StockItems)

	return stats, nil
}

// GetXMLImportsCount retrieves XML imports count from PostgreSQL
func (s *CarPlateService) GetXMLImportsCount() int {
	log.Printf("🔍 Starting XML imports count query...")

	// Verificar se as configurações PostgreSQL estão disponíveis
	if s.Config.PGHost == "" || s.Config.PGUser == "" || s.Config.PGPassword == "" || s.Config.PGDatabase == "" {
		log.Printf("⚠️ PostgreSQL configuration not available: Host=%s, User=%s, Database=%s", s.Config.PGHost, s.Config.PGUser, s.Config.PGDatabase)
		return 0 // Usar fallback para SQL Server
	}

	// Conectar ao PostgreSQL
	pgDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		s.Config.PGHost, s.Config.PGPort, s.Config.PGUser, s.Config.PGPassword, s.Config.PGDatabase, s.Config.PGSSLMode)

	log.Printf("🔗 Connecting to PostgreSQL: %s:%s/%s", s.Config.PGHost, s.Config.PGPort, s.Config.PGDatabase)

	pgDB, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Printf("❌ Error connecting to PostgreSQL: %v", err)
		return 0 // Usar fallback para SQL Server
	}
	defer pgDB.Close()

	// Testar conexão com timeout
	if err := pgDB.Ping(); err != nil {
		log.Printf("❌ Error pinging PostgreSQL: %v", err)
		return 0 // Usar fallback para SQL Server
	}

	log.Printf("✅ PostgreSQL connected successfully")

	// Executar consulta correta
	query := `SELECT COUNT(*) FROM codako_bi.integrator.fato_statusvenda`
	log.Printf("🔍 Executing query: %s", query)

	var count int
	err = pgDB.QueryRow(query).Scan(&count)
	if err != nil {
		log.Printf("❌ Error getting XML imports count from PostgreSQL: %v", err)
		return 0 // Usar fallback para SQL Server
	}

	log.Printf("✅ XML imports count from PostgreSQL: %d", count)
	return count
}

// GetStockItemsCount retrieves stock items count from Oracle
func (s *CarPlateService) GetStockItemsCount() int {
	log.Printf("🔍 Starting stock items count query...")

	// Verificar se as configurações Oracle estão disponíveis
	if s.Config.OracleHost == "" || s.Config.OracleUser == "" || s.Config.OraclePassword == "" || s.Config.OracleService == "" {
		log.Printf("⚠️ Oracle configuration not available: Host=%s, User=%s, Service=%s", s.Config.OracleHost, s.Config.OracleUser, s.Config.OracleService)
		return 888 // Valor de fallback diferente para identificar
	}

	// Conectar ao Oracle
	oracleDSN := fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
		s.Config.OracleUser, s.Config.OraclePassword, s.Config.OracleHost, s.Config.OraclePort, s.Config.OracleService)

	log.Printf("🔗 Connecting to Oracle: %s:%s/%s", s.Config.OracleHost, s.Config.OraclePort, s.Config.OracleService)

	oracleDB, err := sql.Open("oracle", oracleDSN)
	if err != nil {
		log.Printf("❌ Error connecting to Oracle: %v", err)
		return 888
	}
	defer oracleDB.Close()

	// Testar conexão
	if err := oracleDB.Ping(); err != nil {
		log.Printf("❌ Error pinging Oracle: %v", err)
		return 888
	}

	log.Printf("✅ Oracle connected successfully")

	// Executar consulta correta
	query := `SELECT COUNT(DISTINCT cod_item) FROM nbs.CRANI_PECAS_ITENS WHERE cod_empresa IN (1,3,17,31,34,35,40,41,43,144,45,47,48,140)`
	log.Printf("🔍 Executing query: %s", query)

	var count int
	err = oracleDB.QueryRow(query).Scan(&count)
	if err != nil {
		log.Printf("❌ Error getting stock items count from Oracle: %v", err)
		return 888
	}

	log.Printf("✅ Stock items count from Oracle: %d", count)
	return count
}
