# E-Commerce API

This is a sample e-commerce API built with Go and Gin framework. It provides endpoints for user authentication, product management, and order processing.

## Features

- User authentication (login, register)
- Product management (list, create, update, delete)
- Order management (create, list, update status, cancel)
- Swagger documentation

## Getting Started

### Prerequisites

- Go 1.16+

### Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/cgzirim/ecommerce-api.git
    cd ecommerce-api
    ```

2. Install dependencies:

    ```sh
    go mod tidy
    ```

3. Set up environment variables:

    Create a `.env` file in the root directory and add the necessary environment variables. Example:

    ```env
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=your_db_user
    DB_PASSWORD=your_db_password
    DB_NAME=ecommerce_db
    ```

4. Run the database migrations:

    ```sh
    go run main.go migrate
    ```

### Running the API

1. Start the server:

    ```sh
    go run main.go
    ```

2. The API will be available at `http://localhost:8080`.

### API Documentation

Swagger documentation is available at `http://localhost:8080/swagger/index.html`.

### Project Structure

- `controllers/`: Contains the handler functions for the API endpoints.
- `db/`: Database connection and migration scripts.
- `middleware/`: Custom middleware functions.
- `docs/`: Swagger documentation files.
