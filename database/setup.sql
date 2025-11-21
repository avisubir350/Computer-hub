-- PC Repair Hub Database Setup Script
-- Updated for Normalized Schema: CUSTOMERS, DEVICE_DETAILS, TICKETS, ORDER_LINE_ITEMS

-- Create database
CREATE DATABASE IF NOT EXISTS pcrepairhub 
CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;

USE pcrepairhub;

-- 1. USERS table (Staff/Engineer)
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

-- 2. CUSTOMERS table (Client Information)
CREATE TABLE IF NOT EXISTS customers (
    customer_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20) NOT NULL,
    address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    zip VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_customer_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. DEVICE_DETAILS table (Equipment Information)
CREATE TABLE IF NOT EXISTS device_details (
    device_id VARCHAR(50) PRIMARY KEY,
    customer_id VARCHAR(50) NOT NULL,
    type VARCHAR(255) NOT NULL,
    brand VARCHAR(255) NOT NULL,
    model VARCHAR(255),
    serial_no VARCHAR(255),
    password VARCHAR(255),
    accessories_received TEXT,
    under_warranty BOOLEAN NOT NULL DEFAULT FALSE,
    warranty_no VARCHAR(255),
    warranty_exp_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. TICKETS table (Core Job and Status - Replaces old 'orders' table)
CREATE TABLE IF NOT EXISTS tickets (
    ticket_id VARCHAR(50) PRIMARY KEY,
    customer_id VARCHAR(50) NOT NULL,
    device_id VARCHAR(50) NOT NULL,
    assigned_engineer_id VARCHAR(50),
    ticket_type ENUM('Diagnostics Call', 'Service Call') NOT NULL,
    issue_description TEXT NOT NULL,
    data_backup_consent ENUM('backed_up', 'no_backup_no_service', 'request_backup') NOT NULL,
    expected_delivery_date DATE,
    status ENUM('New Order', 'Diagnostics', 'In Progress', 'Ready for Delivery', 'Collected') DEFAULT 'New Order',
    total_cost DECIMAL(10,2) NOT NULL,
    created_by VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_updated_by VARCHAR(50),
    
    INDEX idx_ticket_status (status),
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id) ON DELETE CASCADE,
    FOREIGN KEY (device_id) REFERENCES device_details(device_id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_engineer_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (last_updated_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. ORDER_LINE_ITEMS table (Billing Items)
CREATE TABLE IF NOT EXISTS order_line_items (
    item_id VARCHAR(50) PRIMARY KEY,
    ticket_id VARCHAR(50) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    rate DECIMAL(10,2) NOT NULL,
    discount_percent DECIMAL(5,2) DEFAULT 0.00,
    final_price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ticket_id) REFERENCES tickets(ticket_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- Insert sample admin user (unchanged)
INSERT IGNORE INTO users (id, full_name, email, phone, password, role, created_at, updated_at) 
VALUES ('ADMIN-001', 'System Administrator', 'admin@pchub.com', '+91 98765 43210', 'admin123', 'Administrator', NOW(), NOW());

-- Insert sample regular user
INSERT IGNORE INTO users (id, full_name, email, phone, password, role, created_at, updated_at) 
VALUES ('USER-001', 'John Doe', 'john@example.com', '+91 87654 32109', 'password123', 'User', NOW(), NOW());


-- Sample data insertion for the new schema is complex and skipped for brevity.
-- The application logic now handles inserting data into all these tables in a single transaction.

-- Show tables and sample data
SHOW TABLES;
SELECT 'Users Table:' as Info;
SELECT id, full_name, email, phone, role, created_at FROM users;
SELECT 'Customers Table:' as Info;
SELECT customer_id, name, phone, email FROM customers;
SELECT 'Tickets Table:' as Info;
SELECT ticket_id, status, total_cost, created_at FROM tickets;