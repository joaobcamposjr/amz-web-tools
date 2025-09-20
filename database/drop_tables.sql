-- AMZ Web Tools Portal - Drop Tables Script
-- SQL Server Database Cleanup Script

USE [portal];
GO

-- =============================================
-- Drop Foreign Key Constraints First
-- =============================================

IF EXISTS (SELECT * FROM sys.foreign_keys WHERE name = 'FK_user_sessions_users')
BEGIN
    ALTER TABLE [dbo].[user_sessions] DROP CONSTRAINT [FK_user_sessions_users];
    PRINT 'Dropped FK_user_sessions_users constraint';
END

IF EXISTS (SELECT * FROM sys.foreign_keys WHERE name = 'FK_integration_logs_users')
BEGIN
    ALTER TABLE [dbo].[integration_logs] DROP CONSTRAINT [FK_integration_logs_users];
    PRINT 'Dropped FK_integration_logs_users constraint';
END

IF EXISTS (SELECT * FROM sys.foreign_keys WHERE name = 'FK_import_logs_users')
BEGIN
    ALTER TABLE [dbo].[import_logs] DROP CONSTRAINT [FK_import_logs_users];
    PRINT 'Dropped FK_import_logs_users constraint';
END

-- =============================================
-- Drop Indexes
-- =============================================

IF EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_users_email')
BEGIN
    DROP INDEX [IX_users_email] ON [dbo].[users];
    PRINT 'Dropped IX_users_email index';
END

IF EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_user_sessions_token')
BEGIN
    DROP INDEX [IX_user_sessions_token] ON [dbo].[user_sessions];
    PRINT 'Dropped IX_user_sessions_token index';
END

IF EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_plate_cache_plate')
BEGIN
    DROP INDEX [IX_plate_cache_plate] ON [dbo].[plate_cache];
    PRINT 'Dropped IX_plate_cache_plate index';
END

IF EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_depara_products_product_code')
BEGIN
    DROP INDEX [IX_depara_products_product_code] ON [dbo].[depara_products];
    PRINT 'Dropped IX_depara_products_product_code index';
END

IF EXISTS (SELECT * FROM sys.indexes WHERE name = 'IX_stock_items_brand_sku')
BEGIN
    DROP INDEX [IX_stock_items_brand_sku] ON [dbo].[stock_items];
    PRINT 'Dropped IX_stock_items_brand_sku index';
END

-- =============================================
-- Drop Tables
-- =============================================

IF EXISTS (SELECT * FROM sysobjects WHERE name='import_logs' AND xtype='U')
BEGIN
    DROP TABLE [dbo].[import_logs];
    PRINT 'Dropped table [import_logs]';
END

IF EXISTS (SELECT * FROM sysobjects WHERE name='integration_logs' AND xtype='U')
BEGIN
    DROP TABLE [dbo].[integration_logs];
    PRINT 'Dropped table [integration_logs]';
END

IF EXISTS (SELECT * FROM sysobjects WHERE name='stock_items' AND xtype='U')
BEGIN
    DROP TABLE [dbo].[stock_items];
    PRINT 'Dropped table [stock_items]';
END

IF EXISTS (SELECT * FROM sysobjects WHERE name='depara_products' AND xtype='U')
BEGIN
    DROP TABLE [dbo].[depara_products];
    PRINT 'Dropped table [depara_products]';
END

IF EXISTS (SELECT * FROM sysobjects WHERE name='plate_cache' AND xtype='U')
BEGIN
    DROP TABLE [dbo].[plate_cache];
    PRINT 'Dropped table [plate_cache]';
END

IF EXISTS (SELECT * FROM sysobjects WHERE name='user_sessions' AND xtype='U')
BEGIN
    DROP TABLE [dbo].[user_sessions];
    PRINT 'Dropped table [user_sessions]';
END

IF EXISTS (SELECT * FROM sysobjects WHERE name='users' AND xtype='U')
BEGIN
    DROP TABLE [dbo].[users];
    PRINT 'Dropped table [users]';
END

PRINT '=============================================';
PRINT 'All tables dropped successfully!';
PRINT '=============================================';
GO

