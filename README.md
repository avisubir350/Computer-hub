# PC Repair Hub

A comprehensive computer repair management system with user registration, service order tracking, and dashboard analytics.

## Features

- **User Management**: Registration, login, password reset
- **Service Orders**: Create, track, and manage repair orders
- **Dashboard**: Real-time metrics and order status tracking
- **Indian Standards**: Phone number validation and INR pricing
- **MySQL Database**: Persistent data storage with proper relationships

## Tech Stack

- **Frontend**: HTML5, CSS3 (Tailwind), JavaScript (Vanilla)
- **Backend**: Go (Golang) with MySQL
- **Database**: MySQL 8.0+
- **Icons**: Phosphor Icons
- **Fonts**: Inter (Google Fonts)

## Prerequisites

- Go 1.21 or higher
- MySQL 8.0 or higher
- Git

## Installation

### 1. Clone the Repository
```bash
git clone <repository-url>
cd pcrepairhub
```

### 2. Database Setup

#### Install MySQL
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install mysql-server

# macOS (using Homebrew)
brew install mysql

# Windows
# Download from https://dev.mysql.com/downloads/mysql/
```

#### Create Database
```bash
# Login to MySQL
mysql -u root -p

# Run the setup script
source database/setup.sql
```

### 3. Environment Configuration
```bash
# Copy environment template
cp .env.example .env

# Edit with your MySQL credentials
nano .env
```

### 4. Install Go Dependencies
```bash
go mod tidy
```

### 5. Run the Application
```bash
# Start the Go server
go run main.go
```

### 6. Access the Application
- Open your browser and navigate to the HTML files:
  - `login.html` - Login page
  - `register.html` - Registration page
  - `index.html` - Dashboard (requires login)
  - `forgot-password.html` - Password reset

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/forgot-password` - Password reset

### Orders
- `GET /api/v1/orders` - Get all orders
- `POST /api/v1/orders/create` - Create new order
- `PUT /api/v1/orders/update-status` - Update order status

### System
- `GET /api/v1/health` - Health check
- `GET /api/v1/dashboard/metrics` - Dashboard metrics

## Database Schema

### Users Table
```sql
- id (VARCHAR(50), PRIMARY KEY)
- full_name (VARCHAR(255))
- email (VARCHAR(255), UNIQUE)
- phone (VARCHAR(20))
- password (VARCHAR(255))
- role (VARCHAR(50))
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

### Orders Table
```sql
- id (VARCHAR(50), PRIMARY KEY)
- customer_name (VARCHAR(255))
- customer_email (VARCHAR(255))
- customer_phone (VARCHAR(20))
- device_type (VARCHAR(255))
- device_model (VARCHAR(255))
- services (JSON)
- issue_description (TEXT)
- status (ENUM)
- total_cost (DECIMAL(10,2))
- created_by (VARCHAR(50))
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
- last_updated_by (VARCHAR(50))
```

## Configuration

### Environment Variables
- `DB_HOST` - MySQL host (default: localhost)
- `DB_PORT` - MySQL port (default: 3306)
- `DB_USER` - MySQL username (default: root)
- `DB_PASSWORD` - MySQL password
- `DB_NAME` - Database name (default: pcrepairhub)
- `PORT` - Server port (default: 8080)

### Default Credentials
- **Admin**: admin@pchub.com / admin123
- **Sample User**: john@example.com / password123

## Development

### Project Structure
```
pcrepairhub/
├── main.go              # Go backend server
├── go.mod               # Go dependencies
├── database/
│   └── setup.sql        # Database schema and sample data
├── .env.example         # Environment template
├── index.html           # Dashboard page
├── login.html           # Login page
├── register.html        # Registration page
├── forgot-password.html # Password reset page
└── README.md           # This file
```

### Adding New Features
1. Update database schema in `database/setup.sql`
2. Add new structs and services in `main.go`
3. Create new API endpoints
4. Update frontend HTML/JavaScript

## Security Notes

⚠️ **Important**: This is a development version. For production:

1. **Hash Passwords**: Implement bcrypt password hashing
2. **JWT Tokens**: Add proper JWT authentication
3. **Input Validation**: Add comprehensive input sanitization
4. **HTTPS**: Use TLS/SSL certificates
5. **Environment Variables**: Use secure environment variable management
6. **Database Security**: Use connection pooling and prepared statements
7. **Rate Limiting**: Implement API rate limiting

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check MySQL is running: `sudo systemctl status mysql`
   - Verify credentials in `.env` file
   - Ensure database exists: `SHOW DATABASES;`

2. **Go Dependencies Error**
   - Run: `go mod tidy`
   - Check Go version: `go version`

3. **Port Already in Use**
   - Change PORT in `.env` file
   - Kill existing process: `lsof -ti:8080 | xargs kill`

4. **CORS Issues**
   - Serve HTML files through a web server
   - Use Live Server extension in VS Code

## License

This project is for educational purposes. Please ensure proper security measures before production use.# Computer-hub
