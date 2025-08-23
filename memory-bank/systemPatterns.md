# System Patterns: Symper One Auth

## Architecture Overview
Supabase Auth follows a clean, layered architecture with clear separation of concerns:

```
┌─────────────────┐
│   HTTP Layer    │ ← Chi Router, Middleware, CORS
├─────────────────┤
│   API Layer     │ ← Handlers, Authentication, Validation
├─────────────────┤
│  Business Logic │ ← User Management, Token Generation, MFA
├─────────────────┤
│   Data Layer    │ ← Models, Database Operations
├─────────────────┤
│  Storage Layer  │ ← PostgreSQL, Migrations
└─────────────────┘
```

## Core Design Patterns

### 1. Command Pattern (CLI)
- **Location**: `cmd/` directory
- **Pattern**: Each command is a separate module with its own handler
- **Commands**: `serve`, `migrate`, `admin`, `version`
- **Benefits**: Clean separation of CLI functionality, easy to extend

### 2. Repository Pattern (Data Access)
- **Location**: `internal/models/` and `internal/storage/`
- **Pattern**: Database operations abstracted through model methods
- **Key Models**: `User`, `Factor`, `Session`, `RefreshToken`, `Identity`
- **Benefits**: Database-agnostic business logic, testable data operations

### 3. Middleware Chain Pattern (HTTP)
- **Location**: `internal/api/middleware.go`
- **Pattern**: Composable middleware for cross-cutting concerns
- **Middleware**: Authentication, Rate Limiting, CORS, Request ID, Logging
- **Benefits**: Reusable, composable request processing

### 4. Factory Pattern (API Creation)
- **Location**: `internal/api/api.go`
- **Pattern**: API instances created with configuration and dependencies
- **Usage**: `NewAPI()`, `NewAPIWithVersion()`
- **Benefits**: Consistent initialization, dependency injection

### 5. Strategy Pattern (Authentication Methods)
- **Location**: `internal/api/` (various auth handlers)
- **Pattern**: Different authentication strategies for different methods
- **Strategies**: Password, OAuth, Magic Link, Phone, MFA, SAML
- **Benefits**: Extensible authentication methods, clean separation

## Key Components

### HTTP Router (Chi)
```go
// Router setup with middleware chain
r := chi.NewRouter()
r.Use(middleware.RequestID)
r.Use(middleware.RealIP)
r.Use(middleware.Logger)
r.Use(cors.Handler(corsOptions))
```

### Database Models
```go
type User struct {
    ID                uuid.UUID          `json:"id" db:"id"`
    Email             storage.NullString `json:"email" db:"email"`
    EncryptedPassword *string            `json:"-" db:"encrypted_password"`
    // ... other fields
}
```

### Configuration Management
```go
type GlobalConfiguration struct {
    API      APIConfiguration
    Database DatabaseConfiguration
    JWT      JWTConfiguration
    External ExternalConfiguration
    // ... other config sections
}
```

## Data Flow Patterns

### 1. Authentication Flow
```
Request → Middleware → Handler → Business Logic → Database → Response
```

### 2. Token Generation Flow
```
User Auth → JWT Claims → Sign Token → Store Refresh Token → Return Tokens
```

### 3. OAuth Flow
```
Redirect → Provider → Callback → User Creation/Linking → Token Generation
```

## Security Patterns

### 1. JWT Token Management
- **Access Tokens**: Short-lived (1 hour default), stateless
- **Refresh Tokens**: Long-lived, stored in database, rotatable
- **Token Rotation**: Optional security feature to prevent token reuse

### 2. Password Security
- **Hashing**: bcrypt with configurable cost
- **Validation**: Configurable complexity requirements
- **Recovery**: Secure token-based password reset

### 3. Rate Limiting
- **Implementation**: Configurable rate limits per endpoint
- **Storage**: In-memory or Redis-backed
- **Granularity**: Per-IP, per-user, per-endpoint

### 4. Audit Logging
- **Pattern**: Comprehensive logging of all authentication events
- **Storage**: Database-backed audit log entries
- **Fields**: User ID, IP address, action, timestamp, metadata

## Integration Patterns

### 1. Webhook System
- **Pattern**: Event-driven architecture for custom logic
- **Events**: User creation, login, password change, etc.
- **Delivery**: HTTP POST with signed payloads
- **Retry Logic**: Exponential backoff for failed deliveries

### 2. Email Integration
- **Pattern**: Template-based email system
- **Providers**: SMTP, custom templates
- **Templates**: Confirmation, recovery, magic link, invite
- **Customization**: URL templates, custom subjects

### 3. SMS Integration
- **Pattern**: Provider abstraction for SMS delivery
- **Providers**: Twilio, MessageBird, TextLocal, Vonage
- **Templates**: Configurable OTP message templates
- **Validation**: E.164 phone number format

## Error Handling Patterns

### 1. API Error Responses
```go
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"msg"`
    Details string `json:"details,omitempty"`
}
```

### 2. Database Error Handling
- **Pattern**: Specific error types for different database conditions
- **Errors**: Unique constraint violations, foreign key errors, etc.
- **Recovery**: Graceful degradation and user-friendly messages

### 3. External Service Errors
- **Pattern**: Circuit breaker pattern for external dependencies
- **Fallbacks**: Graceful degradation when external services fail
- **Retry Logic**: Exponential backoff with jitter

## Observability Patterns

### 1. Structured Logging
- **Format**: JSON-structured logs with consistent fields
- **Levels**: Debug, Info, Warn, Error, Fatal
- **Context**: Request ID, user ID, correlation IDs

### 2. Metrics Collection
- **Pattern**: OpenTelemetry-based metrics
- **Metrics**: Request counts, response times, error rates
- **Exporters**: Prometheus, OTLP

### 3. Distributed Tracing
- **Pattern**: OpenTelemetry traces across service boundaries
- **Spans**: HTTP requests, database operations, external calls
- **Context**: Trace propagation across service calls

## Testing Patterns

### 1. Unit Testing
- **Pattern**: Table-driven tests with comprehensive coverage
- **Mocking**: Database and external service mocks
- **Fixtures**: Reusable test data and setup

### 2. Integration Testing
- **Pattern**: Full API testing with real database
- **Setup**: Docker-based test database
- **Cleanup**: Transaction rollback or database reset

### 3. End-to-End Testing
- **Pattern**: Complete authentication flows
- **Coverage**: All authentication methods and edge cases
- **Environment**: Isolated test environment

## Deployment Patterns

### 1. Container Deployment
- **Pattern**: Multi-stage Docker builds
- **Optimization**: Minimal runtime image with security updates
- **Configuration**: Environment variable injection

### 2. Database Migrations
- **Pattern**: Versioned, forward-only migrations
- **Automation**: Automatic migration on startup
- **Rollback**: Manual rollback procedures documented

### 3. Configuration Management
- **Pattern**: Environment-based configuration with defaults
- **Hot Reloading**: Development-time configuration updates
- **Validation**: Startup-time configuration validation
