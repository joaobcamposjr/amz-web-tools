package handlers

import (
	"net/http"
	"strconv"

	"amz-web-tools/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetAvailableTables returns list of available integration tables
func (h *Handlers) GetAvailableTables(c *gin.Context) {
	tables, err := h.dePara.GetAvailableTables()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to get available tables",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Tables retrieved successfully",
		Data:    tables,
	})
}

// GetTableOptions returns available options for dynamic table selection
func (h *Handlers) GetTableOptions(c *gin.Context) {
	options, err := h.dePara.GetTableOptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to get table options",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Table options retrieved successfully",
		Data:    options,
	})
}

// SearchDeParaProducts searches products based on criteria
func (h *Handlers) SearchDeParaProducts(c *gin.Context) {
	var req models.DeParaSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	// Default pagination values
	page := 1
	pageSize := 15

	// Parse pagination parameters from query string
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	products, totalCount, err := h.dePara.SearchProducts(req.TableName, req.Query, req.SearchBy, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to search products",
			Error:   err.Error(),
		})
		return
	}

	// Calculate pagination info
	totalPages := (totalCount + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Products retrieved successfully",
		Data: gin.H{
			"products":    products,
			"count":       len(products),
			"total_count": totalCount,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}

// GetDeParaProduct gets a single product by ID
func (h *Handlers) GetDeParaProduct(c *gin.Context) {
	id := c.Param("id")
	tableName := c.Query("table")

	if tableName == "" {
		tableName = "MercadoLivre" // Default
	}

	product, err := h.dePara.GetProductByID(tableName, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Product not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Product retrieved successfully",
		Data:    product,
	})
}

// CreateDeParaProduct creates a new product
func (h *Handlers) CreateDeParaProduct(c *gin.Context) {
	var req models.CreateDeParaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	// Get user info for audit
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	userName := c.GetString("user_name")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err := h.dePara.CreateProduct(req, userID, userEmail, userName, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create product",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Product created successfully",
		Data: gin.H{
			"id": req.ID,
		},
	})
}

// UpdateDeParaProduct updates an existing product
func (h *Handlers) UpdateDeParaProduct(c *gin.Context) {
	id := c.Param("id")
	tableName := c.Query("table")

	if tableName == "" {
		tableName = "MercadoLivre" // Default
	}

	var req models.UpdateDeParaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	// Get user info for audit
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	userName := c.GetString("user_name")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err := h.dePara.UpdateProduct(tableName, id, req, userID, userEmail, userName, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update product",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Product updated successfully",
		Data: gin.H{
			"id": id,
		},
	})
}

// DeleteDeParaProduct deletes a product
func (h *Handlers) DeleteDeParaProduct(c *gin.Context) {
	id := c.Param("id")
	tableName := c.Query("table")

	if tableName == "" {
		tableName = "MercadoLivre" // Default
	}

	// Get user info for audit
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	userName := c.GetString("user_name")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err := h.dePara.DeleteProduct(tableName, id, userID, userEmail, userName, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete product",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Product deleted successfully",
		Data: gin.H{
			"id": id,
		},
	})
}

// GetAuditLogs retrieves audit logs for a specific product
func (h *Handlers) GetAuditLogs(c *gin.Context) {
	tableName := c.Query("table")
	recordID := c.Query("record_id")
	limitStr := c.Query("limit")

	if tableName == "" {
		tableName = "amazonas_psa.mercadolivre_base"
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	logs, err := h.audit.GetAuditLogs(tableName, recordID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve audit logs",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Audit logs retrieved successfully",
		Data: gin.H{
			"logs":  logs,
			"count": len(logs),
		},
	})
}

// ExecuteRollback executes a rollback operation
func (h *Handlers) ExecuteRollback(c *gin.Context) {
	auditLogID := c.Param("audit_id")
	if auditLogID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Audit log ID is required",
		})
		return
	}

	// Get user info for audit
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	userName := c.GetString("user_name")

	err := h.audit.ExecuteRollback(auditLogID, userID, userEmail, userName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to execute rollback",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Rollback executed successfully",
		Data: gin.H{
			"audit_log_id": auditLogID,
		},
	})
}

// GetDeParaProducts (legacy handler for compatibility)
func (h *Handlers) GetDeParaProducts(c *gin.Context) {
	// This is a legacy handler, redirect to search with empty query to get all
	var req models.DeParaSearchRequest
	req.TableName = c.DefaultQuery("table", "MercadoLivre")
	req.Query = ""
	req.SearchBy = "sku"

	products, totalCount, err := h.dePara.SearchProducts(req.TableName, req.Query, req.SearchBy, 1, 15)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to get products",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Products retrieved successfully",
		Data: gin.H{
			"products":    products,
			"count":       len(products),
			"total_count": totalCount,
		},
	})
}
