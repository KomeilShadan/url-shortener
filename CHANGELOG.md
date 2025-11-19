# Changelog

All notable changes to the Janus project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2024-11-19

### 🎉 Major Release - Complete Project Refactor

This is a major release with significant architectural improvements, production-grade enhancements, and namespace migration to `janus`.

### Added

#### Repository Pattern
- ✅ Implemented repository layer with clean interfaces
- ✅ `LinkRepository` interface for database operations
- ✅ `CacheRepository` interface for Redis operations
- ✅ Repository container with singleton pattern
- ✅ MongoDB repository implementation with proper error handling
- ✅ Redis cache repository implementation

#### Service Layer
- ✅ In-memory link cache with time-bucketed expiration
- ✅ Configurable TTL, cleanup interval, and max size
- ✅ Thread-safe cache operations with RWMutex
- ✅ Automatic cleanup goroutine for expired entries
- ✅ `GenerateShortLinkHash` function with backward compatibility

#### Logging Enhancements
- ✅ GELF (Graylog Extended Log Format) support
- ✅ Graylog integration for centralized logging
- ✅ Enhanced logger with `Info`, `Debug`, and `Warn` functions
- ✅ Improved syslog implementation
- ✅ Comprehensive logging throughout the application
- ✅ Structured logging with context and metadata

#### Configuration
- ✅ MongoDB connection pool configuration (min/max pool size, timeouts)
- ✅ Redis connection configuration (pool size, timeouts, idle connections)
- ✅ Graylog configuration (facility, protocol, address)
- ✅ Fallback URL configuration
- ✅ Environment-based configuration with sensible defaults

#### API Improvements
- ✅ `GetLink` endpoint for retrieving links without redirect
- ✅ Request timeout handling (5 seconds for DB operations)
- ✅ Context deadline detection and proper error responses
- ✅ Enhanced validation with better error messages
- ✅ Cache-first strategy for link resolution
- ✅ Proper HTTP status codes (408 for timeouts, 404 for not found)

#### Middleware Enhancements
- ✅ Improved authentication middleware with development mode support
- ✅ Enhanced rate limiting with Redis
- ✅ Rate limit headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`)
- ✅ Better error messages and logging

#### Server & Infrastructure
- ✅ Graceful shutdown with signal handling (SIGINT, SIGTERM)
- ✅ HTTP server with production-grade timeouts (read, write, idle)
- ✅ Custom panic recovery middleware with detailed logging
- ✅ Server error channel for startup error detection
- ✅ Health check endpoints (`/s/ping`, `/s/health`)
- ✅ Detailed startup logging with all configuration parameters

#### Main Application
- ✅ Global panic recovery with stack traces
- ✅ Configuration validation at startup
- ✅ MongoDB connection retry logic with exponential backoff
- ✅ Redis connection testing with ping
- ✅ Comprehensive startup/shutdown logging
- ✅ Proper resource cleanup with context timeouts

#### Deployment & DevOps
- ✅ Enhanced CLI script (`janus`) with color output and help menu
- ✅ Docker Compose configuration with health checks
- ✅ Comprehensive `.env.example` with all configuration options

### Changed

#### Architecture
- 🔄 Refactored handlers to use repository pattern
- 🔄 Removed direct MongoDB/Redis client injection
- 🔄 Implemented dependency injection through repositories
- 🔄 Separated concerns: handlers → services → repositories
- 🔄 Improved error handling throughout the stack

#### Database Operations
- 🔄 All DB operations now use context with timeouts
- 🔄 Configurable connection pool settings
- 🔄 Better error messages with context
- 🔄 Proper connection lifecycle management

#### Logging
- 🔄 Standardized logging format across the application
- 🔄 Added structured logging with metadata
- 🔄 Improved log levels (Debug, Info, Warn, Error)
- 🔄 Better error context in logs

#### Configuration
- 🔄 MongoDB: Hardcoded values → Environment variables
- 🔄 Redis: Added `Addr` field for unified address configuration
- 🔄 Enhanced with timeout and pool size configurations
- 🔄 Better default values for production use

### Fixed

- 🐛 Fixed logger interface usage in main.go
- 🐛 Fixed MongoDB connection pool configuration
- 🐛 Fixed Redis address configuration
- 🐛 Fixed error handling in handlers
- 🐛 Fixed cache invalidation on link updates
- 🐛 Fixed graceful shutdown issues
- 🐛 Fixed panic recovery logging

### Security

- 🔒 Added request timeouts to prevent hanging connections
- 🔒 Implemented proper authentication checks
- 🔒 Added rate limiting with Redis
- 🔒 Improved input validation
- 🔒 Added self-domain protection
- 🔒 Secure error messages (no sensitive data leakage)

### Performance

- ⚡ Multi-layer caching (in-memory + Redis)
- ⚡ Connection pooling for MongoDB and Redis
- ⚡ Configurable pool sizes and timeouts
- ⚡ Efficient cache cleanup with time-bucketing
- ⚡ Reduced database queries with cache-first strategy

### Documentation

- 📝 Comprehensive `.env.example` with all options
- 📝 Enhanced CLI help menu
- 📝 Improved code comments
- 📝 Better error messages for developers

### Migration Guide

If upgrading from v1.x:

1. Update Docker container names in scripts
2. Update `.env` file with new configuration options
3. Run `go mod tidy` to update dependencies
4. Rebuild Docker images

### Technical Debt Addressed

- ✅ Removed hardcoded configuration values
- ✅ Eliminated direct database client usage in handlers
- ✅ Improved error handling consistency
- ✅ Added proper logging throughout
- ✅ Implemented graceful shutdown
- ✅ Added comprehensive validation

---

## [1.0.0] - Previous Release

Initial release with basic URL shortener functionality.

