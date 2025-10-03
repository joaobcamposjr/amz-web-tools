package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TestPlateCache tests the plate cache functionality
func (h *Handlers) TestPlateCache(c *gin.Context) {
	// Test plate
	testPlate := "ABC1234"

	// Get user ID from context
	userID := ""
	if userIDInterface, exists := c.Get("user_id"); exists {
		if userIDStr, ok := userIDInterface.(string); ok {
			userID = userIDStr
		}
	}

	// Test cache functionality
	result, err := h.carPlate.GetCarPlate(testPlate, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to test plate cache",
			"error":   err.Error(),
		})
		return
	}

	// Check cache status
	cacheStats, err := h.carPlate.GetCacheStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get cache stats",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Plate cache test completed",
		"test_plate":  testPlate,
		"source":      result.Source,
		"cache_stats": cacheStats,
		"data":        result.Data,
	})
}
