-- Create audit log table in portal database
USE portal;

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='audit_logs' AND xtype='U')
BEGIN
    CREATE TABLE audit_logs (
        id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
        table_name NVARCHAR(100) NOT NULL,
        record_id NVARCHAR(100) NOT NULL,
        operation NVARCHAR(20) NOT NULL, -- 'INSERT', 'UPDATE', 'DELETE'
        user_id UNIQUEIDENTIFIER,
        user_email NVARCHAR(255),
        user_name NVARCHAR(255),
        old_values NVARCHAR(MAX), -- JSON com valores anteriores
        new_values NVARCHAR(MAX), -- JSON com valores novos
        changed_fields NVARCHAR(MAX), -- JSON com campos alterados
        ip_address NVARCHAR(45),
        user_agent NVARCHAR(500),
        created_at DATETIME2 DEFAULT GETDATE(),
        rollback_data NVARCHAR(MAX), -- Dados para rollback
        FOREIGN KEY (user_id) REFERENCES integration.dbo.users(id)
    );
    
    -- Create index for better performance
    CREATE INDEX IX_audit_logs_table_record ON audit_logs(table_name, record_id);
    CREATE INDEX IX_audit_logs_user_id ON audit_logs(user_id);
    CREATE INDEX IX_audit_logs_created_at ON audit_logs(created_at);
    
    PRINT 'Table audit_logs created successfully in portal database';
END
ELSE
BEGIN
    PRINT 'Table audit_logs already exists in portal database';
END







