# Technical Context: Symper One Auth

## Technology Stack

### Core Language & Runtime
- **Go 1.23.7**: Latest stable Go version with modern features
- **CGO_ENABLED=0**: Static binary compilation for container deployment
- **Build Flags**: Version injection via ldflags

### Database & Storage
- **PostgreSQL**: Primary database (required, no alternatives)
- **Pop/Soda**: Database ORM and migration framework
- **Connection Pooling**: Configurable max pool size
- **Schema Namespace**: Configurable table prefix support

### HTTP Framework & Routing
- **Chi Router**: Lightweight, fast HTTP router
- **Middleware Stack**: Request ID, logging, CORS, rate limiting
- **CORS Support**: Configurable cross-origin resource sharing
- **Real IP Detection**: X-Forwarded-For and similar headers

### Authentication & Security
- **JWT Libraries**: 
  - `golang-jwt/jwt/v5`: JWT token generation and validation
  - `lestrrat-go/jwx/v2`: Advanced JWT operations
- **Cryptography**:
  - `golang.org/x/crypto`: Password hashing (bcrypt)
  - Custom crypto utilities for token generation
- **OAuth2**: `golang.org/x/oauth2` for external provider integration
- **WebAuthn**: `go-webauthn/webauthn` for FIDO2 support
- **SAML**: `crewjam/saml` for enterprise SSO

### External Integrations
- **Email**: `gopkg.in/gomail.v2` for SMTP email sending
- **SMS Providers**:
  - Twilio API integration
  - MessageBird API integration
  - TextLocal API integration
  - Vonage API integration
- **CAPTCHA**: hCaptcha and Turnstile support
- **HIBP**: Have I Been Pwned password checking

### Observability & Monitoring
- **OpenTelemetry**: Complete observability stack
  - `go.opentelemetry.io/otel`: Core telemetry
  - `go.opentelemetry.io/otel/exporters/prometheus`: Metrics export
  - `go.opentelemetry.io/otel/exporters/otlp`: OTLP export
- **Logging**: `sirupsen/logrus` with structured JSON output
- **Metrics**: Prometheus-compatible metrics
- **Tracing**: Distributed tracing with context propagation

### Development & Build Tools
- **Make**: Build automation and task management
- **Docker**: Multi-stage builds and development environment
- **Docker Compose**: Development orchestration
- **CompileDaemon**: Hot reloading in development
- **Git**: Version control with tag-based versioning

### Testing Framework
- **Testing**: Go standard library testing
- **Testify**: `stretchr/testify` for assertions and mocks
- **Coverage**: Built-in coverage reporting
- **Race Detection**: Concurrent testing with race detector

### Configuration Management
- **Environment Variables**: Primary configuration method
- **dotenv**: `joho/godotenv` for .env file support
- **envconfig**: `kelseyhightower/envconfig` for struct binding
- **File Watching**: `fsnotify/fsnotify` for hot reloading

## Development Environment

### Prerequisites
- **Go 1.23.7+**: Required for building
- **Docker & Docker Compose**: For development environment
- **PostgreSQL**: Database (can be containerized)
- **Make**: Build tool

### Development Setup
```bash
# Clone and setup
git clone <repository>
cd symper-one-auth

# Install dependencies
make deps
make dev-deps

# Start development environment
make dev

# Build binary
make build
```

### Development Tools
- **Soda**: Database migrations and management
- **gosec**: Security vulnerability scanning
- **staticcheck**: Static analysis
- **oapi-codegen**: OpenAPI client generation
- **exhaustive**: Enum exhaustiveness checking

### Hot Reloading
- **CompileDaemon**: Watches for Go file changes
- **Config Reloading**: Watches config directory for changes
- **Pattern Matching**: Excludes binaries and specific files
- **Automatic Rebuild**: Triggers on source changes

## Build & Deployment

### Build Process
```makefile
# Multi-architecture builds
CGO_ENABLED=0 go build $(FLAGS)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(FLAGS) -o auth-arm64
```

### Docker Strategy
- **Multi-stage Build**: Separate build and runtime stages
- **Base Images**:
  - Build: `golang:1.23.7-alpine3.20`
  - Runtime: `alpine:3` (latest for security updates)
- **Security**: Non-root user (supabase:1000)
- **Size Optimization**: Static binary, minimal runtime

### Container Configuration
```dockerfile
# Runtime environment
ENV GOTRUE_DB_MIGRATIONS_PATH /usr/local/etc/auth/migrations
USER supabase
CMD ["auth"]
```

### Environment Variables
- **Prefix**: All config uses `GOTRUE_` prefix
- **Fallbacks**: Some variables have non-prefixed alternatives
- **Validation**: Startup-time configuration validation
- **Secrets**: Secure handling of JWT secrets and API keys

## Database Architecture

### Schema Management
- **Migrations**: Versioned SQL migrations in `migrations/`
- **Auto-migration**: Runs on startup by default
- **Manual Migration**: `./auth migrate` command
- **Rollback**: Manual process, no automatic rollback

### Key Tables
- **users**: Core user accounts and authentication data
- **identities**: External provider identity linking
- **sessions**: User session management
- **refresh_tokens**: Refresh token storage and rotation
- **mfa_factors**: Multi-factor authentication factors
- **mfa_challenges**: MFA challenge tracking
- **audit_log_entries**: Security audit trail
- **flow_state**: OAuth/PKCE flow state management
- **sso_providers**: SAML SSO provider configuration
- **oauth_clients**: OAuth client management

### Database Features
- **UUID Primary Keys**: All entities use UUIDs
- **Soft Deletes**: Some entities support soft deletion
- **Indexes**: Optimized for common query patterns
- **Constraints**: Foreign keys and unique constraints
- **Row Level Security**: PostgreSQL RLS support

## Security Architecture

### Token Management
- **JWT Structure**: Standard claims + custom claims
- **Signing**: HMAC-SHA256 (configurable algorithm)
- **Expiration**: Configurable token lifetime
- **Refresh**: Secure refresh token rotation

### Password Security
- **Hashing**: bcrypt with configurable cost
- **Complexity**: Configurable password requirements
- **Breach Detection**: HIBP integration
- **Recovery**: Secure token-based reset

### Rate Limiting
- **Implementation**: In-memory or Redis-backed
- **Granularity**: Per-endpoint, per-IP, per-user
- **Configuration**: Flexible rate limit definitions
- **Headers**: Standard rate limit response headers

### Audit & Compliance
- **Audit Logging**: All authentication events logged
- **IP Tracking**: Client IP address logging
- **Session Tracking**: Complete session lifecycle
- **Admin Actions**: Administrative action audit trail

## API Design

### REST Principles
- **Resource-based URLs**: `/users`, `/sessions`, `/factors`
- **HTTP Methods**: Proper use of GET, POST, PUT, DELETE
- **Status Codes**: Appropriate HTTP status codes
- **Content Types**: JSON request/response bodies

### Authentication
- **Bearer Tokens**: JWT tokens in Authorization header
- **Admin Endpoints**: Service role or admin role required
- **Public Endpoints**: Settings, health checks
- **CORS**: Configurable cross-origin support

### Error Handling
- **Consistent Format**: Structured error responses
- **Error Codes**: Application-specific error codes
- **Localization**: Configurable error messages
- **Security**: No sensitive data in error responses

### OpenAPI Specification
- **Documentation**: Complete API specification
- **Client Generation**: Automated client library generation
- **Validation**: Request/response validation
- **Testing**: API contract testing

## Performance & Scalability

### Horizontal Scaling
- **Stateless Design**: No server-side session state
- **Database Scaling**: PostgreSQL read replicas supported
- **Load Balancing**: Standard HTTP load balancing
- **Caching**: JWT validation caching

### Performance Optimizations
- **Connection Pooling**: Database connection reuse
- **Prepared Statements**: SQL query optimization
- **Indexes**: Optimized database queries
- **Compression**: HTTP response compression

### Monitoring & Alerting
- **Health Checks**: `/health` endpoint for monitoring
- **Metrics**: Prometheus-compatible metrics
- **Tracing**: Distributed tracing support
- **Logging**: Structured logging for analysis

## Deployment Considerations

### Infrastructure Requirements
- **CPU**: Minimal CPU requirements
- **Memory**: ~100MB base memory usage
- **Storage**: PostgreSQL database storage
- **Network**: HTTPS termination recommended

### Configuration Management
- **Environment Variables**: 12-factor app configuration
- **Secrets Management**: External secret injection
- **Feature Flags**: Configuration-based feature toggles
- **Multi-environment**: Environment-specific configurations

### Security Hardening
- **TLS**: HTTPS/TLS termination at load balancer
- **Firewall**: Database access restrictions
- **Secrets**: Secure JWT secret management
- **Updates**: Regular security update process
