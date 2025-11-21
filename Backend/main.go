package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/cors"
)

// --- Domain Models (structs for Normalized Tables) ---

// Customer represents the structure of Client Information
type Customer struct {
	ID      string `json:"customer_id" db:"customer_id"`
	Name    string `json:"name" db:"name"`
	Email   string `json:"email" db:"email"`
	Phone   string `json:"phone" db:"phone"`
	Address string `json:"address" db:"address"`
	City    string `json:"city" db:"city"`
	State   string `json:"state" db:"state"`
	Zip     string `json:"zip" db:"zip"`
}

// DeviceDetail represents the structure of Equipment Information
type DeviceDetail struct {
	ID              string    `json:"device_id" db:"device_id"`
	Type            string    `json:"device_type" db:"type"`
	Brand           string    `json:"device_brand" db:"brand"`
	Model           string    `json:"device_model" db:"model"`
	SerialNo        string    `json:"device_serial_no" db:"serial_no"`
	Password        string    `json:"device_password" db:"password"`
	Accessories     string    `json:"accessories_received" db:"accessories_received"` // String (can be JSON)
	UnderWarranty   bool      `json:"under_warranty" db:"under_warranty"`
	WarrantyNo      string    `json:"warranty_no" db:"warranty_no"`
	WarrantyExpDate time.Time `json:"warranty_exp_date" db:"warranty_exp_date"`
	CustomerID      string    `json:"customer_id" db:"customer_id"`
}

// LineItem represents a single service or part charged on a ticket
type LineItem struct {
	ID              string  `json:"item_id" db:"item_id"`
	TicketID        string  `json:"ticket_id" db:"ticket_id"`
	ServiceName     string  `json:"serviceName" db:"service_name"`
	Rate            float64 `json:"rate" db:"rate"`
	DiscountPercent float64 `json:"discountPercent" db:"discount_percent"`
	FinalPrice      float64 `json:"finalPrice" db:"final_price"`
}

// TicketInput is the aggregate structure for receiving a new ticket via API
type TicketInput struct {
	// Customer Fields
	CustomerName    string `json:"customerName"`
	CustomerEmail   string `json:"customerEmail"`
	CustomerPhone   string `json:"customerPhone"`
	CustomerAddress string `json:"customerAddress"`
	CustomerCity    string `json:"customerCity"`
	CustomerState   string `json:"customerState"`
	CustomerZip     string `json:"customerZip"`

	// Device Fields
	DeviceType          string `json:"deviceType"`
	DeviceBrand         string `json:"deviceBrand"`
	DeviceModelNo       string `json:"deviceModelNo"`
	DeviceSerialNo      string `json:"deviceSerialNo"`
	DevicePassword      string `json:"devicePassword"`
	AccessoriesReceived string `json:"accessoriesReceived"`

	// Ticket Core Fields
	TicketType           string `json:"ticketType"`
	AssignedEngineerID   string `json:"engineerId"`
	IssueDescription     string `json:"issueDescription"`
	DataBackup           string `json:"dataBackup"`
	UnderWarranty        bool   `json:"underWarranty"`
	WarrantyNo           string `json:"warrantyNo"`
	WarrantyExpDate      string `json:"warrantyExpDate"`
	ExpectedDeliveryDate string `json:"expectedDeliveryDate"`

	// Financials
	ServiceLineItems []LineItem `json:"serviceLineItems"`
	TotalCost        float64    `json:"totalCost"`
	CreatedBy        string     `json:"createdBy"`
}

// Ticket is a simplified structure for retrieving joined data (replaces old Order struct)
type Ticket struct {
	ID            string    `json:"id" db:"id"`
	CustomerName  string    `json:"customer_name"`
	CustomerPhone string    `json:"customer_phone"`
	DeviceType    string    `json:"device_type"`
	DeviceModel   string    `json:"device_model"`
	Status        string    `json:"status"`
	TotalCost     float64   `json:"total_cost"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DashboardMetrics holds the aggregated data for the operational dashboard.
type DashboardMetrics struct {
	TotalOpenOrders  int     `json:"total_open_orders"`
	ReadyForDelivery int     `json:"ready_for_delivery"`
	TotalRevenueYTD  float64 `json:"total_revenue_ytd"`
}

// OrderService handles ticket database operations across multiple tables
type OrderService struct {
	db *sql.DB
}

func NewOrderService(database *sql.DB) *OrderService {
	return &OrderService{db: database}
}

// Helper function to generate IDs
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// CreateTicket performs a multi-table transaction to insert a new ticket
func (os *OrderService) CreateTicket(input *TicketInput) error {
	// 1. Start Transaction
	tx, err := os.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Transaction panicked and rolled back: %v", r)
		} else if err != nil {
			tx.Rollback()
			log.Printf("Transaction failed and rolled back: %v", err)
		} else {
			err = tx.Commit()
			if err != nil {
				log.Printf("Transaction commit failed: %v", err)
			}
		}
	}()

	// Generate IDs
	customerID := generateID("CUST")
	deviceID := generateID("DEV")
	ticketID := generateID("TICKET")

	// Handle date parsing
	warrantyExpDate, _ := time.Parse("2006-01-02", input.WarrantyExpDate)
	expectedDeliveryDate, _ := time.Parse("2006-01-02", input.ExpectedDeliveryDate)

	// 2. Insert into CUSTOMERS
	customerQuery := `
        INSERT INTO customers (customer_id, name, email, phone, address, city, state, zip)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err = tx.Exec(customerQuery, customerID, input.CustomerName, input.CustomerEmail,
		input.CustomerPhone, input.CustomerAddress, input.CustomerCity,
		input.CustomerState, input.CustomerZip)
	if err != nil {
		return fmt.Errorf("failed to insert customer: %w", err)
	}

	// 3. Insert into DEVICE_DETAILS
	deviceQuery := `
        INSERT INTO device_details (device_id, customer_id, type, brand, model, serial_no, password, accessories_received, under_warranty, warranty_no, warranty_exp_date)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err = tx.Exec(deviceQuery, deviceID, customerID, input.DeviceType, input.DeviceBrand,
		input.DeviceModelNo, input.DeviceSerialNo, input.DevicePassword,
		input.AccessoriesReceived, input.UnderWarranty, input.WarrantyNo, warrantyExpDate)
	if err != nil {
		return fmt.Errorf("failed to insert device details: %w", err)
	}

	// 4. Insert into TICKETS
	ticketQuery := `
        INSERT INTO tickets (ticket_id, customer_id, device_id, assigned_engineer_id, ticket_type, 
                             issue_description, data_backup_consent, expected_delivery_date, status, 
                             total_cost, created_by, created_at, updated_at, last_updated_by)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), ?)
    `
	_, err = tx.Exec(ticketQuery, ticketID, customerID, deviceID, input.AssignedEngineerID, input.TicketType,
		input.IssueDescription, input.DataBackup, expectedDeliveryDate, "New Order",
		input.TotalCost, input.CreatedBy, input.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to insert ticket: %w", err)
	}

	// 5. Insert into ORDER_LINE_ITEMS
	lineItemQuery := `
        INSERT INTO order_line_items (item_id, ticket_id, service_name, rate, discount_percent, final_price)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	for i, item := range input.ServiceLineItems {
		itemID := fmt.Sprintf("%s-ITEM-%d", ticketID, i+1)
		_, err = tx.Exec(lineItemQuery, itemID, ticketID, item.ServiceName, item.Rate, item.DiscountPercent, item.FinalPrice)
		if err != nil {
			return fmt.Errorf("failed to insert line item %d: %w", i+1, err)
		}
	}

	return nil
}

// GetAllOrders retrieves all tickets (simplified join for dashboard display)
func (os *OrderService) GetAllOrders() ([]Ticket, error) {
	query := `
        SELECT 
			t.ticket_id, c.name, c.phone, d.type, d.model, 
			t.status, t.total_cost, t.created_by, t.created_at, t.updated_at
        FROM tickets t
        JOIN customers c ON t.customer_id = c.customer_id
        JOIN device_details d ON t.device_id = d.device_id
        ORDER BY t.created_at DESC
    `

	rows, err := os.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []Ticket
	for rows.Next() {
		var t Ticket
		var createdBySQL sql.NullString // Use NullString for nullable foreign keys

		err := rows.Scan(
			&t.ID, &t.CustomerName, &t.CustomerPhone, &t.DeviceType, &t.DeviceModel,
			&t.Status, &t.TotalCost, &createdBySQL, &t.CreatedAt, &t.UpdatedAt)

		if err != nil {
			log.Printf("Error scanning ticket row: %v", err)
			continue
		}
		t.CreatedBy = createdBySQL.String

		tickets = append(tickets, t)
	}

	return tickets, nil
}

// UpdateOrderStatus updates the status of a ticket
func (os *OrderService) UpdateOrderStatus(orderID, status, updatedBy string) error {
	query := `UPDATE tickets SET status = ?, updated_at = NOW(), last_updated_by = ? WHERE ticket_id = ?`
	_, err := os.db.Exec(query, status, updatedBy, orderID)
	return err
}

// GetOrdersByStatus is a placeholder and needs full implementation with joins for the new schema
func (os *OrderService) GetOrdersByStatus(status string) ([]Ticket, error) {
	// For now, return an empty slice and an error indicating it needs full implementation
	return []Ticket{}, fmt.Errorf("GetOrdersByStatus not fully implemented for new schema. Use GetAllOrders for now.")
}

// --- Global Database Connection and Setup ---

var db *sql.DB

// Database configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func getDBConfig() DBConfig {
	return DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		Database: getEnv("DB_NAME", "pcrepairhub"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDatabase() {
	config := getDBConfig()

	// Create DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.Database)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connection pool initialized successfully.")

	// Create tables if they don't exist
	createTables()
}

// createTables is updated to reflect the new normalized schema
func createTables() {
	// 1. Users table (Staff/Engineer)
	usersTable := `
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
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 2. Customers table (Client Information)
	customersTable := `
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
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 3. Device Details table (Equipment Information)
	deviceDetailsTable := `
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
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 4. Tickets table (Core Job and Status - Replaces old 'orders' table)
	ticketsTable := `
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
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// 5. Order Line Items table (Billing Items)
	orderLineItemsTable := `
    CREATE TABLE IF NOT EXISTS order_line_items (
        item_id VARCHAR(50) PRIMARY KEY,
        ticket_id VARCHAR(50) NOT NULL,
        service_name VARCHAR(255) NOT NULL,
        rate DECIMAL(10,2) NOT NULL,
        discount_percent DECIMAL(5,2) DEFAULT 0.00,
        final_price DECIMAL(10,2) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (ticket_id) REFERENCES tickets(ticket_id) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// Execute table creation in order
	if _, err := db.Exec(usersTable); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	if _, err := db.Exec(customersTable); err != nil {
		log.Fatalf("Failed to create customers table: %v", err)
	}
	if _, err := db.Exec(deviceDetailsTable); err != nil {
		log.Fatalf("Failed to create device_details table: %v", err)
	}
	if _, err := db.Exec(ticketsTable); err != nil {
		log.Fatalf("Failed to create tickets table: %v", err)
	}
	if _, err := db.Exec(orderLineItemsTable); err != nil {
		log.Fatalf("Failed to create order_line_items table: %v", err)
	}

	log.Println("Database tables created/verified successfully.")
}

// --- User Service (Unchanged) ---

// User represents a user account in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	FullName  string    `json:"full_name" db:"full_name"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone" db:"phone"`
	Password  string    `json:"password" db:"password"` // In production, this would be hashed
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserService handles user database operations
type UserService struct {
	db *sql.DB
}

func NewUserService(database *sql.DB) *UserService {
	return &UserService{db: database}
}

func (us *UserService) CreateUser(user *User) error {
	query := `
        INSERT INTO users (id, full_name, email, phone, password, role, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
    `

	_, err := us.db.Exec(query, user.ID, user.FullName, user.Email, user.Phone, user.Password, user.Role)
	return err
}

func (us *UserService) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `
        SELECT id, full_name, email, phone, password, role, created_at, updated_at
        FROM users WHERE email = ?
    `

	err := us.db.QueryRow(query, email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.Phone,
		&user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserService) GetUserByEmailAndPhone(email, phone string) (*User, error) {
	user := &User{}
	query := `
        SELECT id, full_name, email, phone, password, role, created_at, updated_at
        FROM users WHERE email = ? AND phone = ?
    `

	err := us.db.QueryRow(query, email, phone).Scan(
		&user.ID, &user.FullName, &user.Email, &user.Phone,
		&user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserService) UpdateUserPassword(userID, newPassword string) error {
	query := `UPDATE users SET password = ?, updated_at = NOW() WHERE id = ?`
	_, err := us.db.Exec(query, newPassword, userID)
	return err
}

func (us *UserService) EmailExists(email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = ?`
	err := us.db.QueryRow(query, email).Scan(&count)
	return count > 0, err
}

// --- Handler Functions ---

// Global service instances
var userService *UserService
var orderService *OrderService

// HealthCheckHandler provides a simple status check.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "PC Repair Hub API is operational")
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Basic validation
	if newUser.FullName == "" || newUser.Email == "" || newUser.Phone == "" || newUser.Password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Check if email already exists
	exists, err := userService.EmailExists(newUser.Email)
	if err != nil {
		log.Printf("Error checking email existence: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Set required fields for the new user
	newUser.ID = fmt.Sprintf("USER-%d", time.Now().UnixNano())
	newUser.Role = "User"

	// TODO: Hash the password before storing (use bcrypt)
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	// newUser.Password = string(hashedPassword)

	// Create user in database
	err = userService.CreateUser(&newUser)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s registered with email %s.", newUser.ID, newUser.Email)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
		"user_id": newUser.ID,
		"email":   newUser.Email,
	})
}

// LoginHandler handles user authentication
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Basic validation
	if loginRequest.Email == "" || loginRequest.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Check database for registered users
	user, err := userService.GetUserByEmail(loginRequest.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found - use same delay to prevent timing attacks
			time.Sleep(100 * time.Millisecond)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		log.Printf("Error retrieving user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: Use bcrypt to compare hashed password
	// For now, direct comparison (NOT SECURE - use bcrypt in production)
	if user.Password != loginRequest.Password {
		time.Sleep(100 * time.Millisecond) // Prevent timing attacks
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	log.Printf("User %s logged in successfully.", user.Email)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.FullName,
			"phone": user.Phone,
			"role":  user.Role,
		},
		"token": "demo-jwt-token", // In real app, generate actual JWT
	})
}

// GetDashboardMetricsHandler retrieves and aggregates key operational data.
func GetDashboardMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type result struct {
		Metric string
		Count  int
		Value  float64
	}

	results := make(chan result)

	// Goroutine 1: Get Total Open Orders (TICKETS table)
	go func() {
		// This should query the 'tickets' table: SELECT COUNT(*) FROM tickets WHERE status NOT IN ('Ready for Delivery', 'Collected')
		time.Sleep(10 * time.Millisecond)
		results <- result{Metric: "OpenOrders", Count: 150} // Mocked result
	}()

	// Goroutine 2: Get Ready for Delivery Count (TICKETS table)
	go func() {
		// This should query the 'tickets' table: SELECT COUNT(*) FROM tickets WHERE status = 'Ready for Delivery'
		time.Sleep(5 * time.Millisecond)
		results <- result{Metric: "ReadyCount", Count: 35} // Mocked result
	}()

	// Goroutine 3: Calculate YTD Revenue (TICKETS table)
	go func() {
		// This should query the 'tickets' table: SELECT SUM(total_cost) FROM tickets WHERE status = 'Collected' AND YEAR(created_at) = YEAR(CURDATE())
		time.Sleep(20 * time.Millisecond)
		results <- result{Metric: "Revenue", Value: 28550.75} // Mocked result
	}()

	// Collect results from goroutines
	metrics := DashboardMetrics{}
	received := 0
	for res := range results {
		switch res.Metric {
		case "OpenOrders":
			metrics.TotalOpenOrders = res.Count
		case "ReadyCount":
			metrics.ReadyForDelivery = res.Count
		case "Revenue":
			metrics.TotalRevenueYTD = res.Value
		}
		received++
		if received == 3 {
			close(results)
		}
	}

	json.NewEncoder(w).Encode(metrics)
}

// CreateOrderHandler handles the submission of a new service order (now a multi-table insert).
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var input TicketInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	// Basic validation
	if input.CustomerName == "" || input.CustomerPhone == "" || input.IssueDescription == "" {
		http.Error(w, "Customer name, phone, and issue description are required", http.StatusBadRequest)
		return
	}

	// Set created by for the transaction (Default to Admin if not provided by token/session)
	if input.CreatedBy == "" {
		input.CreatedBy = "ADMIN-001"
	}

	// Create ticket in database using the transactional function
	err = orderService.CreateTicket(&input)
	if err != nil {
		log.Printf("Error creating ticket: %v", err)
		http.Error(w, "Failed to create ticket: Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Ticket created for %s.", input.CustomerName)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":       "Ticket created successfully via normalized schema",
		"customer_name": input.CustomerName,
	})
}

// GetOrdersHandler retrieves all orders
func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	orders, err := orderService.GetAllOrders()
	if err != nil {
		log.Printf("Error retrieving orders: %v", err)
		http.Error(w, "Failed to retrieve orders", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orders)
}

// UpdateOrderStatusHandler updates the status of an order
func UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "PUT" {
		http.Error(w, "Only PUT method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var updateRequest struct {
		OrderID   string `json:"order_id"` // Corresponds to ticket_id
		Status    string `json:"status"`
		UpdatedBy string `json:"updated_by"`
	}

	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if updateRequest.OrderID == "" || updateRequest.Status == "" {
		http.Error(w, "Order ID and status are required", http.StatusBadRequest)
		return
	}

	// Validate status values
	validStatuses := []string{"New Order", "Diagnostics", "In Progress", "Ready for Delivery", "Collected"}
	isValidStatus := false
	for _, status := range validStatuses {
		if updateRequest.Status == status {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	err = orderService.UpdateOrderStatus(updateRequest.OrderID, updateRequest.Status, updateRequest.UpdatedBy)
	if err != nil {
		log.Printf("Error updating order status: %v", err)
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	log.Printf("Order %s status updated to %s by %s", updateRequest.OrderID, updateRequest.Status, updateRequest.UpdatedBy)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Order status updated successfully",
	})
}

// ForgotPasswordHandler handles password reset requests
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var resetRequest struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"new_password,omitempty"`
		Step     string `json:"step"` // "verify" or "reset"
	}

	err := json.NewDecoder(r.Body).Decode(&resetRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if resetRequest.Step == "verify" {
		// Verify user exists with email and phone
		user, err := userService.GetUserByEmailAndPhone(resetRequest.Email, resetRequest.Phone)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "No account found with this email and phone combination", http.StatusNotFound)
				return
			}
			log.Printf("Error verifying user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User verified successfully",
			"user_id": user.ID,
		})
	} else if resetRequest.Step == "reset" {
		// Reset password
		if resetRequest.Password == "" {
			http.Error(w, "New password is required", http.StatusBadRequest)
			return
		}

		user, err := userService.GetUserByEmailAndPhone(resetRequest.Email, resetRequest.Phone)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "No account found with this email and phone combination", http.StatusNotFound)
				return
			}
			log.Printf("Error finding user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// TODO: Hash the new password before storing
		err = userService.UpdateUserPassword(user.ID, resetRequest.Password)
		if err != nil {
			log.Printf("Error updating password: %v", err)
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		log.Printf("Password reset successfully for user %s", user.Email)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Password reset successfully",
		})
	} else {
		http.Error(w, "Invalid step parameter", http.StatusBadRequest)
	}
}

// --- Main Server Function ---

func main() {
	// Initialize database connection
	initDatabase()
	defer db.Close()

	// Initialize services
	userService = NewUserService(db)
	orderService = NewOrderService(db)

	// --- CONFIGURE CORS MIDDLEWARE ---
	corsHandler := cors.Default().Handler(http.DefaultServeMux)
	// ---------------------------------

	// Define the API routes
	http.HandleFunc("/api/v1/health", HealthCheckHandler)
	http.HandleFunc("/api/v1/dashboard/metrics", GetDashboardMetricsHandler)
	http.HandleFunc("/api/v1/orders", GetOrdersHandler)
	http.HandleFunc("/api/v1/orders/create", CreateOrderHandler)
	http.HandleFunc("/api/v1/orders/update-status", UpdateOrderStatusHandler)
	http.HandleFunc("/api/v1/auth/register", RegisterHandler)
	http.HandleFunc("/api/v1/auth/login", LoginHandler)
	http.HandleFunc("/api/v1/auth/forgot-password", ForgotPasswordHandler)

	// Start the server
	port := getEnv("PORT", "8080")
	if port[0] != ':' {
		port = ":" + port
	}

	log.Printf("PC Repair Hub Backend API starting on http://localhost%s", port)
	log.Printf("Database: %s", getDBConfig().Database)

	// ListenAndServe uses the CORS-wrapped handler (corsHandler)
	if err := http.ListenAndServe(port, corsHandler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
