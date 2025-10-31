# Python Microservice with MySQL

A 2-tier microservice architecture with Flask REST API and MySQL database.

## Architecture
- **Presentation Tier**: Flask REST API with CRUD operations
- **Data Tier**: MySQL database with SQLAlchemy ORM

## Quick Start

### Using Docker Compose (Recommended)
```bash
docker-compose up --build
```

### Manual Setup
1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Start MySQL server and create database:
```sql
CREATE DATABASE microservice_db;
```

3. Run the application:
```bash
python run.py
```

## API Endpoints

- `GET /health` - Health check
- `GET /users` - Get all users
- `POST /users` - Create user
- `GET /users/{id}` - Get user by ID
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user

## Example Usage

Create a user:
```bash
curl -X POST http://localhost:5000/users \
  -H "Content-Type: application/json" \
  -d '{"username": "john_doe", "email": "john@example.com"}'
```

Get all users:
```bash
curl http://localhost:5000/users
```