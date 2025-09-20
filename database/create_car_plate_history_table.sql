-- Create car plate consultation history table
USE integration;

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='car_plate_history' AND xtype='U')
BEGIN
    CREATE TABLE car_plate_history (
        id UNIQUEIDENTIFIER DEFAULT NEWID() PRIMARY KEY,
        plate VARCHAR(10) NOT NULL,
        response_data NVARCHAR(MAX),
        status VARCHAR(20) NOT NULL, -- 'success', 'error', 'not_found'
        error_message NVARCHAR(500),
        created_at DATETIME2 DEFAULT GETDATE(),
        user_id UNIQUEIDENTIFIER,
        FOREIGN KEY (user_id) REFERENCES users(id)
    );
    
    PRINT 'Table car_plate_history created successfully';
END
ELSE
BEGIN
    PRINT 'Table car_plate_history already exists';
END







