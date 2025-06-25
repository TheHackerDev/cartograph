# Development Guide

This guide covers development workflows for Cartograph, including debugging, testing, and contributing to the project.

## Prerequisites

- Docker and Docker Compose
- Visual Studio Code, Cursor, or Windsurf IDE
- Go extension for your IDE
- Git

## Debug Build Setup

Cartograph includes a specialized debug build configuration that allows you to debug the Go application running inside Docker containers directly from your IDE.

### Starting the Debug Build

The debug build uses a separate Docker Compose configuration that includes debugging capabilities:

```bash
docker compose -f compose-dbg.yaml up --build
```

This command:
- Builds the application with debugging symbols
- Starts the application using [Delve](https://github.com/go-delve/delve), the Go debugger
- Exposes port `40000` for remote debugging connections
- Grants necessary container privileges for debugging

### IDE Configuration

#### Visual Studio Code / Cursor / Windsurf

1. **Create Debug Configuration**

   Create or update `.vscode/launch.json` in your project root:

   ```json
   {
       "version": "0.2.0",
       "configurations": [
           {
               "name": "Connect to Cartograph (Docker)",
               "type": "go",
               "request": "attach",
               "mode": "remote",
               "remotePath": "${workspaceFolder}",
               "port": 40000,
               "host": "127.0.0.1",
               "showLog": true,
               "dlvLoadConfig": {
                   "followPointers": true,
                   "maxVariableRecurse": 3,
                   "maxStringLen": 1024,
                   "maxArrayValues": 64,
                   "maxStructFields": -1
               }
           }
       ]
   }
   ```

2. **Start Debugging**

   - Open any Go file in the project (e.g., `cmd/cartograph/main.go`)
   - Set breakpoints by clicking in the gutter next to line numbers
   - Open the "Run and Debug" view (Ctrl+Shift+D / Cmd+Shift+D)
   - Select "Connect to Cartograph (Docker)" from the dropdown
   - Click the green play button to attach

3. **Debugging Features**

   Once connected, you can:
   - Step through code execution
   - Inspect variables and their values
   - Evaluate expressions in the debug console
   - View call stacks
   - Set conditional breakpoints

### Debug Build Architecture

The debug build differs from the production build in several key ways:

- **Dockerfile**: Uses `build/cartograph/docker/dev/Dockerfile` instead of the production version
- **Debugging Port**: Exposes port `40000` for Delve connections
- **Container Privileges**: Includes `SYS_PTRACE` capability and disables AppArmor for debugging
- **Debug Environment**: Sets `DEBUG=true` for verbose logging

### Troubleshooting Debug Setup

**Connection Issues**
- Ensure the debug build is running: `docker compose -f compose-dbg.yaml ps`
- Check that port 40000 is accessible: `telnet localhost 40000`
- Verify no other services are using port 40000

**Breakpoint Issues**
- Ensure the `remotePath` in `launch.json` matches your workspace structure
- Check that the Go extension is properly installed and enabled
- Verify source code matches the running container version

**Performance Issues**
- The debug build runs slower than production due to debugging overhead
- Consider setting fewer breakpoints for better performance
- Use conditional breakpoints to reduce interruptions

## Development Workflow

### Local Development Setup

1. **Clone the Repository**
   ```bash
   git clone https://github.com/TheHackerDev/cartograph.git
   cd cartograph
   ```

2. **Start Development Environment**
   ```bash
   # For debugging
   docker compose -f compose-dbg.yaml up --build
   
   # For regular development
   docker compose up --build
   ```

3. **Access Services**
   - Web UI: http://localhost or https://localhost
   - API: http://localhost:8000
   - Proxy: http://localhost:8080
   - Database: localhost:5444 (debug) or localhost:5432 (production)

### Code Structure

The project follows a modular architecture:

```
cartograph/
├── cmd/                    # Application entry points
│   ├── cartograph/        # Main application
│   ├── ca-generator/      # Certificate generation utility
│   └── vectorizer/        # Vectorization utility
├── internal/              # Private application code
│   ├── analyzer/          # ML analysis and vectorization
│   ├── apiHunter/         # API detection and analysis
│   ├── config/            # Configuration management
│   ├── crawler/           # Web crawling functionality
│   ├── dns/               # DNS resolution
│   ├── mapper/            # Network mapping and visualization
│   ├── proxy/             # HTTP/HTTPS proxy server
│   ├── shared/            # Shared utilities and types
│   └── webui/             # Web interface
└── docs/                  # Documentation
```

### Database Access

During development, you can access the PostgreSQL database directly:

```bash
# Connect to database (debug build)
docker compose -f compose-dbg.yaml exec postgres psql -U cartograph

# Connect to database (production build)
docker compose exec postgres psql -U cartograph
```

### Adding New Features

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Implement Changes**
   - Follow Go best practices and project conventions
   - Add appropriate tests
   - Update documentation as needed

3. **Test Changes**
   - Use the debug build to test functionality
   - Verify all existing tests pass
   - Test with real web traffic through the proxy

4. **Submit Pull Request**
   - Ensure code follows project style guidelines
   - Include clear description of changes
   - Reference any related issues

### Testing

Cartograph includes a comprehensive testing infrastructure that validates all critical functionality through a modern Go-based test runner. The testing system is designed to ensure reliability and prevent regressions across database connections, certificate generation, and API operations.

#### Test Architecture

The project includes several types of testing executed in three distinct phases:

- **Unit Tests**: Test individual functions and methods (Configuration, Database, Concurrency)
- **Certificate Generation Tests**: Validate production CA generator functionality
- **Integration Tests**: Test complete server startup and full CRUD API operations
- **End-to-End Tests**: Validate complete workflows with real database interactions

#### Running the Complete Test Suite

The recommended way to run tests is using the dedicated test runner, which provides comprehensive validation of all system components:

```bash
# Run the complete test suite (recommended)
docker compose -f compose-test.yaml up --abort-on-container-exit

# Clean up after tests
docker compose -f compose-test.yaml down
```

This command executes the full three-phase test suite and provides detailed output for each testing phase.

#### Test Phases Explained

**Phase 1: Unit Tests**
- Validates database connection initialization and cleanup
- Tests configuration management and concurrent access safety
- Ensures the critical deadlock fix is working correctly
- Tests database monitor startup and proper error handling

**Phase 2: Certificate Generation**
- Uses the production CA generator to create both RSA and ECDSA certificate chains
- Validates that the same certificate generation process used in production works correctly
- Sets up certificates in all required locations for HTTP server and JWT signing
- Tests certificate distribution and runtime setup

**Phase 3: Integration Tests**
- Starts a complete Cartograph server with generated certificates
- Validates server startup and readiness detection
- Performs comprehensive CRUD testing on the targets API:
  - GET (empty configuration)
  - POST (create target) 
  - GET (verify target exists)
  - DELETE (remove target)
  - GET (verify target removed)

#### Understanding Test Output

The test runner provides structured logging with clear phase separation:

```
🚀 Starting Cartograph Test Runner
=== Phase 1: Running Unit Tests ===
✅ Unit tests passed
=== Phase 2: Setting Up Integration Environment ===
✅ Certificates generated
=== Phase 3: Running Integration Tests ===
✅ Integration tests passed
🎉 ALL TESTS PASSED! 🎉
```

**Expected "Error" Messages**: During unit tests, you may see error messages like:
```
error="error waiting for database notification: read tcp ... use of closed network connection"
```
These are **expected and normal** - they indicate that the database monitor is properly detecting connection closures during test cleanup, which validates that our critical deadlock fix is working correctly.

#### Running Individual Test Types

For development and debugging, you can run specific test types:

```bash
# Run only unit tests
go test ./internal/config/

# Run all Go unit tests
go test ./...

# Run only the certificate generation phase
go run ./cmd/cartograph-test/ # (modify to stop after Phase 2)
```

#### Test Infrastructure Benefits

The current test infrastructure provides several advantages over traditional testing approaches:

- **Production-Realistic**: Uses the same CA generator and server startup process as production
- **Zero External Dependencies**: No OpenSSL or other external tools required
- **Comprehensive Coverage**: Tests database, certificates, server startup, and API operations
- **Clean Environment**: Each test run starts with a fresh database and clean state
- **Fast Execution**: Typically completes in under 2 minutes
- **Clear Results**: Structured output makes it easy to identify issues

#### Test Environment Configuration

The test suite uses a dedicated Docker Compose configuration (`compose-test.yaml`) that includes:

- **Isolated Database**: PostgreSQL instance on port 5445 (avoiding conflicts)
- **Test Runner Container**: Go environment with all dependencies
- **Network Isolation**: Dedicated Docker network for test communication
- **Automatic Cleanup**: Resources are automatically cleaned up after tests

#### Troubleshooting Tests

**Test Failures**:
1. Check that Docker is running and has sufficient resources
2. Ensure no other services are using the required ports (5445)
3. Look for specific error messages in the structured test output
4. Verify database connectivity if Phase 1 fails
5. Check certificate generation if Phase 2 fails
6. Validate server startup if Phase 3 fails

**Performance Issues**:
- Tests typically run in 60-120 seconds
- Slower performance may indicate resource constraints
- Monitor Docker resource usage during test execution

**Database Connection Issues**:
- The test runner waits for PostgreSQL to be ready before starting
- Connection errors in Phase 1 may indicate database startup issues
- Check Docker logs if database initialization fails

#### Contributing Test Improvements

When adding new features, ensure they are covered by appropriate tests:

1. **Unit Tests**: Add to `internal/config/config_test.go` for configuration changes
2. **Integration Tests**: Modify the test runner in `cmd/cartograph-test/main.go` for API changes  
3. **Certificate Tests**: Update certificate generation if TLS functionality changes
4. **Documentation**: Update this guide for any testing workflow changes

The test infrastructure is designed to be easily extensible while maintaining comprehensive coverage of all critical functionality.

### Contributing Guidelines

- Follow the existing code style and conventions
- Write clear, descriptive commit messages
- Include tests for new functionality
- Update documentation for user-facing changes
- Ensure the debug build works with your changes

## Advanced Development Topics

### Custom Vectorization Models

The analyzer module supports custom machine learning models for traffic analysis. See the `internal/analyzer/` directory for implementation details.

### Plugin Development

Cartograph uses a plugin-based architecture. New plugins should follow the patterns established in existing modules like `mapper` and `analyzer`.

### Certificate Management

The application automatically generates and manages TLS certificates for MITM proxy functionality. See `internal/shared/http/certificates/` for implementation details.

## Getting Help

- **Issues**: Report bugs and feature requests on [GitHub Issues](https://github.com/TheHackerDev/cartograph/issues)
- **Discussions**: Join conversations on [GitHub Discussions](https://github.com/TheHackerDev/cartograph/discussions)
- **Documentation**: Check the main [README](README.md) for general usage information 