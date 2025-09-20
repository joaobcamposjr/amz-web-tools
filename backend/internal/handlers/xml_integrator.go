package handlers

import (
	"net/http"

	"amz-web-tools/backend/internal/models"

	"github.com/gin-gonic/gin"
)

type XMLIntegratorRequest struct {
	NumPedido string `json:"num_pedido" binding:"required"`
}

// ProcessXMLIntegration processes XML integration for a specific order
func (h *Handlers) ProcessXMLIntegration(c *gin.Context) {
	var req XMLIntegratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Parâmetros inválidos",
			Error:   err.Error(),
		})
		return
	}

	if req.NumPedido == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "num_pedido é obrigatório",
		})
		return
	}

	// Processar integração XML
	result, err := h.xmlIntegrator.ProcessXMLIntegration(req.NumPedido)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Erro ao processar integração XML",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetXMLIntegrationLogs returns logs for a specific process
func (h *Handlers) GetXMLIntegrationLogs(c *gin.Context) {
	processID := c.Param("process_id")
	if processID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Process ID é obrigatório",
		})
		return
	}

	// Buscar logs do serviço
	logs, err := h.xmlIntegrator.GetLogs(processID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Erro ao recuperar logs",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Logs recuperados com sucesso",
		Data: map[string]interface{}{
			"logs": logs,
		},
	})
}
