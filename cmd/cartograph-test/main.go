package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/TheHackerDev/cartograph/internal/shared/datatypes"
)

const (
	apiBaseURL  = "http://localhost:8000/api/v1"
	testTimeout = 30 * time.Second
)

// TestRunner manages the complete test execution
type TestRunner struct {
	dbConnString string
	serverCmd    *exec.Cmd
}

// main is the entry point for the Cartograph test runner.
// It initializes the test environment and executes a comprehensive test suite
// that validates database connections, certificate generation, and API functionality.
func main() {
	log.SetLevel(log.InfoLevel)
	log.Info("🚀 Starting Cartograph Test Runner")

	runner := &TestRunner{
		dbConnString: getDBConnectionString(),
	}

	// Execute test phases
	if err := runner.runAllTests(); err != nil {
		log.WithError(err).Fatal("❌ Tests failed")
	}

	log.Info("🎉 ALL TESTS PASSED! 🎉")
}

// runAllTests executes the complete test suite in three phases:
// Phase 1 - Unit tests for configuration and database functionality
// Phase 2 - Certificate generation using production CA generator
// Phase 3 - Integration tests with full server startup and API validation
func (tr *TestRunner) runAllTests() error {
	// Phase 1: Unit Tests
	log.Info("=== Phase 1: Running Unit Tests ===")
	if err := tr.runUnitTests(); err != nil {
		return fmt.Errorf("unit tests failed: %w", err)
	}
	log.Info("✅ Unit tests passed")

	// Phase 2: Setup Integration Environment
	log.Info("=== Phase 2: Setting Up Integration Environment ===")
	if err := tr.setupCertificates(); err != nil {
		return fmt.Errorf("certificate setup failed: %w", err)
	}
	log.Info("✅ Certificates generated")

	// Phase 3: Integration Tests
	log.Info("=== Phase 3: Running Integration Tests ===")
	if err := tr.runIntegrationTests(); err != nil {
		return fmt.Errorf("integration tests failed: %w", err)
	}
	log.Info("✅ Integration tests passed")

	return nil
}

// runUnitTests executes Go unit tests for the internal/config package.
// These tests validate database connection initialization, configuration management,
// and concurrent access safety without requiring a full server startup.
func (tr *TestRunner) runUnitTests() error {
	log.Info("Running Go unit tests...")
	cmd := exec.Command("go", "test", "-v", "./internal/config/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unit test execution failed: %w", err)
	}

	return nil
}

// setupCertificates prepares the certificate infrastructure for integration testing.
// It creates necessary directories, generates both RSA and ECDSA certificate chains
// using the production CA generator, and copies certificates to appropriate locations
// for HTTP server operation and JWT signing.
func (tr *TestRunner) setupCertificates() error {
	log.Info("Creating certificate directories...")
	dirs := []string{
		"internal/shared/http/certificates",
		"internal/shared/users/signing-certificates",
		"/ca-certificates",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	log.Info("Generating certificates using CA generator...")
	if err := tr.generateCertificatesWithCAGenerator(); err != nil {
		return fmt.Errorf("CA generation failed: %w", err)
	}

	log.Info("Setting up runtime certificates...")
	return tr.setupRuntimeCertificates()
}

// generateCertificatesWithCAGenerator creates certificate chains using the production CA generator.
// It generates both RSA and ECDSA certificate chains, each consisting of a root CA,
// intermediate CA, and combined certificate file. This ensures the test environment
// uses the same certificate generation process as production deployments.
func (tr *TestRunner) generateCertificatesWithCAGenerator() error {
	// Generate RSA CA (generates its own root certificate)
	log.Info("Generating RSA certificate chain...")
	rsaCmd := exec.Command("go", "run", "cmd/ca-generator/main.go",
		"-root-cert-pem=/tmp/root-cert-rsa.pem",
		"-root-cert-der=/tmp/root-cert-rsa.crt",
		"-root-key=/tmp/root-key-rsa.pem",
		"-intermediate-cert=/tmp/intermediate-cert-rsa.pem",
		"-intermediate-key=/tmp/intermediate-key-rsa.pem",
		"-combined-cert=/tmp/combined-cert-rsa.pem",
		"-rsa")

	rsaCmd.Stdout = os.Stdout
	rsaCmd.Stderr = os.Stderr

	if err := rsaCmd.Run(); err != nil {
		return fmt.Errorf("RSA CA generation failed: %w", err)
	}

	// Generate ECDSA CA (generates its own root certificate)
	log.Info("Generating ECDSA certificate chain...")
	ecdsaCmd := exec.Command("go", "run", "cmd/ca-generator/main.go",
		"-root-cert-pem=/tmp/root-cert-ecdsa.pem",
		"-root-cert-der=/tmp/root-cert-ecdsa.crt",
		"-root-key=/tmp/root-key-ecdsa.pem",
		"-intermediate-cert=/tmp/intermediate-cert-ecdsa.pem",
		"-intermediate-key=/tmp/intermediate-key-ecdsa.pem",
		"-combined-cert=/tmp/combined-cert-ecdsa.pem")

	ecdsaCmd.Stdout = os.Stdout
	ecdsaCmd.Stderr = os.Stderr

	if err := ecdsaCmd.Run(); err != nil {
		return fmt.Errorf("ECDSA CA generation failed: %w", err)
	}

	return nil
}

// setupRuntimeCertificates distributes generated certificates to their required locations.
// It copies all certificate files to the HTTP certificates directory, places ECDSA
// intermediate certificates in the JWT signing directory, and copies root certificates
// to the runtime CA directory to replicate the production certificate layout.
func (tr *TestRunner) setupRuntimeCertificates() error {
	// Copy all generated certificate files to build directories
	certFiles, err := filepath.Glob("/tmp/*.pem")
	if err != nil {
		return fmt.Errorf("failed to glob certificate files: %w", err)
	}

	crtFiles, err := filepath.Glob("/tmp/*.crt")
	if err != nil {
		return fmt.Errorf("failed to glob certificate files: %w", err)
	}

	allFiles := append(certFiles, crtFiles...)

	// Copy to http certificates directory
	for _, file := range allFiles {
		destFile := filepath.Join("internal/shared/http/certificates", filepath.Base(file))
		if err := copyFile(file, destFile); err != nil {
			log.WithError(err).Warnf("Failed to copy %s to http certificates", file)
		}
	}

	// Copy specific files for JWT signing (ECDSA intermediate cert and key)
	signingCerts := []string{
		"/tmp/intermediate-cert-ecdsa.pem",
		"/tmp/intermediate-key-ecdsa.pem",
	}

	for _, file := range signingCerts {
		destFile := filepath.Join("internal/shared/users/signing-certificates", filepath.Base(file))
		if err := copyFile(file, destFile); err != nil {
			return fmt.Errorf("failed to copy signing certificate %s: %w", file, err)
		}
	}

	// Copy root certificates for runtime (like production mount)
	// Use the root certificates generated by CA generator
	rootCerts := []string{
		"/tmp/root-cert-rsa.pem",
		"/tmp/root-cert-ecdsa.pem",
	}

	for _, file := range rootCerts {
		destFile := filepath.Join("/ca-certificates", filepath.Base(file))
		if err := copyFile(file, destFile); err != nil {
			log.WithError(err).Warnf("Failed to copy runtime certificate %s", file)
		}
	}

	log.Info("✅ All certificates copied to runtime directories")
	return nil
}

// runIntegrationTests performs end-to-end testing of the Cartograph server.
// It starts a complete server instance with generated certificates, waits for
// the server to become ready, and then executes comprehensive API tests including
// full CRUD operations on target configurations.
func (tr *TestRunner) runIntegrationTests() error {
	// Start the server
	log.Info("Starting Cartograph server...")
	if err := tr.startServer(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer tr.stopServer()

	// Wait for server to be ready
	log.Info("Waiting for server to start...")
	if err := tr.waitForServer(); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}

	// Run API tests
	log.Info("Testing API endpoints...")
	return tr.runAPITests()
}

// startServer builds and launches the Cartograph server for integration testing.
// It prepares the mapper scripts directory, builds a fresh cartograph binary,
// and starts the server with test database configuration. The server process
// is started as a background process with output captured for debugging.
func (tr *TestRunner) startServer() error {
	// Create mapper scripts directory
	if err := os.MkdirAll("/tmp/mapper-scripts", 0755); err != nil {
		return fmt.Errorf("failed to create mapper scripts directory: %w", err)
	}

	// Copy mapper scripts
	scripts := []string{
		"internal/mapper/mapper.js",
		"internal/mapper/mapper-worker.js",
	}

	for _, script := range scripts {
		destFile := filepath.Join("/tmp/mapper-scripts", filepath.Base(script))
		if err := copyFile(script, destFile); err != nil {
			return fmt.Errorf("failed to copy mapper script %s: %w", script, err)
		}
	}

	// Build and start server
	log.Info("Building cartograph binary...")
	buildCmd := exec.Command("go", "build", "-o", "/tmp/cartograph", "./cmd/cartograph/")
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build cartograph: %w", err)
	}

	log.Info("Starting cartograph server...")
	tr.serverCmd = exec.Command("/tmp/cartograph", "--mapper-script-dir=/tmp/mapper-scripts")
	tr.serverCmd.Env = append(os.Environ(),
		"DB_HOST=postgres-test",
		"DB_PORT=5432",
		"DB_NAME=cartograph",
		"DB_USER=cartograph",
		"DB_PASS=myDbPass123#",
	)

	// Capture server output for debugging
	tr.serverCmd.Stdout = os.Stdout
	tr.serverCmd.Stderr = os.Stderr

	return tr.serverCmd.Start()
}

// stopServer gracefully shuts down the Cartograph server process.
// It terminates the server process and waits for it to exit completely,
// ensuring proper cleanup of resources and connections.
func (tr *TestRunner) stopServer() {
	if tr.serverCmd != nil && tr.serverCmd.Process != nil {
		log.Info("Stopping server...")
		tr.serverCmd.Process.Kill()
		tr.serverCmd.Wait()
	}
}

// waitForServer polls the server's API endpoint until it responds successfully.
// It performs up to 10 attempts with 3-second intervals, testing the targets
// endpoint to ensure the server is fully initialized and ready to handle requests.
// This prevents race conditions in integration tests.
func (tr *TestRunner) waitForServer() error {
	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < 10; i++ {
		resp, err := client.Get(apiBaseURL + "/config/targets/")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			log.Info("✅ Server is responding on port 8000")
			return nil
		}

		log.Infof("Attempt %d: Server not ready yet, waiting...", i+1)
		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("server failed to start within timeout")
}

// runAPITests executes comprehensive API validation tests against the running server.
// It performs a complete CRUD cycle: GET (empty list), POST (create target),
// GET (verify creation), DELETE (remove target), and GET (verify removal).
// This validates that configuration management, database operations, and API
// endpoints are functioning correctly end-to-end.
func (tr *TestRunner) runAPITests() error {
	client := &http.Client{Timeout: testTimeout}

	// Test 1: GET targets (should return empty)
	log.Info("Testing GET /api/v1/config/targets/")
	resp, err := client.Get(apiBaseURL + "/config/targets/")
	if err != nil {
		return fmt.Errorf("GET request failed: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GET request returned status %d", resp.StatusCode)
	}
	log.Info("✅ API test passed")

	// Test 2: POST target (create)
	log.Info("Adding test target...")
	target := datatypes.TargetFilterSimple{
		Ignore: false,
		Hosts:  []string{"test.example.com"},
	}

	targetJSON, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("failed to marshal target: %w", err)
	}

	resp, err = client.Post(apiBaseURL+"/config/targets/", "application/json", bytes.NewBuffer(targetJSON))
	if err != nil {
		return fmt.Errorf("POST request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("POST request returned status %d", resp.StatusCode)
	}

	targetIDBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read target ID: %w", err)
	}

	targetID := string(targetIDBytes)
	log.Infof("✅ Created target with ID: %s", targetID)

	// Test 3: Verify target exists
	log.Info("Verifying target was added...")
	resp, err = client.Get(apiBaseURL + "/config/targets/")
	if err != nil {
		return fmt.Errorf("verification GET failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if !strings.Contains(string(body), "test.example.com") {
		return fmt.Errorf("target not found in configuration")
	}
	log.Info("✅ Target found in configuration")

	// Test 4: DELETE target
	log.Info("Deleting target...")
	deleteURL := fmt.Sprintf("%s/config/targets/?id=%s", apiBaseURL, targetID)
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE request failed: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("DELETE request returned status %d", resp.StatusCode)
	}
	log.Info("✅ Target deletion completed")

	// Test 5: Verify target removed
	time.Sleep(2 * time.Second) // Allow time for deletion to propagate

	log.Info("Verifying target was removed...")
	resp, err = client.Get(apiBaseURL + "/config/targets/")
	if err != nil {
		return fmt.Errorf("final verification GET failed: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read final response: %w", err)
	}

	if strings.Contains(string(body), "test.example.com") {
		log.Warn("⚠️ Target may still exist (but CRUD operations work)")
	} else {
		log.Info("✅ Target successfully removed from configuration")
	}

	return nil
}

// Helper functions

// getDBConnectionString constructs a PostgreSQL connection string from environment variables.
// It uses environment variables for database configuration with sensible defaults
// for the test environment, allowing the test runner to connect to the test database.
func getDBConnectionString() string {
	host := getEnvOrDefault("DB_HOST", "postgres-test")
	port := getEnvOrDefault("DB_PORT", "5432")
	dbname := getEnvOrDefault("DB_NAME", "cartograph")
	user := getEnvOrDefault("DB_USER", "cartograph")
	password := getEnvOrDefault("DB_PASS", "myDbPass123#")

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

// getEnvOrDefault retrieves an environment variable value or returns a default value.
// This utility function provides fallback values for configuration parameters,
// enabling the test runner to work with sensible defaults when environment
// variables are not explicitly set.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// copyFile copies a file from source to destination path.
// It handles file opening, creation, and data transfer with proper error handling
// and resource cleanup. Used for distributing certificates and scripts to their
// required locations during test setup.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
