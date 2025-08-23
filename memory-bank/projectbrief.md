# Project Brief: Symper One Auth

## Overview
This is **Supabase Auth** (formerly GoTrue), a comprehensive authentication and user management server written in Go. It's a production-ready authentication service that powers Supabase's authentication features.

## Core Purpose
- **Primary Function**: Complete authentication and user management system
- **Target Use Case**: Microservice authentication for modern applications
- **Key Value Proposition**: Production-ready, feature-complete auth service with extensive OAuth provider support

## Key Features
- JWT token issuance and management
- Row Level Security integration with PostgREST
- Comprehensive user management
- Multiple authentication methods:
  - Email/password authentication
  - Magic link authentication
  - Phone number authentication (SMS OTP)
  - Multi-factor authentication (TOTP, Phone, WebAuthn)
  - OAuth providers (Google, Apple, Facebook, GitHub, Discord, etc.)
  - SAML SSO
  - Web3 authentication (Ethereum, Solana)
  - Anonymous authentication

## Architecture
- **Language**: Go 1.23.7
- **Database**: PostgreSQL (required)
- **API Style**: REST API with comprehensive endpoints
- **Deployment**: Docker-ready with multi-stage builds
- **Observability**: OpenTelemetry support for metrics and tracing

## Project Structure
```
symper-one-auth/
├── cmd/                    # CLI commands (serve, migrate, admin, etc.)
├── internal/
│   ├── api/               # REST API handlers and middleware
│   ├── conf/              # Configuration management
│   ├── models/            # Database models (User, Factor, Session, etc.)
│   ├── storage/           # Database connection and utilities
│   ├── crypto/            # Cryptographic utilities
│   ├── mailer/            # Email sending functionality
│   ├── hooks/             # Webhook system
│   └── observability/     # Metrics, tracing, logging
├── migrations/            # Database schema migrations
├── docs/                  # API documentation
└── client/               # Admin client generation
```

## Development Environment
- **Build System**: Make-based with comprehensive targets
- **Development**: Docker Compose with hot reloading
- **Database**: PostgreSQL with automated migrations
- **Testing**: Comprehensive test suite with coverage reporting

## Configuration
- Environment variable based configuration (GOTRUE_ prefix)
- Support for .env files and config directories
- Hot reloading of configuration in development
- Extensive customization options for all features

## Key Dependencies
- Chi router for HTTP routing
- Pop/Soda for database operations and migrations
- JWT libraries for token management
- OAuth2 libraries for external provider integration
- WebAuthn libraries for FIDO2 support
- OpenTelemetry for observability

## Deployment Options
1. **Docker**: Production-ready containers
2. **Binary**: Standalone Go binary
3. **Development**: Docker Compose with hot reloading

## Security Features
- Refresh token rotation
- Rate limiting
- CAPTCHA support
- Password complexity requirements
- Audit logging
- Secure session management
- PKCE support for OAuth flows
