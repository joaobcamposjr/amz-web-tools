-- Script para criar usuário admin
-- Execute este script no SQL Server Management Studio ou via sqlcmd

-- Verificar se o usuário admin já existe
IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@amztools.com')
BEGIN
    -- Inserir usuário admin
    INSERT INTO users (email, password_hash, name, department, role, is_first_login, password_changed_at)
    VALUES (
        'admin@amztools.com',
        '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password
        'Administrador',
        'TI',
        'admin',
        0, -- não é primeiro login
        GETDATE()
    )
    
    PRINT 'Usuário admin criado com sucesso!'
    PRINT 'Email: admin@amztools.com'
    PRINT 'Senha: password'
END
ELSE
BEGIN
    PRINT 'Usuário admin já existe!'
END



