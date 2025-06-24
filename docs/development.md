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

The project includes several types of testing:

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete workflows

Run tests using:
```bash
go test ./...
```

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