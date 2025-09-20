package middleware

import (
	"log"
	"net/http"
	"strings"

	"amz-web-tools/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("🔐 AuthMiddleware: Verificando autenticação para %s %s", c.Request.Method, c.Request.URL.Path)

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		log.Printf("🔐 AuthMiddleware: Authorization header = '%s'", authHeader)

		if authHeader == "" {
			log.Printf("❌ AuthMiddleware: Authorization header vazio")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check if token starts with "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		log.Printf("🔐 AuthMiddleware: Token extraído = '%s'", tokenString)

		if tokenString == authHeader {
			log.Printf("❌ AuthMiddleware: Token não tem formato Bearer")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		log.Printf("🔐 AuthMiddleware: JWT Secret = '%s'", jwtSecret)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("❌ AuthMiddleware: Método de assinatura inválido")
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Printf("❌ AuthMiddleware: Erro ao parsear token: %v", err)
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Printf("❌ AuthMiddleware: Token inválido")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		log.Printf("✅ AuthMiddleware: Token válido!")

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Set user information in context
		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid user ID in token",
			})
			c.Abort()
			return
		}

		userRole, ok := claims["role"].(string)
		if !ok {
			userRole = "user"
		}

		c.Set("user_id", userID)
		c.Set("user_role", userRole)
		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "User role not found",
			})
			c.Abort()
			return
		}

		role := userRole.(string)

		// Define role hierarchy (admin > operacao > atendimento)
		roleHierarchy := map[string]int{
			"atendimento": 1,
			"operacao":    2,
			"admin":       3,
		}

		userLevel, userExists := roleHierarchy[role]
		requiredLevel, requiredExists := roleHierarchy[requiredRole]

		if !userExists || !requiredExists || userLevel < requiredLevel {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminMiddleware checks if user is admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "User role not found",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		if role != "admin" {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
