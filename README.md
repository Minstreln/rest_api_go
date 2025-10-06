# School Management Server

A robust **Golang + MySQL(MariaDB)** backend for managing a school system. This server handles **students, teachers, and executives**, with authentication, role-based access, sorting, filtering, pagination, and full CRUD functionality. Built using Go’s `net/http` package for simplicity and performance.

---

## Table of Contents

- [Features](#features)
- [Technologies](#technologies)
- [Architecture](#architecture)
- [Setup](#setup)
- [Database](#database)
- [API Endpoints](#api-endpoints)
- [Authentication & Security](#authentication--security)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- **User Roles**: Separate roles for students, teachers, and executives.
- **Authentication**: Secure login and token management using JWT.
- **Role-Based Access Control (RBAC)**: Users can only access resources allowed by their role.
- **CRUD Operations**: Create, read, update, and delete records for all entities.
- **Sorting & Filtering**: Supports querying data with sorting, filtering amd pagination.
- **Input Validation & Sanitization**: Prevents XSS attacks and ensures data integrity.
- **RESTful API**: Clean and consistent endpoints for all resources.

---

## Technologies

- **Language**: Go (Golang)
- **Database**: MariaDB / MySQL
- **HTTP Server**: `net/http`
- **Sanitization**: [Bluemonday](https://github.com/microcosm-cc/bluemonday) for XSS protection
- **Password Security**: `password` hashing for security
- **JSON Handling**: `encoding/json`

---

## Architecture

- **Models**: Define `Student`, `Teacher`, and `Executive` structs.
- **Handlers**: HTTP handlers for each CRUD operation.
- **Middleware**: Authentication, authorization, and input sanitization.
- **Database Layer**: Handles connections, queries, and migrations.

---

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/Minstreln/rest_api_go
```

## Configure environment variables - example below

DB_USER=db_user
DB_PASSWORD=db_password
DB_NAME=db_name
SERVER_PORT=server_port
DB_PORT=db_port
HOST=db_host
JWT_SECRET=jwt_secret
JWT_EXPIRES_IN=6000s
RESET_TOKEN_EXP_DURATION=reset_token_exp_duration
CERT_FILE="your_cert.pem"
KEY_FILE="your_key.pem"

```
THIS SERVER USES TLS. YOU CAN DISABLE IT IN THE cmd/api/server.go FILE
```

## Database Migrations

All database migrations are stored in the `internal/migrations/` folder. These migrations set up the tables for **students, teachers, and executives**.

### Folder Structure

internal/migrations/
├── 001_create_execs.sql
├── 002_create_students.sql
├── 003_create_teachers.sql

## Install dependencies

go mod tidy

## Run server

go run cmd/api/server.go

## Postman Collection

You can test all API endpoints using this [Postman Collection](https://subsum.postman.co/workspace/Go-REST-API~3d71388c-8d2d-42ff-ba1d-fbf43a22b38c/collection/27481035-73d58acc-2e49-4c65-9e8d-31f7aacfe470?action=share&creator=27481035&active-environment=27481035-893d893d-6cc9-4551-a0c0-b01fbb2be4c8).
