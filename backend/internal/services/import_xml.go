package services

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"time"

	"amz-web-tools/backend/internal/models"

	"github.com/google/uuid"
)

type ImportXMLService struct {
	db *sql.DB
}

func NewImportXMLService(db *sql.DB) *ImportXMLService {
	return &ImportXMLService{
		db: db,
	}
}

// XMLProduct represents a product in XML format
type XMLProduct struct {
	ID        string `xml:"id"`
	SKU       string `xml:"sku"`
	Name      string `xml:"name"`
	Price     string `xml:"price"`
	Category  string `xml:"category"`
	Brand     string `xml:"brand"`
	Available string `xml:"available"`
}

// XMLData represents the root XML structure
type XMLData struct {
	Products []XMLProduct `xml:"product"`
}

// ImportXML processes XML file upload and starts import process
func (s *ImportXMLService) ImportXML(filename, content, userID string) (*models.ImportXMLResponse, error) {
	// Generate unique ID for this import
	importID := uuid.New().String()

	// Create import log entry
	query := `
		INSERT INTO import_logs (id, user_id, file_name, status, processed_records, total_records, created_at)
		VALUES (?, ?, ?, 'pending', 0, 0, GETDATE())
	`

	_, err := s.db.Exec(query, importID, userID, filename)
	if err != nil {
		log.Printf("❌ Error creating import log: %v", err)
		return nil, fmt.Errorf("failed to create import log: %w", err)
	}

	// Start processing in background
	go s.processXML(importID, filename, content, userID)

	return &models.ImportXMLResponse{
		ID:       importID,
		Filename: filename,
		Status:   "pending",
		Message:  "Arquivo XML enviado com sucesso. Processamento iniciado.",
	}, nil
}

// processXML processes the XML content in background
func (s *ImportXMLService) processXML(importID, filename, content, userID string) {
	log.Printf("🔄 Starting XML processing for import ID: %s", importID)

	// Update status to processing
	s.updateImportStatus(importID, "processing", "Iniciando processamento do arquivo XML", 0, 0, 0)

	// Parse XML
	var xmlData XMLData
	err := xml.Unmarshal([]byte(content), &xmlData)
	if err != nil {
		log.Printf("❌ Error parsing XML: %v", err)
		s.updateImportStatus(importID, "error", fmt.Sprintf("Erro ao processar XML: %v", err), 0, 0, 0)
		return
	}

	totalRecords := len(xmlData.Products)
	log.Printf("📊 Found %d products in XML", totalRecords)

	// Update total records
	s.updateImportStatus(importID, "processing", fmt.Sprintf("Processando %d produtos", totalRecords), 0, totalRecords, 0)

	// Process each product
	processedRecords := 0
	for i, product := range xmlData.Products {
		// Simulate processing time
		time.Sleep(100 * time.Millisecond)

		// Process product (here you would insert into your target table)
		err := s.processProduct(product)
		if err != nil {
			log.Printf("❌ Error processing product %d: %v", i+1, err)
			continue
		}

		processedRecords++
		progress := int(float64(processedRecords) / float64(totalRecords) * 100)

		// Update progress every 10 records or at the end
		if processedRecords%10 == 0 || processedRecords == totalRecords {
			s.updateImportStatus(importID, "processing",
				fmt.Sprintf("Processando produto %d de %d", processedRecords, totalRecords),
				progress, totalRecords, processedRecords)
		}
	}

	// Mark as completed
	s.updateImportStatus(importID, "completed",
		fmt.Sprintf("Importação concluída com sucesso. %d produtos processados.", processedRecords),
		100, totalRecords, processedRecords)

	// Update completed_at timestamp
	query := `UPDATE import_logs SET completed_at = GETDATE() WHERE id = ?`
	_, err = s.db.Exec(query, importID)
	if err != nil {
		log.Printf("❌ Error updating completed_at: %v", err)
	}

	log.Printf("✅ XML processing completed for import ID: %s", importID)
}

// processProduct processes a single product
func (s *ImportXMLService) processProduct(product XMLProduct) error {
	// Here you would implement your actual product processing logic
	// For example, insert into a products table, update existing records, etc.

	log.Printf("📦 Processing product: ID=%s, SKU=%s, Name=%s", product.ID, product.SKU, product.Name)

	// Example: Insert into a products table (you would need to create this table)
	// For now, we'll just log the product data
	return nil
}

// updateImportStatus updates the import status in database
func (s *ImportXMLService) updateImportStatus(importID, status, message string, progress, totalRecords, processedRecords int) {
	query := `
		UPDATE import_logs 
		SET status = ?, processed_records = ?, total_records = ?, error_message = ?
		WHERE id = ?
	`

	var errorMessage string
	if status == "error" {
		errorMessage = message
	}

	_, err := s.db.Exec(query, status, processedRecords, totalRecords, errorMessage, importID)
	if err != nil {
		log.Printf("❌ Error updating import status: %v", err)
	}
}

// GetImportStatus retrieves import status by ID
func (s *ImportXMLService) GetImportStatus(importID string) (*models.ImportStatus, error) {
	query := `
		SELECT id, file_name, status, processed_records, total_records, error_message, created_at, completed_at
		FROM import_logs 
		WHERE id = ?
	`

	var status models.ImportStatus
	var errorMessage sql.NullString
	var completedAt sql.NullTime

	err := s.db.QueryRow(query, importID).Scan(
		&status.ID,
		&status.Filename,
		&status.Status,
		&status.RecordsProcessed,
		&status.RecordsTotal,
		&errorMessage,
		&status.CreatedAt,
		&completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("import not found")
		}
		return nil, fmt.Errorf("failed to get import status: %w", err)
	}

	// Set optional fields
	if errorMessage.Valid {
		status.Error = errorMessage.String
	}
	if completedAt.Valid {
		status.CompletedAt = &completedAt.Time
	}

	// Calculate progress
	if status.RecordsTotal > 0 {
		status.Progress = int(float64(status.RecordsProcessed) / float64(status.RecordsTotal) * 100)
	}

	// Set message based on status
	switch status.Status {
	case "pending":
		status.Message = "Aguardando processamento"
	case "processing":
		status.Message = fmt.Sprintf("Processando %d de %d registros", status.RecordsProcessed, status.RecordsTotal)
	case "completed":
		status.Message = fmt.Sprintf("Importação concluída com sucesso. %d registros processados.", status.RecordsProcessed)
	case "error":
		status.Message = "Erro durante o processamento"
	}

	return &status, nil
}

// GetImportLogs retrieves all import logs for a user
func (s *ImportXMLService) GetImportLogs(userID string, limit int) ([]models.ImportStatus, error) {
	query := `
		SELECT id, file_name, status, processed_records, total_records, error_message, created_at, completed_at
		FROM import_logs 
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" OFFSET 0 ROWS FETCH NEXT %d ROWS ONLY", limit)
	}

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import logs: %w", err)
	}
	defer rows.Close()

	var logs []models.ImportStatus
	for rows.Next() {
		var log models.ImportStatus
		var errorMessage sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&log.ID,
			&log.Filename,
			&log.Status,
			&log.RecordsProcessed,
			&log.RecordsTotal,
			&errorMessage,
			&log.CreatedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan import log: %w", err)
		}

		// Set optional fields
		if errorMessage.Valid {
			log.Error = errorMessage.String
		}
		if completedAt.Valid {
			log.CompletedAt = &completedAt.Time
		}

		// Calculate progress
		if log.RecordsTotal > 0 {
			log.Progress = int(float64(log.RecordsProcessed) / float64(log.RecordsTotal) * 100)
		}

		// Set message based on status
		switch log.Status {
		case "pending":
			log.Message = "Aguardando processamento"
		case "processing":
			log.Message = fmt.Sprintf("Processando %d de %d registros", log.RecordsProcessed, log.RecordsTotal)
		case "completed":
			log.Message = fmt.Sprintf("Importação concluída com sucesso. %d registros processados.", log.RecordsProcessed)
		case "error":
			log.Message = "Erro durante o processamento"
		}

		logs = append(logs, log)
	}

	return logs, nil
}
