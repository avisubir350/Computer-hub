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
)

// --- Domain Models (structs) ---

// Order represents the structure of a repair ticket, mapped to a MySQL table row.
type Order struct {
	ID               string    `json:"id" db:"id"`
	CustomerName     string    `json:"customer_name" db:"customer_name"`
	CustomerEmail    string    `json:"customer_email" db:"customer_email"`
	CustomerPhone    string    `json:"customer_phone" db:"customer_phone"`
	DeviceType       string    `json:"device_type" db:"device_type"`
	DeviceModel      string    `json:"device_model" db:"device_model"`
	Services         []string  `json:"services" db:"services"` // Will be JSON in DB
	IssueDescription string    `json:"issue_description" db:"issue_description"`
	Status           string    `json:"status" db:"status"`
	TotalCost        float64   `json:"total_cost" db:"total_cost"`
	CreatedBy        string    `json:"created_by" db:"created_by"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	LastUpdatedBy    string    `json:"last_updated_by" db:"last_updated_by"`
}

// OrderService handles order database operations
type OrderService struct {
	db *sql.DB
}

func NewOrderService(database *sql.DB) *OrderService {
	return &OrderService{db: database}
}

func (os *OrderService) CreateOrder(order *Order) error {
	servicesJSON, err := json.Marshal(order.Services)
	if err != nil {
		return err
	}
	
	query := `
		INSERT INTO orders (id, customer_name, customer_email, customer_phone, device_type, 
		                   device_model, services, issue_description, status, total_cost, 
		                   created_by, created_at, updated_at, last_updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), ?)
	`
	
	_, err = os.db.Exec(query, order.ID, order.CustomerName, order.CustomerEmail, 
		order.CustomerPhone, order.DeviceType, order.DeviceModel, string(servicesJSON),
		order.IssueDescription, order.Status, order.TotalCost, order.CreatedBy, order.CreatedBy)
	
	return err
}

func (os *OrderService) GetAllOrders() ([]Order, error) {
	query := `
		SELECT id, customer_name, customer_email, customer_phone, device_type, device_model,
		       services, issue_description, status, total_cost, created_by, created_at, 
		       updated_at, COALESCE(last_updated_by, created_by)
		FROM orders ORDER BY created_at DESC
	`
	
	rows, err := os.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []Order
	for rows.Next() {
		var order Order
		var servicesJSON string
		
		err := rows.Scan(&order.ID, &order.CustomerName, &order.CustomerEmail,
			&order.CustomerPhone, &order.DeviceType, &order.DeviceModel,
			&servicesJSON, &order.IssueDescription, &order.Status, &order.TotalCost,
			&order.CreatedBy, &order.CreatedAt, &order.UpdatedAt, &order.LastUpdatedBy)
		
		if err != nil {
			return nil, err
		}
		
		// Parse services JSON
		if err := json.Unmarshal([]byte(servicesJSON), &order.Services); err != nil {
			return nil, err
		}
		
		orders = append(orders, order)
	}
	
	return orders, nil
}

func (os *OrderService) UpdateOrderStatus(orderID, status, updatedBy string) error {
	query := `UPDATE orders SET status = ?, updated_at = NOW(), last_updated_by = ? WHERE id = ?`
	_, err := os.db.Exec(query, status, updatedBy, orderID)
	return err
}

func (os *OrderService) GetOrdersByStatus(status string) ([]Order, error) {
	query := `
		SELECT id, customer_name, customer_email, customer_phone, device_type, device_model,
		       services, issue_description, status, total_cost, created_by, created_at, 
		       updated_at, COALESCE(last_updated_by, created_by)
		FROM orders WHERE status = ? ORDER BY created_at DESC
	`
	
	rows, err := os.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []Order
	for rows.Next() {
		var order Order
		var servicesJSON string
		
		err := rows.Scan(&order.ID, &order.CustomerName, &order.CustomerEmail,
			&order.CustomerPhone, &order.DeviceType, &order.DeviceModel,
			&servicesJSON, &order.IssueDescription, &order.Status, &order.TotalCost,
			&order.CreatedBy, &order.CreatedAt, &order.UpdatedAt, &order.LastUpdatedBy)
		
		if err != nil {
			return nil, err
		}
		
		// Parse services JSON
		if err := json.Unmarshal([]byte(servicesJSON), &order.Services); err != nil {
			return nil, err
		}
		
		orders = append(orders, order)
	}
	
	return orders, nil
}

// DashboardMetrics holds the aggregated data for the operational dashboard.
type DashboardMetrics struct {
	TotalOpenOrders    int `json:"total_open_orders"`
	ReadyForDelivery   int `json:"ready_for_delivery"`
	TotalRevenueYTD    float64 `json:"total_revenue_ytd"`
}

// --- Global Database Connection ---

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

func createTables() {
	// Users table
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

	// Orders/Tickets table
	ordersTable := `
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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	// Execute table creation
	if _, err := db.Exec(usersTable); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	
	if _, err := db.Exec(ordersTable); err != nil {
		log.Fatalf("Failed to create orders table: %v", err)
	}
	
	log.Println("Database tables created/verified successfully.")
}

// --- Handler Functions ---

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

// HealthCheckHandler provides a simple status check.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "PC Repair Hub API is operational")
}

// Global service instances
var userService *UserService
var orderService *OrderService

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
		"email": newUser.Email,
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

	// Demo admin account (hardcoded)
	if loginRequest.Email == "admin@pchub.com" && loginRequest.Password == "admin123" {
		log.Printf("Admin user %s logged in successfully.", loginRequest.Email)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Login successful",
			"user": map[string]string{
				"email": loginRequest.Email,
				"name":  "Admin User",
				"role":  "Administrator",
			},
			"token": "demo-jwt-token", // In real app, generate actual JWT
		})
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
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
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

	// --- Concurrency / Scale Feature ---
	// This function shows how Go handles multiple database queries concurrently
	// using goroutines, which is crucial for high-speed dashboard loading.

	type result struct {
		Metric string
		Count  int
		Value  float64
	}

	results := make(chan result)
	
	// Goroutine 1: Get Total Open Orders
	go func() {
		// Mocked DB query: SELECT COUNT(*) FROM orders WHERE status NOT IN ('Ready', 'Collected')
		// In a real app: row := db.QueryRow(query)
		time.Sleep(10 * time.Millisecond) // Simulate DB latency
		results <- result{Metric: "OpenOrders", Count: 150} // Mocked result
	}()

	// Goroutine 2: Get Ready for Delivery Count
	go func() {
		// Mocked DB query: SELECT COUNT(*) FROM orders WHERE status = 'Ready for Delivery'
		time.Sleep(5 * time.Millisecond)
		results <- result{Metric: "ReadyCount", Count: 35} // Mocked result
	}()

	// Goroutine 3: Calculate YTD Revenue
	go func() {
		// Mocked DB query: SELECT SUM(estimated_cost) FROM orders WHERE status = 'Collected' AND YEAR(created_at) = YEAR(CURDATE())
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

// CreateOrderHandler handles the submission of a new service order.
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var newOrder Order
	err := json.NewDecoder(r.Body).Decode(&newOrder)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Basic validation
	if newOrder.CustomerName == "" || newOrder.CustomerEmail == "" || newOrder.CustomerPhone == "" {
		http.Error(w, "Customer information is required", http.StatusBadRequest)
		return
	}

	// Set required fields for the new order
	newOrder.ID = fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	newOrder.Status = "New Order"

	// Create order in database
	err = orderService.CreateOrder(&newOrder)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	log.Printf("Order %s created for %s.", newOrder.ID, newOrder.CustomerName)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Order created successfully", 
		"order_id": newOrder.ID,
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
		OrderID   string `json:"order_id"`
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
	validStatuses := []string{"New Order", "In Progress", "Ready for Delivery", "Collected"}
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

	// Enable CORS for frontend integration
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Handle other requests normally
		http.DefaultServeMux.ServeHTTP(w, r)
	})

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
	
	// ListenAndServe uses a new goroutine for every incoming request, 
	// leveraging Go's highly efficient concurrency model (goroutines) to handle scale.
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
