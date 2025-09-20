package models

import (
	"database/sql"
	"time"
)

// User represents a user in the system
type User struct {
	ID                string    `json:"id" db:"id"`
	Email             string    `json:"email" db:"email"`
	PasswordHash      string    `json:"-" db:"password_hash"`
	Name              string    `json:"name" db:"name"`
	Department        string    `json:"department" db:"department"`
	Role              string    `json:"role" db:"role"`
	IsFirstLogin      bool      `json:"is_first_login" db:"is_first_login"`
	PasswordChangedAt time.Time `json:"password_changed_at" db:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// UserSession represents a user session
type UserSession struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Name       string `json:"name" binding:"required"`
	Department string `json:"department"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name       string `json:"name" binding:"required"`
	Department string `json:"department"`
}

// UpdatePasswordRequest represents password update request
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// CreateUserRequest represents user creation request (Admin only)
type CreateUserRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Name       string `json:"name" binding:"required"`
	Department string `json:"department" binding:"required"`
	Role       string `json:"role" binding:"required,oneof=admin operacao atendimento"`
}

// UpdateUserRequest represents user update request (Admin only)
type UpdateUserRequest struct {
	ID         string `json:"id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Department string `json:"department" binding:"required"`
	Role       string `json:"role" binding:"required,oneof=admin operacao atendimento"`
}

// ResetPasswordRequest represents password reset request (Admin only)
type ResetPasswordRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// FirstLoginRequest represents first login password change
type FirstLoginRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// PlateCache represents cached plate data
type PlateCache struct {
	ID        string    `json:"id" db:"id"`
	Plate     string    `json:"plate" db:"plate"`
	Data      string    `json:"data" db:"data"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

// DeParaProduct represents a product in the DePara system (NEW STRUCTURE)
type DeParaProduct struct {
	ID               string          `json:"id" db:"id"`
	MLBU             string          `json:"mlbu" db:"mlbu"`
	Type             string          `json:"type" db:"type"`
	SKU              string          `json:"sku" db:"sku"`
	Company          string          `json:"company" db:"company"`
	Permalink        string          `json:"permalink" db:"permalink"`
	ShipCostSlow     sql.NullFloat64 `json:"ship_cost_slow" db:"ship_cost_slow"`
	ShipCostStandard sql.NullFloat64 `json:"ship_cost_standard" db:"ship_cost_standard"`
	ShipCostNextday  sql.NullFloat64 `json:"ship_cost_nextday" db:"ship_cost_nextday"`
	Pictures         []string        `json:"pictures" db:"pictures"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
}

// IntegrationTable represents available integration tables
type IntegrationTable struct {
	ID          string    `json:"id" db:"id"`
	TableName   string    `json:"table_name" db:"table_name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// DeParaSearchRequest represents search parameters
type DeParaSearchRequest struct {
	TableName string `json:"table_name" binding:"required"`
	Query     string `json:"query" binding:"required"`
	SearchBy  string `json:"search_by"` // "id", "mlbu", "sku"
}

// CreateDeParaRequest represents request to create a DePara product
type CreateDeParaRequest struct {
	TableName string `json:"table_name" binding:"required"`
	ID        string `json:"id" binding:"required"`
	SKU       string `json:"sku" binding:"required"`
	Company   string `json:"company" binding:"required"`
	MLBU      string `json:"mlbu"`
	Type      string `json:"type"`
}

// UpdateDeParaRequest represents request to update a DePara product
type UpdateDeParaRequest struct {
	SKU     string `json:"sku" binding:"required"`
	Company string `json:"company" binding:"required"`
}

// CarPlateHistory represents a car plate consultation history entry
type CarPlateHistory struct {
	ID           string    `json:"id" db:"id"`
	Plate        string    `json:"plate" db:"plate"`
	ResponseData string    `json:"response_data" db:"response_data"`
	Status       string    `json:"status" db:"status"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UserID       string    `json:"user_id" db:"user_id"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID            string    `json:"id" db:"id"`
	TableName     string    `json:"table_name" db:"table_name"`
	RecordID      string    `json:"record_id" db:"record_id"`
	Operation     string    `json:"operation" db:"operation"`
	UserID        string    `json:"user_id" db:"user_id"`
	UserEmail     string    `json:"user_email" db:"user_email"`
	UserName      string    `json:"user_name" db:"user_name"`
	OldValues     string    `json:"old_values" db:"old_values"`
	NewValues     string    `json:"new_values" db:"new_values"`
	ChangedFields string    `json:"changed_fields" db:"changed_fields"`
	IPAddress     string    `json:"ip_address" db:"ip_address"`
	UserAgent     string    `json:"user_agent" db:"user_agent"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	RollbackData  string    `json:"rollback_data" db:"rollback_data"`
}

// AuditLogRequest represents request to create audit log
type AuditLogRequest struct {
	TableName     string                 `json:"table_name" binding:"required"`
	RecordID      string                 `json:"record_id" binding:"required"`
	Operation     string                 `json:"operation" binding:"required"`
	OldValues     map[string]interface{} `json:"old_values"`
	NewValues     map[string]interface{} `json:"new_values"`
	ChangedFields []string               `json:"changed_fields"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
}

// StockItem represents a stock item from Oracle
type StockItem struct {
	CodEmpresa        int     `json:"cod_empresa" db:"cod_empresa"`
	NomeEmpresa       string  `json:"nome_empresa" db:"nome_empresa"`
	CodFornecedor     string  `json:"cod_fornecedor" db:"cod_fornecedor"`
	NomeFornecedor    string  `json:"nome_fornecedor" db:"nome_fornecedor"`
	CodItem           string  `json:"cod_item" db:"cod_item"`
	ValorReposicao    float64 `json:"valor_reposicao" db:"valor_reposicao"`
	CustoContabil     float64 `json:"custo_contabil" db:"custo_contabil"`
	ValorVenda        float64 `json:"valor_venda" db:"valor_venda"`
	Estoque           int     `json:"estoque" db:"estoque"`
	Reservado         int     `json:"reservado" db:"reservado"`
	EstoqueDisponivel int     `json:"estoque_disponivel" db:"estoque_disponivel"`
}

// StockSearchRequest represents request to search stock
type StockSearchRequest struct {
	SKU string `json:"sku" binding:"required"`
}

// StockQueryRequest represents stock query request
type StockQueryRequest struct {
	Brand string `json:"brand" binding:"required"`
	SKU   string `json:"sku" binding:"required"`
}

// IntegrationLog represents integration execution log
type IntegrationLog struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	ProcessType string    `json:"process_type" db:"process_type"`
	Status      string    `json:"status" db:"status"`
	Data        string    `json:"data" db:"data"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ExecuteIntegrationRequest represents integration execution request
type ExecuteIntegrationRequest struct {
	Conta       string `json:"conta" binding:"required"`
	Marketplace string `json:"marketplace" binding:"required"`
	NumPedido   string `json:"num_pedido" binding:"required"`
}

// ImportLog represents XML import log
type ImportLog struct {
	ID               string     `json:"id" db:"id"`
	UserID           string     `json:"user_id" db:"user_id"`
	FileName         string     `json:"file_name" db:"file_name"`
	Status           string     `json:"status" db:"status"`
	ProcessedRecords int        `json:"processed_records" db:"processed_records"`
	TotalRecords     int        `json:"total_records" db:"total_records"`
	ErrorMessage     string     `json:"error_message" db:"error_message"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	CompletedAt      *time.Time `json:"completed_at" db:"completed_at"`
}

// ImportXMLRequest represents XML import request
type ImportXMLRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// ImportXMLResponse represents XML import response
type ImportXMLResponse struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// ImportStatus represents import status
type ImportStatus struct {
	ID               string     `json:"id"`
	Filename         string     `json:"filename"`
	Status           string     `json:"status"`
	Progress         int        `json:"progress"`
	Message          string     `json:"message"`
	CreatedAt        time.Time  `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	Error            string     `json:"error,omitempty"`
	RecordsProcessed int        `json:"records_processed,omitempty"`
	RecordsTotal     int        `json:"records_total,omitempty"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// XMLIntegrationLog representa um log do processo de integração XML
type XMLIntegrationLog struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Step      string `json:"step"`
	Message   string `json:"message"`
}

// XMLIntegrationResponse representa a resposta específica para integração XML
type XMLIntegrationResponse struct {
	TotalProcessed int                      `json:"total_processed"`
	SuccessCount   int                      `json:"success_count"`
	ErrorCount     int                      `json:"error_count"`
	Results        []map[string]interface{} `json:"results"`
	Logs           []XMLIntegrationLog      `json:"logs"`
}
