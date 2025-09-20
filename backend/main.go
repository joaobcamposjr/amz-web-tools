package main

import (
	"log"
	"os"

	"amz-web-tools/backend/internal/config"
	"amz-web-tools/backend/internal/database"
	"amz-web-tools/backend/internal/handlers"
	"amz-web-tools/backend/internal/middleware"
	"amz-web-tools/backend/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("No .env file found at ../.env, trying .env: %v", err)
		if err2 := godotenv.Load(".env"); err2 != nil {
			log.Printf("No .env file found at .env either: %v", err2)
			log.Println("Using system environment variables")
		} else {
			log.Println("✅ Loaded .env file from current directory")
		}
	} else {
		log.Println("✅ Loaded .env file from parent directory")
	}

	// Initialize configuration
	cfg := config.Load()
	log.Printf("🔧 Config loaded: Oracle Host=%s, User=%s, Service=%s", cfg.OracleHost, cfg.OracleUser, cfg.OracleService)

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize Gin router
	r := gin.Default()

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORSAllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "Upgrade", "Connection", "Sec-WebSocket-Key", "Sec-WebSocket-Version", "Sec-WebSocket-Protocol"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize handlers
	h, err := handlers.New(db, cfg, wsHub)
	if err != nil {
		log.Fatalf("Failed to initialize handlers: %v", err)
	}

	// Public routes
	public := r.Group("/api/v1")
	{
		public.POST("/auth/login", h.Login)
		public.POST("/auth/register", h.Register)
		// Temporary route for testing car plate without auth
		public.GET("/test/plate-history", h.GetCarPlateHistory)
		public.GET("/test/car-plate/:plate", h.GetCarPlate)
		// Temporary route for testing depara without auth
		public.GET("/test/depara/tables", h.GetAvailableTables)
		public.GET("/test/depara/options", h.GetTableOptions)
		public.POST("/test/depara/search", h.SearchDeParaProducts)
		public.POST("/test/depara", h.CreateDeParaProduct)
		public.GET("/test/depara/:id", h.GetDeParaProduct)
		public.PUT("/test/depara/:id", h.UpdateDeParaProduct)
		public.DELETE("/test/depara/:id", h.DeleteDeParaProduct)

		// Test audit routes
		public.GET("/test/audit/logs", h.GetAuditLogs)
		public.POST("/test/audit/rollback/:audit_id", h.ExecuteRollback)

		// Test stock routes
		public.GET("/test/stock", h.GetStock)
		public.POST("/test/stock/search", h.SearchStock)

		// Test XML Integrator route (temporário para testes)
		public.POST("/test/xml-integrator/process", h.ProcessXMLIntegration)
		public.GET("/test/xml-integrator/logs/:process_id", h.GetXMLIntegrationLogs)

		// WebSocket route para logs em tempo real
		public.GET("/ws/logs", websocket.HandleWebSocket(wsHub))

		// Teste simples para verificar se a rota está funcionando
		public.GET("/test/ws", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "WebSocket route is working"})
		})
	}

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Profile routes
		protected.GET("/profile", h.GetProfile)
		protected.PUT("/profile", h.UpdateProfile)
		protected.PUT("/profile/password", h.UpdatePassword)

		// Car Plate routes
		protected.GET("/car-plate/:plate", h.GetCarPlate)
		protected.GET("/car-plate/history", h.GetCarPlateHistory)

		// Integration routes
		protected.POST("/integration/execute", h.ExecuteIntegration)
		protected.GET("/integration/status/:id", h.GetIntegrationStatus)

		// Import XML routes
		protected.POST("/import/xml", h.ImportXML)
		protected.GET("/import/status/:id", h.GetImportStatus)

		// XML Integrator routes
		protected.POST("/xml-integrator/process", h.ProcessXMLIntegration)
		protected.GET("/xml-integrator/logs/:process_id", h.GetXMLIntegrationLogs)

		// DePara routes
		protected.GET("/depara/tables", h.GetAvailableTables)
		protected.POST("/depara/search", h.SearchDeParaProducts)
		protected.GET("/depara", h.GetDeParaProducts)
		protected.POST("/depara", h.CreateDeParaProduct)
		protected.GET("/depara/:id", h.GetDeParaProduct)
		protected.PUT("/depara/:id", h.UpdateDeParaProduct)
		protected.DELETE("/depara/:id", h.DeleteDeParaProduct)

		// Audit routes
		protected.GET("/audit/logs", h.GetAuditLogs)
		protected.POST("/audit/rollback/:audit_id", h.ExecuteRollback)

		// Stock routes
		protected.GET("/stock", h.GetStock)
		protected.POST("/stock/search", h.SearchStock)

		// First login routes
		protected.POST("/auth/first-login", h.ChangePasswordFirstLogin)

		// Dashboard routes
		protected.GET("/dashboard/stats", h.GetDashboardStats)
		protected.GET("/dashboard/test-tables", h.TestDashboardTables)
		protected.POST("/dashboard/populate-test", h.PopulateTestData)
		protected.GET("/dashboard/debug-queries", h.DebugQueries)
		protected.GET("/test/tables-check", h.TestTablesCheck)
	}

	// Admin-only routes
	admin := r.Group("/api/v1")
	admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	admin.Use(middleware.AdminMiddleware())
	{
		// User management routes
		admin.POST("/users", h.CreateUser)
		admin.GET("/users", h.GetAllUsers)
		admin.PUT("/users", h.UpdateUser)
		admin.POST("/users/reset-password", h.ResetUserPassword)
		admin.DELETE("/users/:id", h.DeleteUser)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
