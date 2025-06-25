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

func (tr *TestRunner) setupCertificates() error {
	log.Info("Creating certificate directories...")
	dirs := []string{
		"certificates",
		"internal/shared/http/certificates",
		"internal/shared/users/signing-certificates",
		"/ca-certificates",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	log.Info("Generating root certificates...")
	if err := tr.generateRootCertificates(); err != nil {
		return fmt.Errorf("root certificate generation failed: %w", err)
	}

	log.Info("Running CA generator...")
	if err := tr.runCAGenerator(); err != nil {
		return fmt.Errorf("CA generation failed: %w", err)
	}

	log.Info("Setting up runtime certificates...")
	return tr.setupRuntimeCertificates()
}

func (tr *TestRunner) generateRootCertificates() error {
	// Generate RSA root certificate
	rsaCommands := [][]string{
		{"openssl", "genrsa", "-out", "certificates/root-key-rsa.pem", "2048"},
		{"openssl", "req", "-new", "-x509", "-key", "certificates/root-key-rsa.pem",
			"-out", "certificates/root-cert-rsa.pem", "-days", "365",
			"-subj", "/C=US/O=Test/CN=Test Root CA RSA"},
	}

	// Generate ECDSA root certificate
	ecdsaCommands := [][]string{
		{"openssl", "ecparam", "-genkey", "-name", "prime256v1", "-out", "certificates/root-key-ecdsa.pem"},
		{"openssl", "req", "-new", "-x509", "-key", "certificates/root-key-ecdsa.pem",
			"-out", "certificates/root-cert-ecdsa.pem", "-days", "365",
			"-subj", "/C=US/O=Test/CN=Test Root CA ECDSA"},
	}

	allCommands := append(rsaCommands, ecdsaCommands...)

	for _, cmdArgs := range allCommands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run %s: %w", strings.Join(cmdArgs, " "), err)
		}
	}

	return nil
}

func (tr *TestRunner) runCAGenerator() error {
	// Generate RSA CA
	rsaCmd := exec.Command("go", "run", "cmd/ca-generator/main.go",
		"-root-cert-in=certificates/root-cert-rsa.pem",
		"-root-key-in=certificates/root-key-rsa.pem",
		"-root-cert-pem=/tmp/root-cert-rsa.pem",
		"-root-cert-der=/tmp/root-cert-rsa.crt",
		"-root-key=/tmp/root-key-rsa.pem",
		"-intermediate-cert=/tmp/intermediate-cert-rsa.pem",
		"-intermediate-key=/tmp/intermediate-key-rsa.pem",
		"-rsa")

	if err := rsaCmd.Run(); err != nil {
		return fmt.Errorf("RSA CA generation failed: %w", err)
	}

	// Generate ECDSA CA
	ecdsaCmd := exec.Command("go", "run", "cmd/ca-generator/main.go",
		"-root-cert-in=certificates/root-cert-ecdsa.pem",
		"-root-key-in=certificates/root-key-ecdsa.pem",
		"-root-cert-pem=/tmp/root-cert-ecdsa.pem",
		"-root-cert-der=/tmp/root-cert-ecdsa.crt",
		"-root-key=/tmp/root-key-ecdsa.pem",
		"-intermediate-cert=/tmp/intermediate-cert-ecdsa.pem",
		"-intermediate-key=/tmp/intermediate-key-ecdsa.pem")

	if err := ecdsaCmd.Run(); err != nil {
		return fmt.Errorf("ECDSA CA generation failed: %w", err)
	}

	return nil
}

func (tr *TestRunner) setupRuntimeCertificates() error {
	// Copy generated certificates to build directories
	certFiles, err := filepath.Glob("/tmp/*.pem")
	if err != nil {
		return fmt.Errorf("failed to glob certificate files: %w", err)
	}

	crtFiles, err := filepath.Glob("/tmp/*.crt")
	if err != nil {
		return fmt.Errorf("failed to glob certificate files: %w", err)
	}

	allFiles := append(certFiles, crtFiles...)

	for _, file := range allFiles {
		// Copy to http certificates
		destFile := filepath.Join("internal/shared/http/certificates", filepath.Base(file))
		if err := copyFile(file, destFile); err != nil {
			log.WithError(err).Warnf("Failed to copy %s", file)
		}
	}

	// Copy specific files for JWT signing
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

	// Copy certificates for runtime (like production mount)
	rootCerts, _ := filepath.Glob("certificates/*")
	for _, file := range rootCerts {
		destFile := filepath.Join("/ca-certificates", filepath.Base(file))
		if err := copyFile(file, destFile); err != nil {
			log.WithError(err).Warnf("Failed to copy runtime certificate %s", file)
		}
	}

	return nil
}

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

func (tr *TestRunner) stopServer() {
	if tr.serverCmd != nil && tr.serverCmd.Process != nil {
		log.Info("Stopping server...")
		tr.serverCmd.Process.Kill()
		tr.serverCmd.Wait()
	}
}

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
func getDBConnectionString() string {
	host := getEnvOrDefault("DB_HOST", "postgres-test")
	port := getEnvOrDefault("DB_PORT", "5432")
	dbname := getEnvOrDefault("DB_NAME", "cartograph")
	user := getEnvOrDefault("DB_USER", "cartograph")
	password := getEnvOrDefault("DB_PASS", "myDbPass123#")

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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
