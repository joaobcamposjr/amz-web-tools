package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"amz-web-tools/backend/internal/models"
	"amz-web-tools/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// GetCarPlate retrieves car plate information with caching
func (h *Handlers) GetCarPlate(c *gin.Context) {
	plate := c.Param("plate")
	if plate == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Plate parameter is required",
		})
		return
	}

	// Get user ID from context (if authenticated)
	userID := ""
	if userIDInterface, exists := c.Get("user_id"); exists {
		if userIDStr, ok := userIDInterface.(string); ok {
			userID = userIDStr
		}
	}

	// Get plate data using service
	plateResult, err := h.carPlate.GetCarPlate(plate, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Plate not found or API error",
			Error:   err.Error(),
		})
		return
	}

	// Extract chassi if available
	var chassi string
	if plateResult.Data != nil {
		// First try complete chassi from extra object
		if extra, exists := (*plateResult.Data)["extra"]; exists {
			if extraMap, ok := extra.(map[string]interface{}); ok {
				if chassiValue, chassiExists := extraMap["chassi"]; chassiExists {
					if chassiStr, chassiOk := chassiValue.(string); chassiOk && chassiStr != "" {
						chassi = chassiStr
					}
				}
			}
		}
		// If not found in extra, try direct chassi field (with ***)
		if chassi == "" {
			if chassiValue, exists := (*plateResult.Data)["chassi"]; exists {
				if chassiStr, ok := chassiValue.(string); ok && chassiStr != "" {
					chassi = chassiStr
				}
			}
		}
	}

	// Extract brand logo if available
	var brandLogo string
	if plateResult.Data != nil {
		if logoValue, exists := (*plateResult.Data)["logo"]; exists {
			if logoStr, ok := logoValue.(string); ok && logoStr != "" {
				brandLogo = logoStr
			}
		}
	}

	responseData := gin.H{
		"plate_data": plateResult.Data,
		"source":     plateResult.Source,
		"plate":      strings.ToUpper(plate),
	}

	// Add chassi only if it exists
	if chassi != "" {
		responseData["chassi"] = chassi
	}

	// Add brand logo only if it exists
	if brandLogo != "" {
		responseData["brand_logo"] = brandLogo
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Plate data retrieved successfully",
		Data:    responseData,
	})
}

// GetCarPlateHistory retrieves car plate consultation history
func (h *Handlers) GetCarPlateHistory(c *gin.Context) {
	// Get user ID from context (if authenticated)
	userID := ""
	if userIDInterface, exists := c.Get("user_id"); exists {
		if userIDStr, ok := userIDInterface.(string); ok {
			userID = userIDStr
		}
	}

	// Get limit from query parameter (default to 10)
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Get history from service
	history, err := h.carPlate.GetPlateHistory(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve plate history",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Plate history retrieved successfully",
		Data: gin.H{
			"history": history,
			"count":   len(history),
		},
	})
}

// ExecuteIntegration executes an integration process
func (h *Handlers) ExecuteIntegration(c *gin.Context) {
	var req models.ExecuteIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	// Process integration
	integrationReq := services.IntegrationRequest{
		Conta:       req.Conta,
		Marketplace: req.Marketplace,
		NumPedido:   req.NumPedido,
	}

	result, err := h.integration.ProcessIntegration(integrationReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process integration",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Integration processed successfully",
		Data:    result,
	})
}

// GetIntegrationStatus retrieves integration execution status
func (h *Handlers) GetIntegrationStatus(c *gin.Context) {
	integrationID := c.Param("id")
	if integrationID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Integration ID is required",
		})
		return
	}

	// TODO: Implement integration status retrieval
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"integration_id": integrationID,
			"status":         "completed",
			"progress":       100,
		},
	})
}

// ImportXML handles XML file import
func (h *Handlers) ImportXML(c *gin.Context) {
	// TODO: Implement XML file upload and processing
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "XML import service not implemented yet",
	})
}

// GetImportStatus retrieves import process status
func (h *Handlers) GetImportStatus(c *gin.Context) {
	importID := c.Param("id")
	if importID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Import ID is required",
		})
		return
	}

	// TODO: Implement import status retrieval
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"import_id":         importID,
			"status":            "completed",
			"processed_records": 100,
			"total_records":     100,
		},
	})
}

// GetStock retrieves stock information
func (h *Handlers) GetStock(c *gin.Context) {
	sku := c.Query("sku")

	if sku == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "SKU parameter is required",
		})
		return
	}

	items, err := h.stock.SearchStock(sku)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to search stock",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Stock search completed",
		Data: gin.H{
			"sku":   sku,
			"items": items,
			"count": len(items),
		},
	})
}

// SearchStock searches for stock by SKU (POST endpoint)
func (h *Handlers) SearchStock(c *gin.Context) {
	var req models.StockSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	if req.SKU == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "SKU is required",
		})
		return
	}

	items, err := h.stock.SearchStock(req.SKU)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to search stock",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Stock search completed",
		Data: gin.H{
			"sku":   req.SKU,
			"items": items,
			"count": len(items),
		},
	})
}

// ===== USER MANAGEMENT HANDLERS (Admin only) =====

// CreateUser creates a new user (Admin only)
func (h *Handlers) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	user, err := h.auth.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	})
}

// GetAllUsers retrieves all users (Admin only)
func (h *Handlers) GetAllUsers(c *gin.Context) {
	users, err := h.auth.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve users",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

// UpdateUser updates a user (Admin only)
func (h *Handlers) UpdateUser(c *gin.Context) {
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	user, err := h.auth.UpdateUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    user,
	})
}

// ResetUserPassword resets a user's password (Admin only)
func (h *Handlers) ResetUserPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	err := h.auth.ResetUserPassword(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Failed to reset password",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}

// DeleteUser deletes a user (Admin only)
func (h *Handlers) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "User ID is required",
		})
		return
	}

	err := h.auth.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// ===== FIRST LOGIN HANDLERS =====

// ChangePasswordFirstLogin changes password on first login
func (h *Handlers) ChangePasswordFirstLogin(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	var req models.FirstLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	err := h.auth.ChangePasswordFirstLogin(userID.(string), req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Failed to change password",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// ===== DASHBOARD STATISTICS HANDLERS =====

// GetDashboardStats retrieves dashboard statistics
func (h *Handlers) GetDashboardStats(c *gin.Context) {
	stats, err := h.carPlate.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve dashboard statistics",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Dashboard statistics retrieved successfully",
		Data:    stats,
	})
}

// DebugQueries debugs the dashboard queries
func (h *Handlers) TestTablesCheck(c *gin.Context) {
	// Test the verified tables list
	verifiedTables := []string{
		"integration.amazonas_psa.mercadolivre_base",
		"integration.amazonas_renault.mercadolivre_base",
		"integration.amazonas_principal.mercadolivre_base",
		"integration.amazonas_oficial.mercadolivre_base",
		"integration.amazonas_jeep.mercadolivre_base",
		"integration.amazonas_ford.mercadolivre_base",
	}

	var tables []map[string]interface{}
	totalCount := 0

	for _, tableName := range verifiedTables {
		// Check if table exists by trying to query it
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		var count int
		err := h.db.QueryRow(query).Scan(&count)
		if err != nil {
			tables = append(tables, map[string]interface{}{
				"table":      tableName,
				"count":      0,
				"error":      err.Error(),
				"accessible": false,
			})
		} else {
			tables = append(tables, map[string]interface{}{
				"table":      tableName,
				"count":      count,
				"error":      nil,
				"accessible": true,
			})
			totalCount += count
		}
	}

	c.JSON(200, gin.H{
		"success":           true,
		"message":           "Tables verification completed",
		"data":              tables,
		"total_count":       totalCount,
		"accessible_tables": len(tables),
	})
}

func (h *Handlers) DebugQueries(c *gin.Context) {
	debug := gin.H{
		"postgresql_config": gin.H{
			"host":     h.carPlate.Config.PGHost,
			"port":     h.carPlate.Config.PGPort,
			"user":     h.carPlate.Config.PGUser,
			"database": h.carPlate.Config.PGDatabase,
		},
		"oracle_config": gin.H{
			"host":    h.carPlate.Config.OracleHost,
			"port":    h.carPlate.Config.OraclePort,
			"user":    h.carPlate.Config.OracleUser,
			"service": h.carPlate.Config.OracleService,
		},
		"xml_imports_count": h.carPlate.GetXMLImportsCount(),
		"stock_items_count": h.carPlate.GetStockItemsCount(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Debug queries executed",
		Data:    debug,
	})
}

// TestDashboardTables tests which tables exist in the database
func (h *Handlers) TestDashboardTables(c *gin.Context) {
	// Test each table individually
	tables := map[string]string{
		"car_plate_history": "SELECT COUNT(*) FROM car_plate_history",
		"depara_products":   "SELECT COUNT(*) FROM depara_products",
		"xml_imports":       "SELECT COUNT(*) FROM xml_imports",
		"stock_items":       "SELECT COUNT(*) FROM stock_items",
		"users":             "SELECT COUNT(*) FROM users",
	}

	results := make(map[string]interface{})

	for tableName, query := range tables {
		var count int
		err := h.carPlate.GetDB().QueryRow(query).Scan(&count)
		if err != nil {
			results[tableName] = map[string]interface{}{
				"exists": false,
				"error":  err.Error(),
				"count":  0,
			}
		} else {
			results[tableName] = map[string]interface{}{
				"exists": true,
				"error":  nil,
				"count":  count,
			}
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Table test results",
		Data:    results,
	})
}

// PopulateTestData populates the dashboard with test data
func (h *Handlers) PopulateTestData(c *gin.Context) {
	// Insert test data into car_plate_history
	_, err := h.carPlate.GetDB().Exec(`
		INSERT INTO car_plate_history (id, plate, response_data, status, error_message, created_at, user_id)
		VALUES 
		(NEWID(), 'ABC1234', '{"test": "data"}', 'success', '', GETDATE(), '6D9D42FE-0F7F-490F-B081-D1811ECB8AF1'),
		(NEWID(), 'XYZ5678', '{"test": "data"}', 'success', '', GETDATE(), '6D9D42FE-0F7F-490F-B081-D1811ECB8AF1'),
		(NEWID(), 'DEF9012', '{"test": "data"}', 'success', '', GETDATE(), '6D9D42FE-0F7F-490F-B081-D1811ECB8AF1')
	`)
	if err != nil {
		log.Printf("Error inserting test car plate data: %v", err)
	}

	// Insert test data into depara_products
	_, err = h.carPlate.GetDB().Exec(`
		INSERT INTO depara_products (id, mlbu, type, sku, company, permalink, ship_cost_slow, ship_cost_standard, ship_cost_nextday, pictures, created_at, updated_at)
		VALUES 
		(NEWID(), 'MLBU123456', 'product', 'SKU001', 'Company A', 'https://example.com/1', 10.50, 15.00, 25.00, '["https://example.com/img1.jpg"]', GETDATE(), GETDATE()),
		(NEWID(), 'MLBU789012', 'product', 'SKU002', 'Company B', 'https://example.com/2', 12.00, 18.00, 30.00, '["https://example.com/img2.jpg"]', GETDATE(), GETDATE()),
		(NEWID(), 'MLBU345678', 'product', 'SKU003', 'Company C', 'https://example.com/3', 8.75, 12.50, 20.00, '["https://example.com/img3.jpg"]', GETDATE(), GETDATE()),
		(NEWID(), 'MLBU901234', 'product', 'SKU004', 'Company D', 'https://example.com/4', 15.25, 22.00, 35.00, '["https://example.com/img4.jpg"]', GETDATE(), GETDATE()),
		(NEWID(), 'MLBU567890', 'product', 'SKU005', 'Company E', 'https://example.com/5', 9.99, 14.99, 24.99, '["https://example.com/img5.jpg"]', GETDATE(), GETDATE())
	`)
	if err != nil {
		log.Printf("Error inserting test depara data: %v", err)
	}

	// Insert test data into xml_integrator_logs
	_, err = h.carPlate.GetDB().Exec(`
		INSERT INTO xml_integrator_logs (id, process_id, level, step, message, created_at)
		VALUES 
		(NEWID(), 'test-process-1', 'info', '1', 'Test XML import 1', GETDATE()),
		(NEWID(), 'test-process-2', 'info', '1', 'Test XML import 2', GETDATE()),
		(NEWID(), 'test-process-3', 'info', '1', 'Test XML import 3', GETDATE())
	`)
	if err != nil {
		log.Printf("Error inserting test XML data: %v", err)
	}

	// Insert test data into stock_items
	_, err = h.carPlate.GetDB().Exec(`
		INSERT INTO stock_items (id, cod_item, nome_item, estoque, reservado, created_at)
		VALUES 
		(NEWID(), 'ITEM001', 'Test Item 1', 100, 10, GETDATE()),
		(NEWID(), 'ITEM002', 'Test Item 2', 50, 5, GETDATE()),
		(NEWID(), 'ITEM003', 'Test Item 3', 75, 8, GETDATE()),
		(NEWID(), 'ITEM004', 'Test Item 4', 200, 20, GETDATE()),
		(NEWID(), 'ITEM005', 'Test Item 5', 30, 3, GETDATE()),
		(NEWID(), 'ITEM006', 'Test Item 6', 150, 15, GETDATE())
	`)
	if err != nil {
		log.Printf("Error inserting test stock data: %v", err)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Test data populated successfully",
	})
}
