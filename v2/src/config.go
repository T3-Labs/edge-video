package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config representa a configuração do sistema
type Config struct {
	FPS              int                      `yaml:"fps"`
	Quality          int                      `yaml:"quality"`
	AMQP             AMQPConfig               `yaml:"amqp"`
	CircuitBreaker   CircuitBreakerConfig     `yaml:"circuit_breaker"`
	MemoryController MemoryControllerConfig   `yaml:"memory_controller"`
	Cameras          []CamConfig              `yaml:"cameras"`
}

// AMQPConfig configuração do RabbitMQ
type AMQPConfig struct {
	URL              string `yaml:"url"`
	Exchange         string `yaml:"exchange"`
	RoutingKeyPrefix string `yaml:"routing_key_prefix"`
	PrefetchCount    int    `yaml:"prefetch_count"` // QoS: limite de frames não-confirmados (0 = ilimitado)
}

// CamConfig configuração de câmera
type CamConfig struct {
	ID         string `yaml:"id"`
	URL        string `yaml:"url"`
	Exchange   string `yaml:"exchange"`   // Exchange dedicado para esta câmera
	RoutingKey string `yaml:"routing_key"` // Routing key dedicada para esta câmera
}

// LoadConfig carrega configuração do arquivo YAML
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear YAML: %w", err)
	}

	// Validações
	if config.FPS <= 0 || config.FPS > 60 {
		return nil, fmt.Errorf("FPS inválido: %d (deve ser 1-60)", config.FPS)
	}

	if config.Quality < 2 || config.Quality > 31 {
		return nil, fmt.Errorf("Quality inválido: %d (deve ser 2-31)", config.Quality)
	}

	if len(config.Cameras) == 0 {
		return nil, fmt.Errorf("nenhuma câmera configurada")
	}

	// Se prefetch_count não configurado, usa default de 50
	if config.AMQP.PrefetchCount == 0 {
		config.AMQP.PrefetchCount = 50
	}

	// Se circuit_breaker não configurado, usa defaults
	if config.CircuitBreaker.MaxFailures == 0 {
		config.CircuitBreaker = DefaultCircuitBreakerConfig()
	}

	// Se memory_controller não configurado, usa defaults (disabled)
	if config.MemoryController.MaxMemoryMB == 0 {
		config.MemoryController = DefaultMemoryControllerConfig()
	}

	// Valida memory_controller se habilitado
	if config.MemoryController.Enabled {
		if err := ValidateMemoryControllerConfig(config.MemoryController); err != nil {
			return nil, fmt.Errorf("erro na configuração de memory_controller: %w", err)
		}
	}

	return &config, nil
}

// DefaultMemoryControllerConfig retorna configuração padrão para memory controller
func DefaultMemoryControllerConfig() MemoryControllerConfig {
	return MemoryControllerConfig{
		Enabled:          false, // Disabled por padrão
		MaxMemoryMB:      2048,
		WarningPercent:   60.0,
		CriticalPercent:  75.0,
		EmergencyPercent: 85.0,
		CheckInterval:    5000000000, // 5s em nanoseconds
		GCTriggerPercent: 70.0,
	}
}
