package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheHackerDev/cartograph/internal/shared/datatypes"
)

func resetFlags() {
	// Reset the flag package to avoid redefinition panics
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestConfigInitialization(t *testing.T) {
	// Test that Config can be created without panics
	t.Run("Config_creation_succeeds", func(t *testing.T) {
		// Reset flags to avoid redefinition
		resetFlags()

		// Set required environment variables for testing
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Mock command line arguments
		os.Args = []string{"test", "--mapper-script-dir=/tmp/test-scripts"}

		// Create config - should not panic
		cfg, err := NewConfig()
		require.NoError(t, err, "Config creation should not fail")
		require.NotNil(t, cfg, "Config should not be nil")
		defer cfg.Close()

		// Test that database connections are properly initialized
		assert.NotNil(t, cfg.listenDbConn, "Listen database connection should be initialized")
	})

	t.Run("Database_connection_cleanup", func(t *testing.T) {
		// Reset flags to avoid redefinition
		resetFlags()

		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{"test", "--mapper-script-dir=/tmp/test-scripts"}

		cfg, err := NewConfig()
		require.NoError(t, err, "Config creation should not fail")
		require.NotNil(t, cfg)

		// Close should work multiple times without error
		err1 := cfg.Close()
		err2 := cfg.Close()

		assert.NoError(t, err1, "First close should succeed")
		assert.NoError(t, err2, "Second close should not fail")
	})
}

func TestConfigurationUpdates(t *testing.T) {
	t.Run("Add_and_remove_targets", func(t *testing.T) {
		// Reset flags to avoid redefinition
		resetFlags()

		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{"test", "--mapper-script-dir=/tmp/test-scripts"}

		cfg, err := NewConfig()
		require.NoError(t, err, "Config creation should not fail")
		require.NotNil(t, cfg)
		defer cfg.Close()

		// Test adding targets
		testTarget := &datatypes.TargetFilterSimple{
			Ignore: false,
			Hosts:  []string{"test1.example.com", "test2.example.com"},
		}
		targetId, err := cfg.addTargetOrIgnored(testTarget)
		assert.NoError(t, err, "Adding target should succeed")
		assert.NotEmpty(t, targetId, "Target ID should not be empty")

		// Test that target appears in configuration
		allTargets := cfg.GetTargetsAndIgnoredAll()
		target, found := allTargets[targetId]
		assert.True(t, found, "Added target should be found in configuration")
		if found {
			assert.False(t, target.IsIgnore, "Target ignore flag should be false")
			// Check that hosts are present (the structure is different in TargetIgnoreSimple)
			assert.NotEmpty(t, target.Hosts, "Target should have hosts")
		}

		// Test removing targets
		err = cfg.deleteTargetOrIgnored(targetId)
		assert.NoError(t, err, "Deleting target should succeed")

		// Test that target no longer appears
		allTargetsAfterDelete := cfg.GetTargetsAndIgnoredAll()
		_, found = allTargetsAfterDelete[targetId]
		assert.False(t, found, "Deleted target should not be found in configuration")
	})

	t.Run("Concurrent_access_safety", func(t *testing.T) {
		// Reset flags to avoid redefinition
		resetFlags()

		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{"test", "--mapper-script-dir=/tmp/test-scripts"}

		cfg, err := NewConfig()
		require.NoError(t, err, "Config creation should not fail")
		require.NotNil(t, cfg)
		defer cfg.Close()

		// Test concurrent reads and writes
		done := make(chan bool, 3)

		// Concurrent reader 1
		go func() {
			for i := 0; i < 10; i++ {
				_ = cfg.GetTargetsAndIgnoredAll()
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()

		// Concurrent reader 2
		go func() {
			for i := 0; i < 10; i++ {
				_ = cfg.GetTargetsAndIgnoredAll()
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()

		// Concurrent writer
		go func() {
			for i := 0; i < 5; i++ {
				testTarget := &datatypes.TargetFilterSimple{
					Ignore: false,
					Hosts:  []string{"concurrent-test.example.com"},
				}
				targetId, err := cfg.addTargetOrIgnored(testTarget)
				if err == nil {
					cfg.deleteTargetOrIgnored(targetId)
				}
				time.Sleep(2 * time.Millisecond)
			}
			done <- true
		}()

		// Wait for all goroutines with timeout
		timeout := time.After(5 * time.Second)
		for i := 0; i < 3; i++ {
			select {
			case <-done:
				// Good, one goroutine finished
			case <-timeout:
				t.Fatal("Concurrent access test timed out - possible deadlock")
			}
		}
	})
}

func TestDatabaseMonitorStartup(t *testing.T) {
	t.Run("Database_monitor_starts", func(t *testing.T) {
		// Reset flags to avoid redefinition
		resetFlags()

		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{"test", "--mapper-script-dir=/tmp/test-scripts"}

		cfg, err := NewConfig()
		require.NoError(t, err, "Config creation should not fail")
		require.NotNil(t, cfg)
		defer cfg.Close()

		// Give database monitor time to start
		time.Sleep(100 * time.Millisecond)

		// Test that we can still interact with configuration
		// (this would hang if database monitor had a deadlock)
		targets := cfg.GetTargetsAndIgnoredAll()
		assert.NotNil(t, targets, "Should be able to get targets without hanging")
	})
}
