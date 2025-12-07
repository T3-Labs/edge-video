package registration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/T3-Labs/edge-video/pkg/config"
	"github.com/T3-Labs/edge-video/pkg/logger"
)

// RegistrationPayload representa o payload enviado para a API de registro
type RegistrationPayload struct {
	Cameras     []CameraInfo `json:"cameras"`
	Namespace   string       `json:"namespace"`
	RabbitMQURL string       `json:"rabbitmq_url"`
	RoutingKey  string       `json:"routing_key"`
	Exchange    string       `json:"exchange"`
	Vhost       string       `json:"vhost"`
}

// CameraInfo representa informações de uma câmera
type CameraInfo struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Client gerencia o registro do serviço na API
type Client struct {
	apiURL     string
	httpClient *http.Client
	enabled    bool
}

// NewClient cria um novo cliente de registro
func NewClient(apiURL string, enabled bool) *Client {
	return &Client{
		apiURL:  apiURL,
		enabled: enabled,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Register envia os dados de registro para a API
func (c *Client) Register(ctx context.Context, cfg *config.Config, vhost string) error {
	if !c.enabled {
		return nil
	}

	if c.apiURL == "" {
		return fmt.Errorf("registration API URL is empty")
	}

	// Converte as câmeras para o formato do payload
	cameras := make([]CameraInfo, len(cfg.Cameras))
	for i, cam := range cfg.Cameras {
		cameras[i] = CameraInfo{
			ID:  cam.ID,
			URL: cam.URL,
		}
	}

	payload := RegistrationPayload{
		Cameras:     cameras,
		Namespace:   vhost,
		RabbitMQURL: cfg.RabbitMQ.URL,
		RoutingKey:  cfg.RabbitMQ.RoutingKey,
		Exchange:    cfg.RabbitMQ.Exchange,
		Vhost:       vhost,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal registration payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send registration request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("registration failed with status code: %d", resp.StatusCode)
	}

	if logger.Log != nil {
		logger.Log.Infow("Successfully registered with API",
			"api_url", c.apiURL,
			"vhost", vhost,
			"cameras_count", len(cameras))
	}

	return nil
}

// RegisterWithRetry tenta registrar o serviço na API com retry a cada 1 minuto em caso de falha
// Executa em background (goroutine) e continua tentando até ter sucesso
func (c *Client) RegisterWithRetry(ctx context.Context, cfg *config.Config, vhost string) {
	if !c.enabled {
		return
	}

	// Primeira tentativa imediata
	err := c.Register(ctx, cfg, vhost)
	if err == nil {
		return // Sucesso na primeira tentativa
	}

	if logger.Log != nil {
		logger.Log.Warnw("Failed to register with API, will retry every 1 minute",
			"error", err,
			"api_url", c.apiURL)
	}

	// Inicia goroutine para retry em background
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				if logger.Log != nil {
					logger.Log.Info("Registration retry stopped due to context cancellation")
				}
				return
			case <-ticker.C:
				err := c.Register(ctx, cfg, vhost)
				if err == nil {
					if logger.Log != nil {
						logger.Log.Info("Successfully registered with API after retry")
					}
					return // Para o retry após sucesso
				}
				if logger.Log != nil {
					logger.Log.Warnw("Registration retry failed, will try again in 1 minute",
						"error", err)
				}
			}
		}
	}()
}
