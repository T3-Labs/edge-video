package registration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/T3-Labs/edge-video/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080/api/register", true)

	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080/api/register", client.apiURL)
	assert.True(t, client.enabled)
	assert.NotNil(t, client.httpClient)
}

func TestRegister_Disabled(t *testing.T) {
	client := NewClient("http://localhost:8080/api/register", false)

	cfg := &config.Config{}
	err := client.Register(context.Background(), cfg, "test_vhost")

	assert.NoError(t, err)
}

func TestRegister_EmptyURL(t *testing.T) {
	client := NewClient("", true)

	cfg := &config.Config{}
	err := client.Register(context.Background(), cfg, "test_vhost")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API URL is empty")
}

func TestRegister_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload RegistrationPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)

		assert.Equal(t, "test_vhost", payload.Vhost)
		assert.Equal(t, "test_vhost", payload.Namespace)
		assert.Equal(t, "test_exchange", payload.Exchange)
		assert.Len(t, payload.Cameras, 2)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, true)

	cfg := &config.Config{
		RabbitMQ: config.RabbitMQConfig{
			URL:        "amqp://user:pass@host:5672/test_vhost",
			Exchange:   "test_exchange",
			RoutingKey: "camera.",
		},
		Cameras: []config.CameraConfig{
			{ID: "cam1", URL: "rtsp://test1"},
			{ID: "cam2", URL: "rtsp://test2"},
		},
	}

	err := client.Register(context.Background(), cfg, "test_vhost")
	assert.NoError(t, err)
}

func TestRegister_ServerError(t *testing.T) {
	// Mock server que retorna erro 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, true)

	cfg := &config.Config{
		RabbitMQ: config.RabbitMQConfig{
			URL:        "amqp://user:pass@host:5672/test_vhost",
			Exchange:   "test_exchange",
			RoutingKey: "camera.",
		},
		Cameras: []config.CameraConfig{
			{ID: "cam1", URL: "rtsp://test1"},
		},
	}

	err := client.Register(context.Background(), cfg, "test_vhost")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code: 500")
}

func TestRegister_ContextCancellation(t *testing.T) {
	// Mock server com delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, true)

	cfg := &config.Config{
		RabbitMQ: config.RabbitMQConfig{
			URL:      "amqp://user:pass@host:5672/test_vhost",
			Exchange: "test_exchange",
		},
		Cameras: []config.CameraConfig{
			{ID: "cam1", URL: "rtsp://test1"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.Register(ctx, cfg, "test_vhost")
	assert.Error(t, err)
}

func TestRegisterWithRetry_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, true)

	cfg := &config.Config{
		RabbitMQ: config.RabbitMQConfig{
			URL:      "amqp://user:pass@host:5672/test_vhost",
			Exchange: "test_exchange",
		},
		Cameras: []config.CameraConfig{
			{ID: "cam1", URL: "rtsp://test1"},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// RegisterWithRetry deve retornar imediatamente se sucesso na primeira tentativa
	client.RegisterWithRetry(ctx, cfg, "test_vhost")

	// Pequeno delay para garantir que a goroutine foi iniciada se necessário
	time.Sleep(100 * time.Millisecond)
}

func TestRegisterWithRetry_Disabled(t *testing.T) {
	client := NewClient("http://localhost:8080/api/register", false)

	cfg := &config.Config{}
	ctx := context.Background()

	// Não deve fazer nada se desabilitado
	client.RegisterWithRetry(ctx, cfg, "test_vhost")
}

func TestRegistrationPayload_MarshalJSON(t *testing.T) {
	payload := RegistrationPayload{
		Cameras: []CameraInfo{
			{ID: "cam1", URL: "rtsp://test1"},
			{ID: "cam2", URL: "rtsp://test2"},
		},
		Namespace:   "test_namespace",
		RabbitMQURL: "amqp://user:pass@host:5672/vhost",
		RoutingKey:  "camera.",
		Exchange:    "test_exchange",
		Vhost:       "test_vhost",
	}

	jsonData, err := json.Marshal(payload)
	assert.NoError(t, err)

	var decoded RegistrationPayload
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, payload.Namespace, decoded.Namespace)
	assert.Equal(t, payload.Vhost, decoded.Vhost)
	assert.Len(t, decoded.Cameras, 2)
}
