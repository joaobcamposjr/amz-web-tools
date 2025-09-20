package services

import (
	"database/sql"
	"fmt"
	"time"

	"amz-web-tools/backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a password with its hash
func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates a JWT token for a user
func (s *AuthService) GenerateJWT(userID, role, secret string, expireHours int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * time.Duration(expireHours)).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(req *models.RegisterRequest) (*models.User, error) {
	// Check if user already exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = @p1", req.Email).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, sql.ErrNoRows // User already exists
	}

	// Hash password
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Insert user
	query := `
		INSERT INTO users (email, password_hash, name, department, role)
		OUTPUT INSERTED.id, INSERTED.email, INSERTED.name, INSERTED.department, INSERTED.role, INSERTED.created_at, INSERTED.updated_at
		VALUES (@p1, @p2, @p3, @p4, 'user')`

	var user models.User
	err = s.db.QueryRow(query, req.Email, hashedPassword, req.Name, req.Department).Scan(
		&user.ID, &user.Email, &user.Name, &user.Department, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// LoginUser authenticates a user and returns user info
func (s *AuthService) LoginUser(req *models.LoginRequest) (*models.User, error) {
	query := `
		SELECT CAST(id AS NVARCHAR(36)), email, password_hash, name, department, role, is_first_login, password_changed_at, created_at, updated_at
		FROM users WHERE email = @p1`

	var user models.User
	err := s.db.QueryRow(query, req.Email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Department, &user.Role,
		&user.IsFirstLogin, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err // User not found
		}
		return nil, err
	}

	// Check password
	if !s.CheckPassword(req.Password, user.PasswordHash) {
		return nil, sql.ErrNoRows // Invalid password
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	query := `
		SELECT CAST(id AS NVARCHAR(36)), email, name, department, role, is_first_login, password_changed_at, created_at, updated_at
		FROM users WHERE id = @p1`

	var user models.User
	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Department, &user.Role,
		&user.IsFirstLogin, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUserProfile updates user profile information
func (s *AuthService) UpdateUserProfile(userID string, req *models.UpdateProfileRequest) (*models.User, error) {
	query := `
		UPDATE users 
		SET name = @p1, department = @p2, updated_at = GETDATE()
		OUTPUT INSERTED.id, INSERTED.email, INSERTED.name, INSERTED.department, INSERTED.role, INSERTED.created_at, INSERTED.updated_at
		WHERE id = @p3`

	var user models.User
	err := s.db.QueryRow(query, req.Name, req.Department, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Department, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUserPassword updates user password
func (s *AuthService) UpdateUserPassword(userID, currentPassword, newPassword string) error {
	// First, verify current password
	var currentHash string
	err := s.db.QueryRow("SELECT password_hash FROM users WHERE id = @p1", userID).Scan(&currentHash)
	if err != nil {
		return err
	}

	if !s.CheckPassword(currentPassword, currentHash) {
		return sql.ErrNoRows // Invalid current password
	}

	// Hash new password
	newHash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	_, err = s.db.Exec("UPDATE users SET password_hash = @p1, updated_at = GETDATE() WHERE id = @p2", newHash, userID)
	return err
}

// CreateUser creates a new user (Admin only)
func (s *AuthService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = @p1", req.Email).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("user already exists")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Insert user
	query := `
		INSERT INTO users (email, password_hash, name, department, role, is_first_login, password_changed_at)
		OUTPUT INSERTED.id, INSERTED.email, INSERTED.name, INSERTED.department, INSERTED.role, INSERTED.is_first_login, INSERTED.password_changed_at, INSERTED.created_at, INSERTED.updated_at
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)`

	var user models.User
	err = s.db.QueryRow(query, req.Email, hashedPassword, req.Name, req.Department, req.Role, true, time.Now()).Scan(
		&user.ID, &user.Email, &user.Name, &user.Department, &user.Role, &user.IsFirstLogin, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAllUsers retrieves all users (Admin only)
func (s *AuthService) GetAllUsers() ([]models.User, error) {
	query := `
		SELECT id, email, name, department, role, is_first_login, password_changed_at, created_at, updated_at
		FROM users ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.Department, &user.Role,
			&user.IsFirstLogin, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates a user (Admin only)
func (s *AuthService) UpdateUser(req *models.UpdateUserRequest) (*models.User, error) {
	query := `
		UPDATE users 
		SET name = @p1, department = @p2, role = @p3, updated_at = GETDATE()
		OUTPUT INSERTED.id, INSERTED.email, INSERTED.name, INSERTED.department, INSERTED.role, INSERTED.is_first_login, INSERTED.password_changed_at, INSERTED.created_at, INSERTED.updated_at
		WHERE id = @p4`

	var user models.User
	err := s.db.QueryRow(query, req.Name, req.Department, req.Role, req.ID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Department, &user.Role,
		&user.IsFirstLogin, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ResetUserPassword resets a user's password (Admin only)
func (s *AuthService) ResetUserPassword(req *models.ResetPasswordRequest) error {
	// Hash new password
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// Update password and reset first login flag
	_, err = s.db.Exec(`
		UPDATE users 
		SET password_hash = @p1, is_first_login = 1, password_changed_at = GETDATE(), updated_at = GETDATE() 
		WHERE id = @p2`,
		hashedPassword, req.UserID)

	return err
}

// ChangePasswordFirstLogin changes password on first login
func (s *AuthService) ChangePasswordFirstLogin(userID, newPassword string) error {
	// Hash new password
	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password and set first login to false
	_, err = s.db.Exec(`
		UPDATE users 
		SET password_hash = @p1, is_first_login = 0, password_changed_at = GETDATE(), updated_at = GETDATE() 
		WHERE id = @p2`,
		hashedPassword, userID)

	return err
}

// DeleteUser deletes a user (Admin only)
func (s *AuthService) DeleteUser(userID string) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = @p1", userID)
	return err
}
