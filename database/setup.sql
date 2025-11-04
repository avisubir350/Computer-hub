-- PC Repair Hub Database Setup Script
-- Run this script to create the database and initial data

-- Create database
CREATE DATABASE IF NOT EXISTS pcrepairhub 
CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;

USE pcrepairhub;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(50) PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'User',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Orders/Tickets table
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(50) PRIMARY KEY,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(20) NOT NULL,
    device_type VARCHAR(255) NOT NULL,
    device_model VARCHAR(255),
    services JSON NOT NULL,
    issue_description TEXT,
    status ENUM('New Order', 'In Progress', 'Ready for Delivery', 'Collected') DEFAULT 'New Order',
    total_cost DECIMAL(10,2) NOT NULL,
    created_by VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_updated_by VARCHAR(50),
    INDEX idx_status (status),
    INDEX idx_customer_email (customer_email),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (last_updated_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample admin user (password: admin123 - should be hashed in production)
INSERT IGNORE INTO users (id, full_name, email, phone, password, role, created_at, updated_at) 
VALUES ('ADMIN-001', 'System Administrator', 'admin@pchub.com', '+91 98765 43210', 'admin123', 'Administrator', NOW(), NOW());

-- Insert sample regular user
INSERT IGNORE INTO users (id, full_name, email, phone, password, role, created_at, updated_at) 
VALUES ('USER-001', 'John Doe', 'john@example.com', '+91 87654 32109', 'password123', 'User', NOW(), NOW());

-- Insert sample orders
INSERT IGNORE INTO orders (id, customer_name, customer_email, customer_phone, device_type, device_model, services, issue_description, status, total_cost, created_by, created_at, updated_at, last_updated_by) 
VALUES 
('ORD-001', 'Rajesh Kumar', 'rajesh@example.com', '+91 98765 43210', 'Laptop', 'Dell XPS 13', '["System Diagnostic & Quote"]', 'Laptop not booting after Windows update', 'New Order', 999.00, 'ADMIN-001', NOW(), NOW(), 'ADMIN-001'),
('ORD-002', 'Priya Sharma', 'priya@example.com', '+91 87654 32109', 'Desktop', 'HP EliteDesk 800', '["Virus & Malware Removal", "Operating System Fresh Install"]', 'Computer running very slow, suspected virus infection', 'In Progress', 3498.00, 'ADMIN-001', DATE_SUB(NOW(), INTERVAL 1 DAY), NOW(), 'ADMIN-001'),
('ORD-003', 'Amit Patel', 'amit@example.com', '+91 76543 21098', 'Printer', 'HP LaserJet Pro M404n', '["Printer Repair & Maintenance"]', 'Printer not printing, paper jam error', 'Ready for Delivery', 799.00, 'ADMIN-001', DATE_SUB(NOW(), INTERVAL 2 DAY), NOW(), 'ADMIN-001');

-- Create indexes for better performance
CREATE INDEX idx_orders_customer_name ON orders(customer_name);
CREATE INDEX idx_orders_device_type ON orders(device_type);
CREATE INDEX idx_users_full_name ON users(full_name);

-- Show tables and sample data
SHOW TABLES;
SELECT 'Users Table:' as Info;
SELECT id, full_name, email, phone, role, created_at FROM users;
SELECT 'Orders Table:' as Info;
SELECT id, customer_name, device_type, status, total_cost, created_at FROM orders;