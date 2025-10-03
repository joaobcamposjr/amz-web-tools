package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"amz-web-tools/backend/internal/models"
)

type DeParaService struct {
	db           *sql.DB
	auditService *AuditService
}

func NewDeParaService(db *sql.DB, auditService *AuditService) *DeParaService {
	return &DeParaService{
		db:           db,
		auditService: auditService,
	}
}

// GetAvailableTables returns list of available integration tables
func (s *DeParaService) GetAvailableTables() ([]models.IntegrationTable, error) {
	log.Printf("🔍 Starting GetAvailableTables...")

	// Since INFORMATION_SCHEMA doesn't show these tables properly, we'll use the verified list
	// and check if each table exists by trying to query it
	verifiedTables := []string{
		"integration.amazonas_psa.mercadolivre_base",
		"integration.amazonas_renault.mercadolivre_base",
		"integration.amazonas_principal.mercadolivre_base",
		"integration.amazonas_oficial.mercadolivre_base",
		"integration.amazonas_jeep.mercadolivre_base",
		"integration.amazonas_ford.mercadolivre_base",
	}

	var tables []models.IntegrationTable
	id := 1

	for _, tableName := range verifiedTables {
		log.Printf("🔍 Checking table: %s", tableName)

		// Check if table exists by trying to query it
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		var count int
		err := s.db.QueryRow(query).Scan(&count)
		if err != nil {
			log.Printf("⚠️ Table %s not accessible: %v", tableName, err)
			continue
		}

		log.Printf("✅ Found table %s with %d rows", tableName, count)

		// Create display name from table name
		parts := strings.Split(tableName, ".")
		var displayName string
		if len(parts) >= 2 {
			tablePart := parts[1] // amazonas_psa.mercadolivre_base

			// Remove "_base" and split by "_"
			tablePart = strings.ReplaceAll(tablePart, "_base", "")
			tableParts := strings.Split(tablePart, "_")

			if len(tableParts) >= 2 {
				conta := strings.Title(tableParts[1]) // PSA, Renault, Principal, etc.
				displayName = fmt.Sprintf("Amazonas %s MercadoLivre", conta)
			} else {
				displayName = tablePart
			}

			log.Printf("📝 Adding table %d: %s (%s)", id, tableName, displayName)

			tables = append(tables, models.IntegrationTable{
				ID:          fmt.Sprintf("%d", id),
				TableName:   tableName,
				DisplayName: displayName,
				IsActive:    true,
				CreatedAt:   time.Now(),
			})
			id++
		} else {
			log.Printf("⚠️ Invalid table name format: %s", tableName)
		}
	}

	log.Printf("✅ Returning %d verified integration tables", len(tables))
	return tables, nil
}

// GetTableOptions returns available options for dynamic table selection
func (s *DeParaService) GetTableOptions() (map[string][]string, error) {
	// Based on the 6 verified tables, return the available options
	options := map[string][]string{
		"empresa":     {"amazonas"},
		"conta":       {"psa", "renault", "principal", "oficial", "jeep", "ford"},
		"marketplace": {"mercadolivre"},
	}

	log.Printf("✅ Generated table options: %+v", options)
	return options, nil
}

// hasColumn checks if a column exists in the specified table
func (s *DeParaService) hasColumn(tableName, columnName string) (bool, error) {
	// Extract schema and table name from full table name (e.g., "integration.amazonas_jeep.mercadolivre_base")
	parts := strings.Split(tableName, ".")
	if len(parts) < 3 {
		return false, fmt.Errorf("invalid table name format: %s", tableName)
	}

	// Get the schema and table parts
	schemaName := parts[0]      // integration
	tableSchema := parts[1]     // amazonas_jeep
	actualTableName := parts[2] // mercadolivre_base

	log.Printf("🔍 Checking column %s in table %s (schema: %s, table_schema: %s, table_name: %s)",
		columnName, tableName, schemaName, tableSchema, actualTableName)

	// Use the working query format that we confirmed works with sqlcmd
	// Query 1: schema=amazonas_renault, table=mercadolivre_base (this one worked!)
	query := `SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @p1 AND TABLE_NAME = @p2 AND COLUMN_NAME = @p3`
	args := []interface{}{tableSchema, actualTableName, columnName}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		log.Printf("⚠️ Query failed for column %s: %v", columnName, err)
		return false, err
	}

	if count > 0 {
		log.Printf("✅ Column %s found (count=%d)", columnName, count)
		return true, nil
	}

	log.Printf("❌ Column %s not found (count=%d)", columnName, count)
	return false, nil
}

// SearchProducts searches products based on criteria with pagination
func (s *DeParaService) SearchProducts(tableName, query, searchBy string, page, pageSize int) ([]models.DeParaProduct, int, error) {
	var whereClause string
	var args []interface{}

	switch searchBy {
	case "id":
		whereClause = "WHERE id = @p1"
		args = append(args, query)
	case "mlbu":
		whereClause = "WHERE mlbu = @p1"
		args = append(args, query)
	case "sku":
		whereClause = "WHERE sku LIKE @p1"
		args = append(args, "%"+query+"%")
	default:
		// Auto-detect search type
		if strings.HasPrefix(query, "MLBU") {
			whereClause = "WHERE mlbu = @p1"
			args = append(args, query)
			log.Printf("🔍 Auto-detected MLBU search for: %s", query)
		} else if strings.HasPrefix(query, "MLB") && len(query) >= 10 {
			whereClause = "WHERE id = @p1"
			args = append(args, query)
			log.Printf("🔍 Auto-detected ID search for: %s", query)
		} else {
			// Default to SKU search for anything else
			whereClause = "WHERE sku LIKE @p1"
			args = append(args, "%"+query+"%")
			log.Printf("🔍 Auto-detected SKU search for: %s", query)
		}
	}

	// Build table name dynamically: {empresa}_{conta}.{marketplace}_base
	// For now, default to amazonas_psa.mercadolivre_base
	actualTableName := s.buildTableName(tableName)

	// Count total results
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, actualTableName, whereClause)
	var totalCount int
	err := s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Check if permalink column exists in this table
	hasPermalink, err := s.hasColumn(actualTableName, "permalink")
	if err != nil {
		log.Printf("⚠️ Warning: Could not check for permalink column: %v", err)
		hasPermalink = false
	}

	// Check for new columns
	hasPrice, err := s.hasColumn(actualTableName, "price")
	if err != nil {
		log.Printf("⚠️ Warning: Could not check for price column: %v", err)
		hasPrice = false
	}
	hasSoldQuantity, err := s.hasColumn(actualTableName, "sold_quantity")
	if err != nil {
		log.Printf("⚠️ Warning: Could not check for sold_quantity column: %v", err)
		hasSoldQuantity = false
	}
	hasTags, err := s.hasColumn(actualTableName, "tags")
	if err != nil {
		log.Printf("⚠️ Warning: Could not check for tags column: %v", err)
		hasTags = false
	}
	hasStatus, err := s.hasColumn(actualTableName, "status")
	if err != nil {
		log.Printf("⚠️ Warning: Could not check for status column: %v", err)
		hasStatus = false
	}
	hasHealth, err := s.hasColumn(actualTableName, "health")
	if err != nil {
		log.Printf("⚠️ Warning: Could not check for health column: %v", err)
		hasHealth = false
	}
	hasCreatedAtSite, err := s.hasColumn(actualTableName, "created_at_site")
	if err != nil {
		log.Printf("⚠️ Warning: Could not check for created_at_site column: %v", err)
		hasCreatedAtSite = false
	}
	hasUpdatedAtSite, err := s.hasColumn(actualTableName, "updated_at_site")
	if err != nil {
		log.Printf("⚠️ Warning: Could not check for updated_at_site column: %v", err)
		hasUpdatedAtSite = false
	}

	// Build query based on available columns with NULL handling
	var selectColumns string
	baseColumns := "id, COALESCE(mlbu, '') as mlbu, COALESCE(type, '') as type, COALESCE(sku, '') as sku, COALESCE(company, '') as company"

	// Add permalink
	if hasPermalink {
		selectColumns = baseColumns + ", COALESCE(permalink, '') as permalink"
	} else {
		selectColumns = baseColumns + ", '' as permalink"
	}

	// Add new columns if they exist
	if hasPrice {
		selectColumns += ", price"
	} else {
		selectColumns += ", NULL as price"
	}

	if hasSoldQuantity {
		selectColumns += ", sold_quantity"
	} else {
		selectColumns += ", NULL as sold_quantity"
	}

	if hasTags {
		selectColumns += ", COALESCE(tags, '') as tags"
	} else {
		selectColumns += ", '' as tags"
	}

	if hasStatus {
		selectColumns += ", COALESCE(status, '') as status"
	} else {
		selectColumns += ", '' as status"
	}

	if hasHealth {
		selectColumns += ", health"
	} else {
		selectColumns += ", NULL as health"
	}

	// Add shipping costs
	selectColumns += ", ship_cost_slow, ship_cost_standard, ship_cost_nextday, COALESCE(pictures, '') as pictures"

	// Add site timestamps if they exist
	if hasCreatedAtSite {
		selectColumns += ", created_at_site"
	} else {
		selectColumns += ", NULL as created_at_site"
	}

	if hasUpdatedAtSite {
		selectColumns += ", updated_at_site"
	} else {
		selectColumns += ", NULL as updated_at_site"
	}

	// Add system timestamps
	selectColumns += ", updated_at, created_at"

	// Debug log
	log.Printf("🔍 Column detection for %s: price=%v, sold_quantity=%v, tags=%v, status=%v, health=%v, created_at_site=%v, updated_at_site=%v",
		actualTableName, hasPrice, hasSoldQuantity, hasTags, hasStatus, hasHealth, hasCreatedAtSite, hasUpdatedAtSite)

	// First get all results for total count
	allResultsQuery := fmt.Sprintf(`
		SELECT %s
		FROM %s 
		%s
		ORDER BY updated_at DESC`, selectColumns, actualTableName, whereClause)

	querySQL := allResultsQuery

	log.Printf("🔍 Searching in %s with query: %s", actualTableName, querySQL)

	rows, err := s.db.Query(querySQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	var allProducts []models.DeParaProduct
	for rows.Next() {
		var product models.DeParaProduct
		var picturesJSON string
		var createdAtSite, updatedAtSite sql.NullTime

		err := rows.Scan(
			&product.ID, &product.MLBU, &product.Type, &product.SKU, &product.Company,
			&product.Permalink, &product.Price, &product.SoldQuantity, &product.Tags,
			&product.Status, &product.Health, &product.ShipCostSlow, &product.ShipCostStandard,
			&product.ShipCostNextday, &picturesJSON, &createdAtSite, &updatedAtSite,
			&product.UpdatedAt, &product.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Parse pictures JSON
		if picturesJSON != "" {
			// Handle Python list format with single quotes
			cleanJSON := strings.ReplaceAll(picturesJSON, "'", "\"")

			// Try to parse as JSON array
			if err := json.Unmarshal([]byte(cleanJSON), &product.Pictures); err != nil {
				log.Printf("⚠️ Warning: Failed to parse pictures for product %s: %v", product.ID, err)
				log.Printf("⚠️ Raw pictures data: %s", picturesJSON)
				product.Pictures = []string{}
			} else {
				log.Printf("✅ Successfully parsed %d pictures for product %s", len(product.Pictures), product.ID)
			}
		} else {
			product.Pictures = []string{}
		}

		// Set site timestamps
		if createdAtSite.Valid {
			product.CreatedAtSite = createdAtSite
		}
		if updatedAtSite.Valid {
			product.UpdatedAtSite = updatedAtSite
		}

		// Debug log for first product
		if len(allProducts) == 0 {
			log.Printf("🔍 First product data: ID=%s, Tags=%v, Price=%v, SoldQuantity=%v, Status=%v, Health=%v",
				product.ID, product.Tags, product.Price, product.SoldQuantity, product.Status, product.Health)
		}

		allProducts = append(allProducts, product)
	}

	// Apply pagination in memory
	totalCount = len(allProducts)
	start := offset
	end := start + pageSize

	// Ensure we don't go out of bounds
	if start >= totalCount {
		start = totalCount
		end = totalCount
	} else if end > totalCount {
		end = totalCount
	}

	var products []models.DeParaProduct
	if len(allProducts) > 0 {
		if start < totalCount && end > start {
			products = allProducts[start:end]
		} else {
			// Fallback: return first pageSize products
			if pageSize > len(allProducts) {
				products = allProducts
			} else {
				products = allProducts[:pageSize]
			}
		}
	}

	log.Printf("✅ Found %d products (page %d, total: %d)", len(products), page, totalCount)
	return products, totalCount, nil
}

// GetProductByID gets a single product by ID
func (s *DeParaService) GetProductByID(tableName, id string) (*models.DeParaProduct, error) {
	actualTableName := s.buildTableName(tableName)
	query := fmt.Sprintf(`
		SELECT id, mlbu, type, sku, company, permalink, price, sold_quantity, tags, status, health,
		       ship_cost_slow, ship_cost_standard, ship_cost_nextday, 
		       pictures, created_at_site, updated_at_site, updated_at, created_at
		FROM %s 
		WHERE id = @p1`, actualTableName)

	var product models.DeParaProduct
	var picturesJSON string
	var createdAtSite, updatedAtSite sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&product.ID, &product.MLBU, &product.Type, &product.SKU, &product.Company,
		&product.Permalink, &product.Price, &product.SoldQuantity, &product.Tags,
		&product.Status, &product.Health, &product.ShipCostSlow, &product.ShipCostStandard,
		&product.ShipCostNextday, &picturesJSON, &createdAtSite, &updatedAtSite,
		&product.UpdatedAt, &product.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Parse pictures JSON
	if picturesJSON != "" {
		// Handle Python list format with single quotes
		cleanJSON := strings.ReplaceAll(picturesJSON, "'", "\"")

		// Try to parse as JSON array
		if err := json.Unmarshal([]byte(cleanJSON), &product.Pictures); err != nil {
			log.Printf("⚠️ Warning: Failed to parse pictures for product %s: %v", product.ID, err)
			log.Printf("⚠️ Raw pictures data: %s", picturesJSON)
			product.Pictures = []string{}
		} else {
			log.Printf("✅ Successfully parsed %d pictures for product %s", len(product.Pictures), product.ID)
		}
	} else {
		product.Pictures = []string{}
	}

	// Set site timestamps
	if createdAtSite.Valid {
		product.CreatedAtSite = createdAtSite
	}
	if updatedAtSite.Valid {
		product.UpdatedAtSite = updatedAtSite
	}

	return &product, nil
}

// CreateProduct creates a new product
func (s *DeParaService) CreateProduct(req models.CreateDeParaRequest, userID, userEmail, userName, ipAddress, userAgent string) error {
	actualTableName := s.buildTableName(req.TableName)
	query := fmt.Sprintf(`
		INSERT INTO %s (id, mlbu, type, sku, company, permalink, pictures)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)`, actualTableName)

	// Set default values
	if req.MLBU == "" {
		req.MLBU = req.ID
	}
	if req.Type == "" {
		req.Type = "product"
	}

	permalink := fmt.Sprintf("https://produto.mercadolivre.com.br/%s", req.ID)
	picturesJSON := "[]" // Empty array for new products

	_, err := s.db.Exec(query, req.ID, req.MLBU, req.Type, req.SKU, req.Company, permalink, picturesJSON)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Log audit
	auditReq := models.AuditLogRequest{
		TableName: actualTableName,
		RecordID:  req.ID,
		Operation: "INSERT",
		NewValues: map[string]interface{}{
			"id":        req.ID,
			"mlbu":      req.MLBU,
			"type":      req.Type,
			"sku":       req.SKU,
			"company":   req.Company,
			"permalink": permalink,
			"pictures":  picturesJSON,
		},
		ChangedFields: []string{"id", "mlbu", "type", "sku", "company", "permalink", "pictures"},
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}

	if err := s.auditService.LogOperation(auditReq, userID, userEmail, userName); err != nil {
		log.Printf("⚠️ Warning: Failed to log audit for CREATE: %v", err)
	}

	log.Printf("✅ Created product %s in amazonas_psa.mercadolivre_base", req.ID)
	return nil
}

// UpdateProduct updates an existing product
func (s *DeParaService) UpdateProduct(tableName, id string, req models.UpdateDeParaRequest, userID, userEmail, userName, ipAddress, userAgent string) error {
	actualTableName := s.buildTableName(tableName)

	// Get old values before update
	oldProduct, err := s.GetProductByID(tableName, id)
	if err != nil {
		return fmt.Errorf("failed to get product for audit: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE %s 
		SET sku = @p1, company = @p2, updated_at = GETDATE()
		WHERE id = @p3`, actualTableName)

	result, err := s.db.Exec(query, req.SKU, req.Company, id)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	// Log audit
	changedFields := []string{}
	oldValues := map[string]interface{}{}
	newValues := map[string]interface{}{}

	if oldProduct.SKU != req.SKU {
		changedFields = append(changedFields, "sku")
		oldValues["sku"] = oldProduct.SKU
		newValues["sku"] = req.SKU
	}
	if oldProduct.Company != req.Company {
		changedFields = append(changedFields, "company")
		oldValues["company"] = oldProduct.Company
		newValues["company"] = req.Company
	}

	if len(changedFields) > 0 {
		auditReq := models.AuditLogRequest{
			TableName:     actualTableName,
			RecordID:      id,
			Operation:     "UPDATE",
			OldValues:     oldValues,
			NewValues:     newValues,
			ChangedFields: changedFields,
			IPAddress:     ipAddress,
			UserAgent:     userAgent,
		}

		if err := s.auditService.LogOperation(auditReq, userID, userEmail, userName); err != nil {
			log.Printf("⚠️ Warning: Failed to log audit for UPDATE: %v", err)
		}
	}

	log.Printf("✅ Updated product %s in amazonas_psa.mercadolivre_base", id)
	return nil
}

// DeleteProduct deletes a product
func (s *DeParaService) DeleteProduct(tableName, id string, userID, userEmail, userName, ipAddress, userAgent string) error {
	actualTableName := s.buildTableName(tableName)

	// Get old values before delete
	oldProduct, err := s.GetProductByID(tableName, id)
	if err != nil {
		return fmt.Errorf("failed to get product for audit: %w", err)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = @p1`, actualTableName)

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	// Log audit
	auditReq := models.AuditLogRequest{
		TableName: actualTableName,
		RecordID:  id,
		Operation: "DELETE",
		OldValues: map[string]interface{}{
			"id":        oldProduct.ID,
			"mlbu":      oldProduct.MLBU,
			"type":      oldProduct.Type,
			"sku":       oldProduct.SKU,
			"company":   oldProduct.Company,
			"permalink": oldProduct.Permalink,
			"pictures":  oldProduct.Pictures,
		},
		ChangedFields: []string{"id", "mlbu", "type", "sku", "company", "permalink", "pictures"},
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}

	if err := s.auditService.LogOperation(auditReq, userID, userEmail, userName); err != nil {
		log.Printf("⚠️ Warning: Failed to log audit for DELETE: %v", err)
	}

	log.Printf("✅ Deleted product %s from amazonas_psa.mercadolivre_base", id)
	return nil
}

// buildTableName constructs table name from empresa, conta, marketplace
func (s *DeParaService) buildTableName(tableConfig string) string {
	// If tableConfig is already a full table name, return it
	if strings.Contains(tableConfig, ".") {
		return tableConfig
	}

	// For backward compatibility, if no tableConfig provided, return default
	if tableConfig == "" {
		return "integration.amazonas_psa.mercadolivre_base"
	}

	// Parse tableConfig format: "empresa_conta_marketplace" -> "integration.empresa_conta.marketplace_base"
	// Example: "amazonas_psa_mercadolivre" -> "integration.amazonas_psa.mercadolivre_base"
	parts := strings.Split(tableConfig, "_")
	if len(parts) >= 3 {
		empresa := parts[0]
		conta := parts[1]
		marketplace := parts[2]
		return fmt.Sprintf("integration.%s_%s.%s_base", empresa, conta, marketplace)
	}

	// Fallback to default table
	return "integration.amazonas_psa.mercadolivre_base"
}
