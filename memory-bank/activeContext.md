# Active Context: Symper One Auth

## Current Project State
**Status**: Initial project discovery and setup phase  
**Date**: January 2025  
**Focus**: Understanding the codebase and establishing development environment

## Recent Discoveries

### Project Identity
- This is **Supabase Auth**, a production-ready authentication microservice
- Originally based on Netlify's GoTrue but significantly diverged
- Written in Go 1.23.7 with comprehensive feature set
- Battle-tested in Supabase's production infrastructure

### Architecture Understanding
- Clean layered architecture with clear separation of concerns
- RESTful API with comprehensive endpoint coverage
- PostgreSQL-only database requirement (no alternatives)
- Docker-first deployment strategy with development hot-reloading

### Key Capabilities Identified
- **15+ OAuth providers** supported out of the box
- **Multiple authentication methods**: email/password, magic links, phone/SMS, MFA
- **Enterprise features**: SAML SSO, audit logging, admin APIs
- **Modern security**: JWT with refresh token rotation, rate limiting, CAPTCHA
- **Developer experience**: OpenAPI spec, generated clients, comprehensive docs

## Current Work Focus

### Immediate Tasks
1. ✅ **Project Discovery**: Completed comprehensive codebase analysis
2. 🔄 **Memory Bank Creation**: Creating comprehensive documentation
3. ⏳ **Environment Setup**: Preparing to install and configure
4. ⏳ **Service Startup**: Planning to start the authentication service

### Development Environment Setup Plan
1. **Dependencies**: Install Go 1.23.7, Docker, Docker Compose
2. **Configuration**: Create `.env` file from `example.env`
3. **Database**: Start PostgreSQL container via Docker Compose
4. **Build**: Compile the auth binary using Make
5. **Startup**: Launch the service and verify health endpoint

## Configuration Strategy

### Environment Configuration
- **Base Config**: Using `example.env` as template
- **Database**: PostgreSQL connection to containerized instance
- **JWT Secret**: Generate secure secret for token signing
- **Development Mode**: Enable auto-confirmation for easier testing
- **External Providers**: Initially disabled, can enable as needed

### Key Configuration Decisions
- **Port**: Default 9999 for API server
- **Database**: Use Docker Compose PostgreSQL service
- **Email**: Start with auto-confirm mode (no SMTP required)
- **Logging**: Debug level for development
- **Security**: Development-friendly settings initially

## Technical Decisions Made

### Development Approach
- **Docker-based Development**: Using provided Docker Compose setup
- **Hot Reloading**: Leveraging CompileDaemon for development efficiency
- **Database Migrations**: Automatic migration on startup
- **Configuration**: Environment variable based with .env file support

### Architecture Insights
- **Stateless Design**: JWT-based authentication with database-stored refresh tokens
- **Middleware Chain**: Comprehensive middleware for cross-cutting concerns
- **Repository Pattern**: Clean data access layer with Pop/Soda ORM
- **Command Pattern**: CLI commands for different operational modes

## Next Steps

### Immediate Actions
1. **Complete Memory Bank**: Finish documentation creation
2. **Environment Setup**: Install dependencies and configure environment
3. **Service Startup**: Launch PostgreSQL and Auth service
4. **Health Check**: Verify service is running correctly
5. **Basic Testing**: Test core authentication endpoints

### Future Considerations
- **OAuth Provider Setup**: Configure external authentication providers
- **SMTP Configuration**: Set up email sending for production-like testing
- **Admin API Testing**: Explore administrative functionality
- **Integration Testing**: Test with sample frontend application
- **Production Deployment**: Plan for production-ready configuration

## Development Environment Requirements

### System Prerequisites
- **Go 1.23.7+**: Required for building the application
- **Docker & Docker Compose**: For database and development environment
- **Make**: Build tool for compilation and task automation
- **Git**: Version control (already present)

### Optional Tools
- **PostgreSQL Client**: For direct database access
- **HTTP Client**: Postman, curl, or similar for API testing
- **JWT Debugger**: For token inspection and validation

## Configuration Files Status
- ✅ `example.env`: Comprehensive configuration template available
- ✅ `example.docker.env`: Docker-specific configuration template
- ✅ `docker-compose-dev.yml`: Development environment orchestration
- ✅ `Makefile`: Build and development task automation
- ⏳ `.env`: To be created from example template

## Key Endpoints to Test
- `GET /health`: Service health check
- `GET /settings`: Public configuration
- `POST /signup`: User registration
- `POST /token`: Authentication
- `GET /user`: User profile (authenticated)
- `POST /logout`: Session termination

## Security Considerations for Development
- **JWT Secret**: Generate strong secret for development
- **Database Access**: Secure PostgreSQL connection
- **Rate Limiting**: Understand rate limiting behavior
- **CORS Configuration**: Proper cross-origin setup for frontend integration
- **Admin Access**: Understand admin role requirements and setup

## Monitoring and Observability
- **Health Endpoint**: Available at `/health`
- **Metrics**: Prometheus metrics available (port 9100 in dev)
- **Logging**: Structured JSON logging with configurable levels
- **Tracing**: OpenTelemetry support available but optional for development
