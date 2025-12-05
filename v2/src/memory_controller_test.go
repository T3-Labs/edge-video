package main

import (
	"testing"
	"time"
)

func TestMemoryControllerBasic(t *testing.T) {
	config := MemoryControllerConfig{
		Enabled:          true,
		MaxMemoryMB:      1024,
		WarningPercent:   60.0,
		CriticalPercent:  75.0,
		EmergencyPercent: 85.0,
		CheckInterval:    1 * time.Second,
		GCTriggerPercent: 70.0,
	}

	// Validate config
	if err := ValidateMemoryControllerConfig(config); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// Create controller
	mc := NewMemoryController(config)
	if mc == nil {
		t.Fatal("NewMemoryController returned nil")
	}

	// Check initial level is Normal
	if mc.GetCurrentLevel() != MemoryNormal {
		t.Errorf("Expected initial level to be Normal, got %v", mc.GetCurrentLevel())
	}

	// Start controller
	mc.Start()
	defer mc.Stop()

	// Wait for first check to complete
	time.Sleep(2 * time.Second)

	// Force check
	mc.checkMemory()

	// Get stats
	stats := mc.GetStats()
	t.Logf("Memory stats: Level=%s, Alloc=%d MB, Sys=%d MB, Usage=%.1f%%, NumGC=%d",
		stats.Level, stats.AllocMB, stats.SysMB, stats.UsagePercent, stats.NumGC)

	// Verify stats were populated
	if stats.SysMB == 0 {
		t.Error("Expected non-zero Sys memory")
	}
}

func TestMemoryControllerThrottling(t *testing.T) {
	config := MemoryControllerConfig{
		Enabled:          true,
		MaxMemoryMB:      1024,
		WarningPercent:   60.0,
		CriticalPercent:  75.0,
		EmergencyPercent: 85.0,
		CheckInterval:    1 * time.Second,
		GCTriggerPercent: 70.0,
	}

	mc := NewMemoryController(config)

	// Test throttling at different levels
	testCases := []struct {
		level         MemoryLevel
		expectedDelay int
	}{
		{MemoryNormal, 0},
		{MemoryWarning, 100},
		{MemoryCritical, 500},
		{MemoryEmergency, 2000},
	}

	for _, tc := range testCases {
		mc.currentLevel = tc.level
		shouldThrottle, delayMs := mc.ShouldThrottle("test-cam")

		if tc.expectedDelay == 0 {
			if shouldThrottle {
				t.Errorf("Level %s: Expected no throttling, but got throttle=true", tc.level)
			}
		} else {
			if !shouldThrottle {
				t.Errorf("Level %s: Expected throttling, but got throttle=false", tc.level)
			}
			if delayMs != tc.expectedDelay {
				t.Errorf("Level %s: Expected delay %d ms, got %d ms", tc.level, tc.expectedDelay, delayMs)
			}
		}
	}
}

func TestMemoryControllerCallbacks(t *testing.T) {
	config := MemoryControllerConfig{
		Enabled:          true,
		MaxMemoryMB:      1024,
		WarningPercent:   60.0,
		CriticalPercent:  75.0,
		EmergencyPercent: 85.0,
		CheckInterval:    1 * time.Second,
		GCTriggerPercent: 70.0,
	}

	mc := NewMemoryController(config)

	// Register callback
	mc.RegisterCallback(MemoryWarning, func(stats MemoryStats) {
		t.Logf("Callback triggered for WARNING level: %.1f%%", stats.UsagePercent)
	})

	// Verify callback was registered
	if len(mc.callbacks[MemoryWarning]) != 1 {
		t.Error("Callback was not registered")
	}

	t.Logf("Callback registered successfully")
}

func TestValidateMemoryControllerConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      MemoryControllerConfig
		shouldError bool
	}{
		{
			name: "Valid config",
			config: MemoryControllerConfig{
				Enabled:          true,
				MaxMemoryMB:      2048,
				WarningPercent:   60.0,
				CriticalPercent:  75.0,
				EmergencyPercent: 85.0,
				CheckInterval:    5 * time.Second,
				GCTriggerPercent: 70.0,
			},
			shouldError: false,
		},
		{
			name: "Disabled config (skip validation)",
			config: MemoryControllerConfig{
				Enabled: false,
			},
			shouldError: false,
		},
		{
			name: "Invalid MaxMemoryMB",
			config: MemoryControllerConfig{
				Enabled:          true,
				MaxMemoryMB:      0,
				WarningPercent:   60.0,
				CriticalPercent:  75.0,
				EmergencyPercent: 85.0,
				CheckInterval:    5 * time.Second,
				GCTriggerPercent: 70.0,
			},
			shouldError: true,
		},
		{
			name: "Invalid thresholds order",
			config: MemoryControllerConfig{
				Enabled:          true,
				MaxMemoryMB:      2048,
				WarningPercent:   80.0, // Higher than critical
				CriticalPercent:  75.0,
				EmergencyPercent: 85.0,
				CheckInterval:    5 * time.Second,
				GCTriggerPercent: 70.0,
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMemoryControllerConfig(tc.config)
			if tc.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestMemoryLevelString(t *testing.T) {
	testCases := []struct {
		level    MemoryLevel
		expected string
	}{
		{MemoryNormal, "NORMAL"},
		{MemoryWarning, "WARNING"},
		{MemoryCritical, "CRITICAL"},
		{MemoryEmergency, "EMERGENCY"},
	}

	for _, tc := range testCases {
		result := tc.level.String()
		if result != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, result)
		}
	}
}
