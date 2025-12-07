package config

import (
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type CameraConfig struct {
	ID   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
	URL  string `mapstructure:"url"`
}

// Backwards-compatible AMQP config used by tests/examples
type AMQPConfig struct {
	AmqpURL          string `mapstructure:"amqp_url"`
	Exchange         string `mapstructure:"exchange"`
	RoutingKeyPrefix string `mapstructure:"routing_key_prefix"`
	TTLSeconds       int    `mapstructure:"ttl_seconds"`
}

// MQTT config (kept for compatibility)
type MQTTConfig struct {
	Broker      string `mapstructure:"broker"`
	TopicPrefix string `mapstructure:"topic_prefix"`
	TTLSeconds  int    `mapstructure:"ttl_seconds"`
}

// Metadata config (some examples use cfg.Metadata)
type MetadataConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Exchange   string `mapstructure:"exchange"`
	RoutingKey string `mapstructure:"routing_key"`
	TTLSeconds int    `mapstructure:"ttl_seconds"`
}

type RabbitMQConfig struct {
	URL        string `mapstructure:"url"`
	Exchange   string `mapstructure:"exchange"`
	RoutingKey string `mapstructure:"routing_key"`
	TTLSeconds int    `mapstructure:"ttl_seconds"` // optional per-service override
}

type Compression struct {
	Enabled bool `mapstructure:"enabled"`
	Level   int  `mapstructure:"level"`
}

type Optimization struct {
	MaxWorkers         int    `mapstructure:"max_workers"`
	BufferSize         int    `mapstructure:"buffer_size"`
	WorkerQueueSize    int    `mapstructure:"worker_queue_size"`
	CameraBufferSize   int    `mapstructure:"camera_buffer_size"`
	PersistentBufSize  int    `mapstructure:"persistent_buffer_size"`
	FrameQuality       int    `mapstructure:"frame_quality"`
	FrameResolution    string `mapstructure:"frame_resolution"`
	UsePersistent      bool   `mapstructure:"use_persistent"`
	CircuitMaxFailures int    `mapstructure:"circuit_max_failures"`
	CircuitResetSec    int    `mapstructure:"circuit_reset_seconds"`
}

type RedisConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Address    string `mapstructure:"address"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	TTLSeconds int    `mapstructure:"ttl_seconds"`
	Prefix     string `mapstructure:"prefix"`
}

type RegistrationConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIURL  string `mapstructure:"api_url"`
}

type MemoryConfig struct {
	Enabled          bool    `mapstructure:"enabled"`
	MaxMemoryMB      uint64  `mapstructure:"max_memory_mb"`
	WarningPercent   float64 `mapstructure:"warning_percent"`
	CriticalPercent  float64 `mapstructure:"critical_percent"`
	EmergencyPercent float64 `mapstructure:"emergency_percent"`
	CheckInterval    int     `mapstructure:"check_interval_seconds"`
	GCTriggerPercent float64 `mapstructure:"gc_trigger_percent"`
}

// Config is the application configuration. It includes both the newer RabbitMQ fields
// and legacy-compatible AMQP/MQTT/Metadata fields used by tests and examples.
type Config struct {
	TargetFPS           float64            `mapstructure:"target_fps"`
	UseOptimizedCapture bool               `mapstructure:"use_optimized_capture"`
	Protocol            string             `mapstructure:"protocol"`
	AMQP                AMQPConfig         `mapstructure:"amqp"`
	MQTT                MQTTConfig         `mapstructure:"mqtt"`
	Metadata            MetadataConfig     `mapstructure:"metadata"`
	RabbitMQ            RabbitMQConfig     `mapstructure:"rabbitmq"`
	Redis               RedisConfig        `mapstructure:"redis"`
	Registration        RegistrationConfig `mapstructure:"registration"`
	Compression         Compression        `mapstructure:"compression"`
	Optimization        Optimization       `mapstructure:"optimization"`
	Memory              MemoryConfig       `mapstructure:"memory"`
	Cameras             []CameraConfig     `mapstructure:"cameras"`
	TTLSeconds          int                `mapstructure:"ttl_seconds"` // global default (seconds)
}

// GetFrameInterval calcula o intervalo de tempo entre os frames com base no TargetFPS.
// Retorna um intervalo padrão de 2 FPS se o valor for inválido.
func (c *Config) GetFrameInterval() time.Duration {
	if c.TargetFPS <= 0 {
		return time.Second / 2 // Padrão: 2 FPS
	}
	return time.Duration(float64(time.Second) / c.TargetFPS)
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Default TTL handling: if not provided, use 30s. Allow global ttl_seconds to set both
	if cfg.TTLSeconds <= 0 {
		cfg.TTLSeconds = 30
	}

	if cfg.Redis.TTLSeconds <= 0 {
		cfg.Redis.TTLSeconds = cfg.TTLSeconds
	}

	// Map legacy AMQP fields to internal RabbitMQ fields for compatibility
	if cfg.RabbitMQ.URL == "" && cfg.AMQP.AmqpURL != "" {
		cfg.RabbitMQ.URL = cfg.AMQP.AmqpURL
	}
	if cfg.RabbitMQ.Exchange == "" && cfg.AMQP.Exchange != "" {
		cfg.RabbitMQ.Exchange = cfg.AMQP.Exchange
	}
	if cfg.RabbitMQ.RoutingKey == "" && cfg.AMQP.RoutingKeyPrefix != "" {
		// The older field is a prefix; map it to routing_key as-is for compatibility
		cfg.RabbitMQ.RoutingKey = cfg.AMQP.RoutingKeyPrefix
	}

	// Map Metadata config if present
	if cfg.Metadata.Exchange != "" && cfg.RabbitMQ.Exchange == "" {
		cfg.RabbitMQ.Exchange = cfg.Metadata.Exchange
	}
	if cfg.Metadata.RoutingKey != "" && cfg.RabbitMQ.RoutingKey == "" {
		cfg.RabbitMQ.RoutingKey = cfg.Metadata.RoutingKey
	}

	if cfg.RabbitMQ.TTLSeconds <= 0 {
		// prefer explicit metadata or AMQP TTL if set
		if cfg.Metadata.TTLSeconds > 0 {
			cfg.RabbitMQ.TTLSeconds = cfg.Metadata.TTLSeconds
		} else if cfg.AMQP.TTLSeconds > 0 {
			cfg.RabbitMQ.TTLSeconds = cfg.AMQP.TTLSeconds
		} else {
			cfg.RabbitMQ.TTLSeconds = cfg.TTLSeconds
		}
	}

	// Ensure Protocol has a default
	if cfg.Protocol == "" {
		cfg.Protocol = "amqp"
	}

	return &cfg, nil
}

// ExtractVhostFromAMQP extrai o vhost da URL AMQP.
// Exemplo: amqp://user:pass@host:5672/myvhost -> "myvhost"
// Retorna "/" se nenhum vhost for especificado
func (c *Config) ExtractVhostFromAMQP() string {
	// Use the AMQP amqp_url if present, otherwise RabbitMQ URL
	amqpURL := c.AMQP.AmqpURL
	if amqpURL == "" {
		amqpURL = c.RabbitMQ.URL
	}

	if amqpURL == "" {
		return "/"
	}

	parsedURL, err := url.Parse(amqpURL)
	if err != nil {
		return "/"
	}

	vhost := strings.TrimPrefix(parsedURL.Path, "/")
	if vhost == "" {
		return "/"
	}

	return vhost
}
