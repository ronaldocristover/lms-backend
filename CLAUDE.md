# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based backend for a Learning Management System (LMS) following **Clean Architecture** principles with layered separation of concerns.

## Development Commands

### Building and Running
- `make build` - Build the application to `bin/server`
- `make run` - Run the application (uses default environment)
- `make dev` - Run with environment variables from `.env` file
- `go run ./cmd/server` - Direct run without Make

### Testing
- `make test` - Run all tests
- `go test ./...` - Run tests using Go directly
- Tests cover unit tests for services, repositories, and handlers

### Database Operations
- `make migrate-up` - Run database migrations (requires PostgreSQL)
- `make migrate-down` - Rollback database migrations
- Database auto-migration happens on application startup

### Code Generation
- Swagger docs: Add `// @title` etc. annotations to handlers, then run `swag init`
- Main function in `cmd/server/main.go` has Swagger initialization

## Architecture

### Clean Architecture Layers
The codebase follows Clean Architecture with strict dependency inversion:

1. **Handler Layer** (`internal/handler/`) - HTTP request/response handling
   - RESTful API endpoints
   - Request validation and response formatting
   - Thin layer, delegates to services

2. **Service Layer** (`internal/service/`) - Business logic
   - Use cases and business rules
   - Orchestrate repositories
   - Handle transactions and coordination

3. **Repository Layer** (`internal/repository/`) - Data access
   - GORM for database operations
   - Abstract data access behind interfaces
   - Handle database-specific queries

4. **Model Layer** (`internal/model/`) - Domain entities
   - Structs for all domain entities
   - Database models with GORM tags
   - Relationships between entities

### Key Entities
- User: Authentication, profiles, roles
- Organization: Multi-tenancy support
- Content: Learning materials (videos, documents)
- Series: Content organization
- Session: User learning sessions
- Media: File uploads and management
- Role: RBAC implementation

### Supporting Packages
- `pkg/apierror/` - Custom API error types
- `pkg/pagination/` - Pagination helpers
- `pkg/response/` - HTTP response utilities
- `pkg/validator/` - Validation utilities

## Important Patterns

### Dependency Injection
- Services receive repositories via constructor injection
- Use interface types for dependencies
- Concrete implementations passed at runtime

### Error Handling
- Custom error types in `pkg/apierror/`
- Wrap errors with context
- Return appropriate HTTP status codes

### Authentication
- JWT tokens with refresh mechanism
- Bcrypt password hashing
- Role-based access control (RBAC)
- Claims-based authorization in middleware

### Background Jobs
- Worker pool implementation in `internal/scheduler/`
- Queue-based job processing
- Configurable worker count and queue size

## Configuration

### Environment Variables
Copy `.env.example` to `.env` and configure:
- Database connection (PostgreSQL)
- JWT secrets and expiry
- CORS settings
- File upload limits
- Background job settings

### Database Setup
- PostgreSQL required
- Auto-migration on startup
- Migrations in `migrations/` directory (though recent refactoring may have moved this)

## File Uploads
- Stored in `./uploads` directory
- Max configurable size (default: 10MB)
- Support for various media types
- S3 integration possible via configuration

## Testing Strategy
- Unit tests for services and repositories
- Integration tests for handlers
- Model validation tests
- Test helpers for common fixtures
- Use `testify` for assertions and mocks

## API Documentation
- Swagger/OpenAPI generated from annotations
- Available at `/swagger/index.html` when running
- Add Swagger comments to handler functions
- Use `swag init` to regenerate docs

## Recent Refactoring
The codebase has undergone significant refactoring:
- Interface segregation in UserService
- Context usage improvements
- Main function extraction
- Error wrapping fixes

## Key Dependencies
- Gin for HTTP routing
- GORM for ORM
- Zap for logging
- Viper for configuration
- JWT for authentication
- Swagger for API docs