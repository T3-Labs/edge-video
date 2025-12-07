package config

import (
	"testing"
)

func TestExtractVhostFromAMQP(t *testing.T) {
	tests := []struct {
		name    string
		amqpURL string
		want    string
	}{
		{
			name:    "URL com vhost customizado",
			amqpURL: "amqp://guest:guest@localhost:5672/myvhost",
			want:    "myvhost",
		},
		{
			name:    "URL sem vhost (usa default /)",
			amqpURL: "amqp://guest:guest@localhost:5672",
			want:    "/",
		},
		{
			name:    "URL com vhost vazio (usa /)",
			amqpURL: "amqp://guest:guest@localhost:5672/",
			want:    "/",
		},
		{
			name:    "URL com vhost contendo caracteres especiais",
			amqpURL: "amqp://user:pass@host:5672/client-123",
			want:    "client-123",
		},
		{
			name:    "URL vazia retorna /",
			amqpURL: "",
			want:    "/",
		},
		{
			name:    "URL inválida retorna /",
			amqpURL: "://invalid",
			want:    "/",
		},
		{
			name:    "URL com múltiplos segmentos no path",
			amqpURL: "amqp://localhost:5672/vhost/extra",
			want:    "vhost/extra",
		},
		{
			name:    "URL production-like",
			amqpURL: "amqp://production-user:secret@rabbitmq.prod.com:5672/production-vhost",
			want:    "production-vhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				RabbitMQ: RabbitMQConfig{
					URL: tt.amqpURL,
				},
			}

			got := cfg.ExtractVhostFromAMQP()
			if got != tt.want {
				t.Errorf("ExtractVhostFromAMQP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVhostFromAMQP_RealWorldCases(t *testing.T) {
	tests := []struct {
		name        string
		amqpURL     string
		expectedKey string
	}{
		{
			name:        "Cliente A com vhost dedicado",
			amqpURL:     "amqp://user:pass@rabbitmq:5672/client-a",
			expectedKey: "client-a",
		},
		{
			name:        "Cliente B com vhost dedicado",
			amqpURL:     "amqp://user:pass@rabbitmq:5672/client-b",
			expectedKey: "client-b",
		},
		{
			name:        "Ambiente de desenvolvimento",
			amqpURL:     "amqp://guest:guest@localhost:5672/dev",
			expectedKey: "dev",
		},
		{
			name:        "Ambiente de staging",
			amqpURL:     "amqp://user:pass@staging-mq:5672/staging",
			expectedKey: "staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				RabbitMQ: RabbitMQConfig{
					URL: tt.amqpURL,
				},
			}

			vhost := cfg.ExtractVhostFromAMQP()
			if vhost != tt.expectedKey {
				t.Errorf("Vhost extraído = %v, esperado %v", vhost, tt.expectedKey)
			}

			// Verifica que diferentes URLs produzem vhosts únicos
			// (importante para isolamento multi-tenant)
			t.Logf("URL: %s -> vhost: %s", tt.amqpURL, vhost)
		})
	}
}

func TestVhostUniqueness(t *testing.T) {
	// Testa que diferentes vhosts produzem identificadores únicos
	urls := []string{
		"amqp://localhost:5672/client-a",
		"amqp://localhost:5672/client-b",
		"amqp://localhost:5672/client-c",
	}

	vhosts := make(map[string]bool)

	for _, url := range urls {
		cfg := &Config{
			RabbitMQ: RabbitMQConfig{URL: url},
		}
		vhost := cfg.ExtractVhostFromAMQP()

		if vhosts[vhost] {
			t.Errorf("Vhost duplicado detectado: %s para URL %s", vhost, url)
		}
		vhosts[vhost] = true
	}

	if len(vhosts) != len(urls) {
		t.Errorf("Esperado %d vhosts únicos, obteve %d", len(urls), len(vhosts))
	}
}
