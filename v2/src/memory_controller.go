// memory_controller.go
// Memory Controller - Automatic memory management with 4 levels
// Based on V1.6 implementation, simplified for V2

package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// MemoryLevel represents the current memory pressure level
type MemoryLevel int

const (
	MemoryNormal MemoryLevel = iota
	MemoryWarning
	MemoryCritical
	MemoryEmergency
)

// String returns the string representation of memory level
func (m MemoryLevel) String() string {
	switch m {
	case MemoryNormal:
		return "NORMAL"
	case MemoryWarning:
		return "WARNING"
	case MemoryCritical:
		return "CRITICAL"
	case MemoryEmergency:
		return "EMERGENCY"
	default:
		return "UNKNOWN"
	}
}

// MemoryStats holds current memory statistics
type MemoryStats struct {
	AllocMB      uint64
	TotalAllocMB uint64
	SysMB        uint64
	NumGC        uint32
	Level        MemoryLevel
	UsagePercent float64
	Timestamp    time.Time
}

// ThrottleState holds throttling information for a camera
type ThrottleState struct {
	CameraID     string
	DelayMs      int
	LastThrottle time.Time
}

// MemoryController manages memory levels and triggers GC when needed
type MemoryController struct {
	mu           sync.RWMutex
	config       MemoryControllerConfig
	currentLevel MemoryLevel
	stats        MemoryStats
	callbacks    map[MemoryLevel][]func(MemoryStats)
	gcInProgress bool
	lastGC       time.Time
	throttleMap  map[string]*ThrottleState
	ctx          context.Context
	cancel       context.CancelFunc
}

// MemoryControllerConfig holds configuration for memory controller
type MemoryControllerConfig struct {
	Enabled          bool          `yaml:"enabled"`
	MaxMemoryMB      uint64        `yaml:"max_memory_mb"`
	WarningPercent   float64       `yaml:"warning_percent"`
	CriticalPercent  float64       `yaml:"critical_percent"`
	EmergencyPercent float64       `yaml:"emergency_percent"`
	CheckInterval    time.Duration `yaml:"check_interval"`
	GCTriggerPercent float64       `yaml:"gc_trigger_percent"`
}

// NewMemoryController creates a new memory controller instance
func NewMemoryController(config MemoryControllerConfig) *MemoryController {
	ctx, cancel := context.WithCancel(context.Background())

	mc := &MemoryController{
		config:       config,
		currentLevel: MemoryNormal,
		callbacks:    make(map[MemoryLevel][]func(MemoryStats)),
		throttleMap:  make(map[string]*ThrottleState),
		lastGC:       time.Now(),
		ctx:          ctx,
		cancel:       cancel,
	}

	return mc
}

// Start begins the memory monitoring loop
func (mc *MemoryController) Start() {
	log.Printf("[MemCtrl] Memory Controller INICIADO (max: %d MB, warning: %.0f%%, critical: %.0f%%, emergency: %.0f%%)",
		mc.config.MaxMemoryMB,
		mc.config.WarningPercent,
		mc.config.CriticalPercent,
		mc.config.EmergencyPercent)

	go mc.monitorLoop()
}

// Stop stops the memory monitoring loop
func (mc *MemoryController) Stop() {
	log.Println("[MemCtrl] Memory Controller PARANDO...")
	mc.cancel()
}

// monitorLoop continuously monitors memory usage
func (mc *MemoryController) monitorLoop() {
	ticker := time.NewTicker(mc.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			log.Println("[MemCtrl] Monitor loop FINALIZADO")
			return
		case <-ticker.C:
			mc.checkMemory()
		}
	}
}

// checkMemory checks current memory usage and updates level
func (mc *MemoryController) checkMemory() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Update stats
	mc.mu.Lock()
	mc.stats = MemoryStats{
		AllocMB:      m.Alloc / 1024 / 1024,
		TotalAllocMB: m.TotalAlloc / 1024 / 1024,
		SysMB:        m.Sys / 1024 / 1024,
		NumGC:        m.NumGC,
		Timestamp:    time.Now(),
	}

	// Calculate usage percentage
	if mc.config.MaxMemoryMB > 0 {
		mc.stats.UsagePercent = (float64(mc.stats.AllocMB) / float64(mc.config.MaxMemoryMB)) * 100.0
	}

	// Determine memory level
	oldLevel := mc.currentLevel
	newLevel := mc.calculateLevel(mc.stats.UsagePercent)
	mc.currentLevel = newLevel
	mc.stats.Level = newLevel
	mc.mu.Unlock()

	// Trigger GC if needed
	if mc.shouldTriggerGC() {
		mc.triggerGC()
	}

	// Log level changes
	if oldLevel != newLevel {
		log.Printf("[MemCtrl] Memory level mudou: %s → %s (%.1f%%, %d MB / %d MB)",
			oldLevel, newLevel, mc.stats.UsagePercent, mc.stats.AllocMB, mc.config.MaxMemoryMB)

		// Execute callbacks
		mc.executeCallbacks(newLevel, mc.stats)
	}
}

// calculateLevel determines memory level based on usage percentage
func (mc *MemoryController) calculateLevel(usagePercent float64) MemoryLevel {
	if usagePercent >= mc.config.EmergencyPercent {
		return MemoryEmergency
	} else if usagePercent >= mc.config.CriticalPercent {
		return MemoryCritical
	} else if usagePercent >= mc.config.WarningPercent {
		return MemoryWarning
	}
	return MemoryNormal
}

// shouldTriggerGC checks if GC should be triggered
func (mc *MemoryController) shouldTriggerGC() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Don't trigger if GC is already in progress
	if mc.gcInProgress {
		return false
	}

	// Don't trigger if we just ran GC (debounce)
	if time.Since(mc.lastGC) < 5*time.Second {
		return false
	}

	// Trigger if above GC threshold
	return mc.stats.UsagePercent >= mc.config.GCTriggerPercent
}

// triggerGC manually triggers garbage collection
func (mc *MemoryController) triggerGC() {
	mc.mu.Lock()
	mc.gcInProgress = true
	mc.mu.Unlock()

	beforeMB := mc.stats.AllocMB
	start := time.Now()

	log.Printf("[MemCtrl] Triggering GC (usage: %.1f%%, %d MB)...", mc.stats.UsagePercent, beforeMB)

	// Run GC
	runtime.GC()
	debug.FreeOSMemory()

	duration := time.Since(start)

	// Update stats after GC
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	afterMB := m.Alloc / 1024 / 1024
	freedMB := int64(beforeMB) - int64(afterMB)

	log.Printf("[MemCtrl] GC completo em %v (liberado: %d MB, atual: %d MB)",
		duration, freedMB, afterMB)

	mc.mu.Lock()
	mc.gcInProgress = false
	mc.lastGC = time.Now()
	mc.mu.Unlock()
}

// GetCurrentLevel returns the current memory level (thread-safe)
func (mc *MemoryController) GetCurrentLevel() MemoryLevel {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.currentLevel
}

// GetStats returns current memory statistics (thread-safe)
func (mc *MemoryController) GetStats() MemoryStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.stats
}

// ShouldThrottle checks if a camera should be throttled based on memory level
func (mc *MemoryController) ShouldThrottle(cameraID string) (bool, int) {
	mc.mu.RLock()
	level := mc.currentLevel
	mc.mu.RUnlock()

	switch level {
	case MemoryNormal:
		return false, 0
	case MemoryWarning:
		return true, 100 // 100ms delay
	case MemoryCritical:
		return true, 500 // 500ms delay
	case MemoryEmergency:
		return true, 2000 // 2s delay
	default:
		return false, 0
	}
}

// ApplyThrottle applies throttling delay if needed
func (mc *MemoryController) ApplyThrottle(cameraID string) {
	shouldThrottle, delayMs := mc.ShouldThrottle(cameraID)
	if !shouldThrottle {
		return
	}

	// Update throttle state
	mc.mu.Lock()
	mc.throttleMap[cameraID] = &ThrottleState{
		CameraID:     cameraID,
		DelayMs:      delayMs,
		LastThrottle: time.Now(),
	}
	mc.mu.Unlock()

	// Apply delay
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
}

// RegisterCallback registers a callback for a specific memory level
func (mc *MemoryController) RegisterCallback(level MemoryLevel, callback func(MemoryStats)) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.callbacks[level] = append(mc.callbacks[level], callback)
	log.Printf("[MemCtrl] Callback registrado para level: %s", level)
}

// executeCallbacks executes all callbacks for a given level
func (mc *MemoryController) executeCallbacks(level MemoryLevel, stats MemoryStats) {
	mc.mu.RLock()
	callbacks := mc.callbacks[level]
	mc.mu.RUnlock()

	for _, callback := range callbacks {
		go callback(stats)
	}
}

// GetThrottleStats returns throttle statistics for all cameras
func (mc *MemoryController) GetThrottleStats() map[string]*ThrottleState {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]*ThrottleState)
	for k, v := range mc.throttleMap {
		result[k] = &ThrottleState{
			CameraID:     v.CameraID,
			DelayMs:      v.DelayMs,
			LastThrottle: v.LastThrottle,
		}
	}
	return result
}

// ForceGC forces an immediate garbage collection (for testing)
func (mc *MemoryController) ForceGC() {
	log.Println("[MemCtrl] Force GC solicitado...")
	mc.triggerGC()
}

// PrintStats prints current memory statistics to log
func (mc *MemoryController) PrintStats() {
	stats := mc.GetStats()
	log.Printf("[MemCtrl] Stats: Level=%s, Usage=%.1f%%, Alloc=%d MB, Sys=%d MB, NumGC=%d",
		stats.Level,
		stats.UsagePercent,
		stats.AllocMB,
		stats.SysMB,
		stats.NumGC)
}

// ValidateConfig validates memory controller configuration
func ValidateMemoryControllerConfig(config MemoryControllerConfig) error {
	if !config.Enabled {
		return nil // Skip validation if disabled
	}

	if config.MaxMemoryMB == 0 {
		return fmt.Errorf("max_memory_mb deve ser > 0")
	}

	if config.WarningPercent <= 0 || config.WarningPercent >= 100 {
		return fmt.Errorf("warning_percent deve estar entre 0 e 100")
	}

	if config.CriticalPercent <= config.WarningPercent || config.CriticalPercent >= 100 {
		return fmt.Errorf("critical_percent deve ser > warning_percent e < 100")
	}

	if config.EmergencyPercent <= config.CriticalPercent || config.EmergencyPercent >= 100 {
		return fmt.Errorf("emergency_percent deve ser > critical_percent e < 100")
	}

	if config.CheckInterval <= 0 {
		return fmt.Errorf("check_interval deve ser > 0")
	}

	if config.GCTriggerPercent <= 0 || config.GCTriggerPercent >= 100 {
		return fmt.Errorf("gc_trigger_percent deve estar entre 0 e 100")
	}

	return nil
}
