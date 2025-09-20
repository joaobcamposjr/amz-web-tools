-- Create DePara table for integration.amazonas_psa.mercadolivre_base
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='depara_mercadolivre' AND xtype='U')
CREATE TABLE depara_mercadolivre (
    id NVARCHAR(50) PRIMARY KEY,
    mlbu NVARCHAR(50),
    type NVARCHAR(100),
    sku NVARCHAR(255),
    company NVARCHAR(255),
    permalink NVARCHAR(500),
    ship_cost_slow DECIMAL(10,2),
    ship_cost_standard DECIMAL(10,2),
    ship_cost_nextday DECIMAL(10,2),
    pictures NVARCHAR(MAX), -- JSON array of image URLs
    updated_at DATETIME2 DEFAULT GETDATE(),
    created_at DATETIME2 DEFAULT GETDATE()
);

-- Insert sample data for testing
INSERT INTO depara_mercadolivre (id, mlbu, type, sku, company, permalink, ship_cost_slow, ship_cost_standard, ship_cost_nextday, pictures) VALUES
('MLB123456789', 'MLB123456789', 'product', 'SKU001', 'Amazonas', 'https://produto.mercadolivre.com.br/MLB123456789', 15.50, 25.00, 45.00, '["https://http2.mlstatic.com/D_NQ_NP_123456-MLB123456789_012021-O.jpg", "https://http2.mlstatic.com/D_NQ_NP_234567-MLB123456789_012021-O.jpg", "https://http2.mlstatic.com/D_NQ_NP_345678-MLB123456789_012021-O.jpg"]'),
('MLB987654321', 'MLB987654321', 'product', 'SKU002', 'PSA', 'https://produto.mercadolivre.com.br/MLB987654321', 12.00, 20.00, 35.00, '["https://http2.mlstatic.com/D_NQ_NP_987654-MLB987654321_022021-O.jpg", "https://http2.mlstatic.com/D_NQ_NP_876543-MLB987654321_022021-O.jpg"]'),
('MLB555666777', 'MLB555666777', 'product', 'SKU003', 'MercadoLivre', 'https://produto.mercadolivre.com.br/MLB555666777', 18.75, 30.00, 50.00, '["https://http2.mlstatic.com/D_NQ_NP_555666-MLB555666777_032021-O.jpg"]');

-- Create table for available integration tables
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='integration_tables' AND xtype='U')
CREATE TABLE integration_tables (
    id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
    table_name NVARCHAR(255) NOT NULL,
    display_name NVARCHAR(255) NOT NULL,
    is_active BIT DEFAULT 1,
    created_at DATETIME2 DEFAULT GETDATE()
);

-- Insert available tables
INSERT INTO integration_tables (table_name, display_name) VALUES
('integration.amazonas_psa.mercadolivre_base', 'MercadoLivre'),
('integration.amazonas_psa.amazonas_base', 'Amazonas'),
('integration.amazonas_psa.psa_base', 'PSA');







