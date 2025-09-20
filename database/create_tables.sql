-- AMZ Web Tools Portal - Database Schema
-- SQL Server Database Creation Script

USE [portal];
GO

-- =============================================
-- Users Table
-- =============================================
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='users' AND xtype='U')
BEGIN
    CREATE TABLE [dbo].[users] (
        [id] UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
        [email] NVARCHAR(255) UNIQUE NOT NULL,
        [password_hash] NVARCHAR(255) NOT NULL,
        [name] NVARCHAR(255) NOT NULL,
        [department] NVARCHAR(255),
        [role] NVARCHAR(50) DEFAULT 'user',
        [created_at] DATETIME2 DEFAULT GETDATE(),
        [updated_at] DATETIME2 DEFAULT GETDATE()
    );
    
    PRINT 'Table [users] created successfully';
END
ELSE
BEGIN
    PRINT 'Table [users] already exists';
END
GO

-- =============================================
-- User Sessions Table
-- =============================================
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='user_sessions' AND xtype='U')
BEGIN
    CREATE TABLE [dbo].[user_sessions] (
        [id] UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
        [user_id] UNIQUEIDENTIFIER NOT NULL,
        [token] NVARCHAR(500) NOT NULL,
        [expires_at] DATETIME2 NOT NULL,
        [created_at] DATETIME2 DEFAULT GETDATE(),
        CONSTRAINT [FK_user_sessions_users] FOREIGN KEY ([user_id]) 
            REFERENCES [dbo].[users]([id]) ON DELETE CASCADE
    );
    
    PRINT 'Table [user_sessions] created successfully';
END
ELSE
BEGIN
    PRINT 'Table [user_sessions] already exists';
END
GO

-- =============================================
-- Plate Cache Table
-- =============================================
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='plate_cache' AND xtype='U')
BEGIN
    CREATE TABLE [dbo].[plate_cache] (
        [id] UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
        [plate] NVARCHAR(10) UNIQUE NOT NULL,
        [data] NVARCHAR(MAX) NOT NULL,
        [created_at] DATETIME2 DEFAULT GETDATE(),
        [expires_at] DATETIME2 NOT NULL
    );
    
    PRINT 'Table [plate_cache] created successfully';
END
ELSE
BEGIN
    PRINT 'Table [plate_cache] already exists';
END
GO

-- =============================================
-- DePara Products Table
-- =============================================
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='depara_products' AND xtype='U')
BEGIN
    CREATE TABLE [dbo].[depara_products] (
        [id] UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
        [product_code] NVARCHAR(100) NOT NULL,
        [name] NVARCHAR(255) NOT NULL,
        [description] NVARCHAR(MAX),
        [category] NVARCHAR(100),
        [created_at] DATETIME2 DEFAULT GETDATE(),
        [updated_at] DATETIME2 DEFAULT GETDATE()
    );
    
    PRINT 'Table [depara_products] created successfully';
END
ELSE
BEGIN
    PRINT 'Table [depara_products] already exists';
END
GO

-- =============================================
-- Stock Items Table
-- =============================================
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='stock_items' AND xtype='U')
BEGIN
    CREATE TABLE [dbo].[stock_items] (
        [id] UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
        [brand] NVARCHAR(100) NOT NULL,
        [sku] NVARCHAR(100) NOT NULL,
        [quantity] INT NOT NULL,
        [location] NVARCHAR(255),
        [updated_at] DATETIME2 DEFAULT GETDATE()
    );
    
    PRINT 'Table [stock_items] created successfully';
END
ELSE
BEGIN
    PRINT 'Table [stock_items] already exists';
END
GO

-- =============================================
-- Integration Logs Table
-- =============================================
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='integration_logs' AND xtype='U')
BEGIN
    CREATE TABLE [dbo].[integration_logs] (
        [id] UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
        [user_id] UNIQUEIDENTIFIER NOT NULL,
        [process_type] NVARCHAR(100) NOT NULL,
        [status] NVARCHAR(50) NOT NULL,
        [data] NVARCHAR(MAX),
        [created_at] DATETIME2 DEFAULT GETDATE(),
        CONSTRAINT [FK_integration_logs_users] FOREIGN KEY ([user_id]) 
            REFERENCES [dbo].[users]([id])
    );
    
    PRINT 'Table [integration_logs] created successfully';
END
ELSE
BEGIN
    PRINT 'Table [integration_logs] already exists';
END
GO

-- =============================================
-- Import Logs Table
-- =============================================
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='import_logs' AND xtype='U')
BEGIN
    CREATE TABLE [dbo].[import_logs] (
        [id] UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
        [user_id] UNIQUEIDENTIFIER NOT NULL,
        [file_name] NVARCHAR(255) NOT NULL,
        [status] NVARCHAR(50) NOT NULL,
        [processed_records] INT DEFAULT 0,
        [total_records] INT DEFAULT 0,
        [error_message] NVARCHAR(MAX),
        [created_at] DATETIME2 DEFAULT GETDATE(),
        [completed_at] DATETIME2,
        CONSTRAINT [FK_import_logs_users] FOREIGN KEY ([user_id]) 
            REFERENCES [dbo].[users]([id])
    );
    
    PRINT 'Table [import_logs] created successfully';
END
ELSE
BEGIN
    PRINT 'Table [import_logs] already exists';
END
GO

-- =============================================
-- Create Indexes for Better Performance
-- =============================================

-- Index for users email
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_users_email')
BEGIN
    CREATE INDEX [IX_users_email] ON [dbo].[users]([email]);
    PRINT 'Index [IX_users_email] created successfully';
END

-- Index for user_sessions token
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_user_sessions_token')
BEGIN
    CREATE INDEX [IX_user_sessions_token] ON [dbo].[user_sessions]([token]);
    PRINT 'Index [IX_user_sessions_token] created successfully';
END

-- Index for plate_cache plate
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_plate_cache_plate')
BEGIN
    CREATE INDEX [IX_plate_cache_plate] ON [dbo].[plate_cache]([plate]);
    PRINT 'Index [IX_plate_cache_plate] created successfully';
END

-- Index for depara_products product_code
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_depara_products_product_code')
BEGIN
    CREATE INDEX [IX_depara_products_product_code] ON [dbo].[depara_products]([product_code]);
    PRINT 'Index [IX_depara_products_product_code] created successfully';
END

-- Index for stock_items brand and sku
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_stock_items_brand_sku')
BEGIN
    CREATE INDEX [IX_stock_items_brand_sku] ON [dbo].[stock_items]([brand], [sku]);
    PRINT 'Index [IX_stock_items_brand_sku] created successfully';
END

-- =============================================
-- Insert Default Admin User
-- =============================================
IF NOT EXISTS (SELECT * FROM [dbo].[users] WHERE email = 'admin@amztools.com')
BEGIN
    INSERT INTO [dbo].[users] ([email], [password_hash], [name], [department], [role])
    VALUES (
        'admin@amztools.com',
        '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: password
        'Administrador',
        'TI',
        'admin'
    );
    
    PRINT 'Default admin user created successfully';
    PRINT 'Email: admin@amztools.com';
    PRINT 'Password: password';
END
ELSE
BEGIN
    PRINT 'Default admin user already exists';
END
GO

-- =============================================
-- Insert Sample Data for Testing
-- =============================================

-- Sample DePara Products
IF NOT EXISTS (SELECT * FROM [dbo].[depara_products] WHERE product_code = 'SAMPLE001')
BEGIN
    INSERT INTO [dbo].[depara_products] ([product_code], [name], [description], [category])
    VALUES 
        ('SAMPLE001', 'Filtro de Óleo', 'Filtro de óleo para motores 1.0', 'Filtros'),
        ('SAMPLE002', 'Pastilha de Freio', 'Pastilha de freio dianteira', 'Freios'),
        ('SAMPLE003', 'Vela de Ignição', 'Vela de ignição para motor flex', 'Ignição');
    
    PRINT 'Sample DePara products created successfully';
END

-- Sample Stock Items
IF NOT EXISTS (SELECT * FROM [dbo].[stock_items] WHERE sku = 'FILTRO001')
BEGIN
    INSERT INTO [dbo].[stock_items] ([brand], [sku], [quantity], [location])
    VALUES 
        ('MANN', 'FILTRO001', 50, 'Estoque A'),
        ('BOSCH', 'PASTILHA001', 25, 'Estoque B'),
        ('NGK', 'VELA001', 100, 'Estoque C');
    
    PRINT 'Sample stock items created successfully';
END

PRINT '=============================================';
PRINT 'Database setup completed successfully!';
PRINT '=============================================';
PRINT 'Default admin credentials:';
PRINT 'Email: admin@amztools.com';
PRINT 'Password: password';
PRINT '=============================================';
GO

