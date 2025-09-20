package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"amz-web-tools/backend/internal/config"
	"amz-web-tools/backend/internal/models"
)

type AuditService struct {
	db     *sql.DB
	config *config.Config
}

func NewAuditService(db *sql.DB, cfg *config.Config) *AuditService {
	return &AuditService{
		db:     db,
		config: cfg,
	}
}

// LogOperation logs a CRUD operation to the audit table
func (s *AuditService) LogOperation(req models.AuditLogRequest, userID, userEmail, userName string) error {
	// Convert maps to JSON strings
	oldValuesJSON, err := json.Marshal(req.OldValues)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to marshal old values: %v", err)
		oldValuesJSON = []byte("{}")
	}

	newValuesJSON, err := json.Marshal(req.NewValues)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to marshal new values: %v", err)
		newValuesJSON = []byte("{}")
	}

	changedFieldsJSON, err := json.Marshal(req.ChangedFields)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to marshal changed fields: %v", err)
		changedFieldsJSON = []byte("[]")
	}

	// Create rollback data based on operation
	var rollbackData string
	switch strings.ToUpper(req.Operation) {
	case "INSERT":
		// For INSERT, rollback would be DELETE
		rollbackData = fmt.Sprintf(`{"operation": "DELETE", "record_id": "%s"}`, req.RecordID)
	case "UPDATE":
		// For UPDATE, rollback would be UPDATE with old values
		rollbackData = fmt.Sprintf(`{"operation": "UPDATE", "record_id": "%s", "values": %s}`, req.RecordID, string(oldValuesJSON))
	case "DELETE":
		// For DELETE, rollback would be INSERT with old values
		rollbackData = fmt.Sprintf(`{"operation": "INSERT", "values": %s}`, string(oldValuesJSON))
	}

	// Insert into audit_logs table in portal database
	query := `
		INSERT INTO portal.dbo.audit_logs 
		(table_name, record_id, operation, user_id, user_email, user_name, 
		 old_values, new_values, changed_fields, ip_address, user_agent, rollback_data)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11, @p12)`

	_, err = s.db.Exec(query,
		req.TableName,
		req.RecordID,
		strings.ToUpper(req.Operation),
		userID,
		userEmail,
		userName,
		string(oldValuesJSON),
		string(newValuesJSON),
		string(changedFieldsJSON),
		req.IPAddress,
		req.UserAgent,
		rollbackData,
	)

	if err != nil {
		log.Printf("❌ Failed to log audit operation: %v", err)
		return fmt.Errorf("failed to log audit operation: %w", err)
	}

	log.Printf("✅ Audit log created for %s operation on %s.%s", req.Operation, req.TableName, req.RecordID)
	return nil
}

// GetAuditLogs retrieves audit logs for a specific table and record
func (s *AuditService) GetAuditLogs(tableName, recordID string, limit int) ([]models.AuditLog, error) {
	query := `
		SELECT TOP (@p1) id, table_name, record_id, operation, user_id, user_email, user_name,
		       old_values, new_values, changed_fields, ip_address, user_agent, created_at, rollback_data
		FROM portal.dbo.audit_logs 
		WHERE table_name = @p2 AND record_id = @p3
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query, limit, tableName, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(&log.ID, &log.TableName, &log.RecordID, &log.Operation,
			&log.UserID, &log.UserEmail, &log.UserName, &log.OldValues, &log.NewValues,
			&log.ChangedFields, &log.IPAddress, &log.UserAgent, &log.CreatedAt, &log.RollbackData)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetAuditLogsByUser retrieves audit logs for a specific user
func (s *AuditService) GetAuditLogsByUser(userID string, limit int) ([]models.AuditLog, error) {
	query := `
		SELECT TOP (@p1) id, table_name, record_id, operation, user_id, user_email, user_name,
		       old_values, new_values, changed_fields, ip_address, user_agent, created_at, rollback_data
		FROM portal.dbo.audit_logs 
		WHERE user_id = @p2
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query, limit, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(&log.ID, &log.TableName, &log.RecordID, &log.Operation,
			&log.UserID, &log.UserEmail, &log.UserName, &log.OldValues, &log.NewValues,
			&log.ChangedFields, &log.IPAddress, &log.UserAgent, &log.CreatedAt, &log.RollbackData)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// ExecuteRollback executes a rollback operation based on audit log
func (s *AuditService) ExecuteRollback(auditLogID string, userID, userEmail, userName string) error {
	// Get the audit log
	query := `
		SELECT table_name, record_id, operation, rollback_data
		FROM portal.dbo.audit_logs 
		WHERE id = @p1`

	var tableName, recordID, operation, rollbackData string
	err := s.db.QueryRow(query, auditLogID).Scan(&tableName, &recordID, &operation, &rollbackData)
	if err != nil {
		return fmt.Errorf("failed to get audit log: %w", err)
	}

	// Parse rollback data
	var rollbackInfo map[string]interface{}
	if err := json.Unmarshal([]byte(rollbackData), &rollbackInfo); err != nil {
		return fmt.Errorf("failed to parse rollback data: %w", err)
	}

	// Execute rollback based on original operation
	switch strings.ToUpper(operation) {
	case "INSERT":
		// Rollback INSERT = DELETE
		deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE id = @p1", tableName)
		_, err = s.db.Exec(deleteQuery, recordID)
		if err != nil {
			return fmt.Errorf("failed to execute rollback DELETE: %w", err)
		}

	case "UPDATE":
		// Rollback UPDATE = UPDATE with old values
		values, ok := rollbackInfo["values"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid rollback data for UPDATE")
		}

		// Build UPDATE query dynamically
		setParts := []string{}
		args := []interface{}{}
		argIndex := 1

		for field, value := range values {
			if field != "id" { // Skip ID field
				setParts = append(setParts, fmt.Sprintf("%s = @p%d", field, argIndex))
				args = append(args, value)
				argIndex++
			}
		}

		if len(setParts) > 0 {
			updateQuery := fmt.Sprintf("UPDATE %s SET %s WHERE id = @p%d",
				tableName, strings.Join(setParts, ", "), argIndex)
			args = append(args, recordID)

			_, err = s.db.Exec(updateQuery, args...)
			if err != nil {
				return fmt.Errorf("failed to execute rollback UPDATE: %w", err)
			}
		}

	case "DELETE":
		// Rollback DELETE = INSERT with old values
		values, ok := rollbackInfo["values"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid rollback data for DELETE")
		}

		// Build INSERT query dynamically
		fields := []string{}
		placeholders := []string{}
		args := []interface{}{}

		for field, value := range values {
			fields = append(fields, field)
			placeholders = append(placeholders, fmt.Sprintf("@p%d", len(args)+1))
			args = append(args, value)
		}

		if len(fields) > 0 {
			insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				tableName, strings.Join(fields, ", "), strings.Join(placeholders, ", "))

			_, err = s.db.Exec(insertQuery, args...)
			if err != nil {
				return fmt.Errorf("failed to execute rollback INSERT: %w", err)
			}
		}
	}

	// Log the rollback operation
	rollbackReq := models.AuditLogRequest{
		TableName: tableName,
		RecordID:  recordID,
		Operation: "ROLLBACK",
		NewValues: map[string]interface{}{
			"original_operation": operation,
			"original_audit_id":  auditLogID,
		},
	}

	return s.LogOperation(rollbackReq, userID, userEmail, userName)
}







