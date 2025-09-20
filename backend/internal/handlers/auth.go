package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"amz-web-tools/backend/internal/config"
	"amz-web-tools/backend/internal/models"
	"amz-web-tools/backend/internal/services"
	"amz-web-tools/backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	db            *sql.DB
	config        *config.Config
	auth          *services.AuthService
	carPlate      *services.CarPlateService
	dePara        *services.DeParaService
	audit         *services.AuditService
	stock         *services.StockService
	importXML     *services.ImportXMLService
	xmlIntegrator *services.XMLIntegratorService
	integration   *services.IntegrationService
}

func New(db *sql.DB, cfg *config.Config, wsHub *websocket.Hub) (*Handlers, error) {
	auditService := services.NewAuditService(db, cfg)

	stockService, err := services.NewStockService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stock service: %w", err)
	}

	importXMLService := services.NewImportXMLService(db)

	xmlIntegratorService, err := services.NewXMLIntegratorService(cfg, wsHub)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize XML integrator service: %w", err)
	}

	integrationService := services.NewIntegrationService(db, stockService.GetOracleDB(), xmlIntegratorService.GetPostgresDB())

	return &Handlers{
		db:            db,
		config:        cfg,
		auth:          services.NewAuthService(db),
		carPlate:      services.NewCarPlateService(db, cfg),
		dePara:        services.NewDeParaService(db, auditService),
		audit:         auditService,
		stock:         stockService,
		importXML:     importXMLService,
		xmlIntegrator: xmlIntegratorService,
		integration:   integrationService,
	}, nil
}

// Login handles user login
func (h *Handlers) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	// Authenticate user
	user, err := h.auth.LoginUser(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	// Generate JWT token
	token, err := h.auth.GenerateJWT(user.ID, user.Role, h.config.JWTSecret, h.config.JWTExpireHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data: gin.H{
			"token": token,
			"user": gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"name":       user.Name,
				"department": user.Department,
				"role":       user.Role,
			},
		},
	})
}

// Register handles user registration
func (h *Handlers) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	// Register user
	user, err := h.auth.RegisterUser(&req)
	if err != nil {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Message: "Email already exists",
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data: gin.H{
			"user": gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"name":       user.Name,
				"department": user.Department,
				"role":       user.Role,
			},
		},
	})
}

// GetProfile retrieves user profile
func (h *Handlers) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.auth.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"user": gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"name":       user.Name,
				"department": user.Department,
				"role":       user.Role,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			},
		},
	})
}

// UpdateProfile updates user profile
func (h *Handlers) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	user, err := h.auth.UpdateUserProfile(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data: gin.H{
			"user": gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"name":       user.Name,
				"department": user.Department,
				"role":       user.Role,
				"updated_at": user.UpdatedAt,
			},
		},
	})
}

// UpdatePassword updates user password
func (h *Handlers) UpdatePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	err := h.auth.UpdateUserPassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Current password is incorrect",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password updated successfully",
	})
}
