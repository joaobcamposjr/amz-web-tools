package database

import (
	"database/sql"
	"fmt"
	"log"

	"amz-web-tools/backend/internal/config"

	_ "github.com/microsoft/go-mssqldb"
)

func Initialize(cfg *config.Config) (*sql.DB, error) {
	// SQL Server connection string
	connectionString := fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s;encrypt=disable;trustServerCertificate=true;connection timeout=30",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	log.Printf("Attempting to connect to database: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	// Check if tables exist and create them if they don't
	tables := []string{
		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='users' AND xtype='U')
		CREATE TABLE users (
			id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
			email NVARCHAR(255) UNIQUE NOT NULL,
			password_hash NVARCHAR(255) NOT NULL,
			name NVARCHAR(255) NOT NULL,
			department NVARCHAR(255),
			role NVARCHAR(50) DEFAULT 'user',
			is_first_login BIT DEFAULT 1,
			password_changed_at DATETIME2 DEFAULT GETDATE(),
			created_at DATETIME2 DEFAULT GETDATE(),
			updated_at DATETIME2 DEFAULT GETDATE()
		)`,

		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='user_sessions' AND xtype='U')
		CREATE TABLE user_sessions (
			id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
			user_id UNIQUEIDENTIFIER NOT NULL,
			token NVARCHAR(500) NOT NULL,
			expires_at DATETIME2 NOT NULL,
			created_at DATETIME2 DEFAULT GETDATE(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='plate_cache' AND xtype='U')
		CREATE TABLE plate_cache (
			id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
			plate NVARCHAR(10) UNIQUE NOT NULL,
			data NVARCHAR(MAX) NOT NULL,
			created_at DATETIME2 DEFAULT GETDATE(),
			expires_at DATETIME2 NOT NULL
		)`,

		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='depara_products' AND xtype='U')
		CREATE TABLE depara_products (
			id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
			product_code NVARCHAR(100) NOT NULL,
			name NVARCHAR(255) NOT NULL,
			description NVARCHAR(MAX),
			category NVARCHAR(100),
			created_at DATETIME2 DEFAULT GETDATE(),
			updated_at DATETIME2 DEFAULT GETDATE()
		)`,

		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='stock_items' AND xtype='U')
		CREATE TABLE stock_items (
			id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
			brand NVARCHAR(100) NOT NULL,
			sku NVARCHAR(100) NOT NULL,
			quantity INT NOT NULL,
			location NVARCHAR(255),
			updated_at DATETIME2 DEFAULT GETDATE()
		)`,

		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='integration_logs' AND xtype='U')
		CREATE TABLE integration_logs (
			id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
			user_id UNIQUEIDENTIFIER NOT NULL,
			process_type NVARCHAR(100) NOT NULL,
			status NVARCHAR(50) NOT NULL,
			data NVARCHAR(MAX),
			created_at DATETIME2 DEFAULT GETDATE(),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='import_logs' AND xtype='U')
		CREATE TABLE import_logs (
			id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
			user_id UNIQUEIDENTIFIER NOT NULL,
			file_name NVARCHAR(255) NOT NULL,
			status NVARCHAR(50) NOT NULL,
			processed_records INT DEFAULT 0,
			total_records INT DEFAULT 0,
			error_message NVARCHAR(MAX),
			created_at DATETIME2 DEFAULT GETDATE(),
			completed_at DATETIME2,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='car_plate_history' AND xtype='U')
		CREATE TABLE car_plate_history (
			id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
			plate NVARCHAR(10) NOT NULL,
			response_data NVARCHAR(MAX),
			status NVARCHAR(20) NOT NULL,
			error_message NVARCHAR(500),
			created_at DATETIME2 DEFAULT GETDATE(),
			user_id UNIQUEIDENTIFIER,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}

	for i, query := range tables {
		if _, err := db.Exec(query); err != nil {
			log.Printf("Warning: Failed to create table %d: %v", i+1, err)
			// Continue with other tables even if one fails
		} else {
			log.Printf("Table %d created successfully", i+1)
		}
	}

	// Add new columns to existing users table if they don't exist
	migrationQueries := []string{
		`IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID('users') AND name = 'is_first_login')
		ALTER TABLE users ADD is_first_login BIT DEFAULT 1`,

		`IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID('users') AND name = 'password_changed_at')
		ALTER TABLE users ADD password_changed_at DATETIME2 DEFAULT GETDATE()`,
	}

	for i, query := range migrationQueries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("Warning: Failed to execute migration %d: %v", i+1, err)
		} else {
			log.Printf("Migration %d executed successfully", i+1)
		}
	}

	return nil
}
