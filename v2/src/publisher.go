package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher gerencia publicação no RabbitMQ com auto-reconnect
type Publisher struct {
	amqpURL       string
	conn          *amqp.Connection
	channel       *amqp.Channel
	exchange      string
	routingKey    string // Routing key COMPLETA (não é mais prefixo)
	prefetchCount int    // QoS: limite de frames não-confirmados (0 = ilimitado)

	mu            sync.Mutex
	publishMu     sync.Mutex // Mutex DEDICADO para serializar publicações (channel.Publish não é thread-safe!)
	publishCount  uint64
	publishErrors uint64
	reconnecting  bool
	connected     bool

	// Publisher Confirms (rastreamento de entregas)
	confirmsChan  chan amqp.Confirmation
	confirmsCount uint64 // Total de confirms recebidos (ACK)
	nacksCount    uint64 // Total de NACKs recebidos (rejeições)

	notifyClose chan *amqp.Error
	done        chan struct{}
}

// NewPublisher cria um novo publisher com auto-reconnect
func NewPublisher(amqpURL, exchange, routingKey string, prefetchCount int) (*Publisher, error) {
	p := &Publisher{
		amqpURL:       amqpURL,
		exchange:      exchange,
		routingKey:    routingKey,    // Usa routing_key completa
		prefetchCount: prefetchCount, // QoS configurável
		done:          make(chan struct{}),
	}

	// Conecta inicialmente com retry
	err := p.connectWithRetry(10, 5*time.Second)
	if err != nil {
		return nil, err
	}

	// Monitora conexão em background
	go p.monitorConnection()

	log.Printf("✓ Conectado ao RabbitMQ - Exchange: %s", exchange)

	return p, nil
}

// connectWithRetry tenta conectar com retry exponencial
func (p *Publisher) connectWithRetry(maxRetries int, initialDelay time.Duration) error {
	delay := initialDelay

	for i := 0; i < maxRetries; i++ {
		err := p.connect()
		if err == nil {
			p.mu.Lock()
			p.connected = true
			p.mu.Unlock()
			return nil
		}

		log.Printf("⚠ Tentativa %d/%d falhou: %v. Retry em %v...", i+1, maxRetries, err, delay)
		time.Sleep(delay)

		// Backoff exponencial: 5s, 10s, 20s (max 30s)
		delay *= 2
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
	}

	return fmt.Errorf("falha após %d tentativas", maxRetries)
}

// connect estabelece conexão com RabbitMQ
func (p *Publisher) connect() error {
	var err error

	// Conecta
	p.conn, err = amqp.Dial(p.amqpURL)
	if err != nil {
		return fmt.Errorf("falha ao conectar: %w", err)
	}

	// Cria canal
	p.channel, err = p.conn.Channel()
	if err != nil {
		p.conn.Close()
		return fmt.Errorf("falha ao criar canal: %w", err)
	}

	// Declara exchange
	err = p.channel.ExchangeDeclare(
		p.exchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		p.channel.Close()
		p.conn.Close()
		return fmt.Errorf("falha ao declarar exchange: %w", err)
	}

	// CONFIGURA QoS (Quality of Service)
	// Limita quantos frames não-confirmados podem estar em trânsito
	// Isso previne:
	// - Consumer overflow (consumer recebe milhares de frames de uma vez)
	// - Memory overflow no consumer
	// - Processamento em lote que causa latência
	err = p.channel.Qos(
		p.prefetchCount, // prefetchCount: configurável via config.yaml (0 = ilimitado)
		0,               // prefetchSize: sem limite de bytes (0 = ilimitado)
		false,           // global: false = aplica apenas a este channel
	)
	if err != nil {
		p.channel.Close()
		p.conn.Close()
		return fmt.Errorf("falha ao configurar QoS: %w", err)
	}

	// HABILITA PUBLISHER CONFIRMS
	// Isso faz o RabbitMQ enviar confirmações (ACK/NACK) para cada mensagem publicada
	err = p.channel.Confirm(false)
	if err != nil {
		p.channel.Close()
		p.conn.Close()
		return fmt.Errorf("falha ao habilitar publisher confirms: %w", err)
	}

	// Canal para receber confirmações
	p.confirmsChan = p.channel.NotifyPublish(make(chan amqp.Confirmation, 1000))

	// Inicia goroutine para processar confirmações
	go p.handleConfirms()

	// Monitora fechamento de conexão
	p.notifyClose = make(chan *amqp.Error)
	p.conn.NotifyClose(p.notifyClose)

	log.Printf("✓ QoS configurado: prefetch=%d | Publisher Confirms habilitado para exchange: %s", p.prefetchCount, p.exchange)

	return nil
}

// handleConfirms processa confirmações (ACK/NACK) do RabbitMQ
func (p *Publisher) handleConfirms() {
	for {
		select {
		case <-p.done:
			return

		case confirm, ok := <-p.confirmsChan:
			if !ok {
				// Canal fechado (reconexão em andamento)
				return
			}

			p.mu.Lock()
			if confirm.Ack {
				// ACK: Frame entregue com sucesso ao RabbitMQ
				p.confirmsCount++
			} else {
				// NACK: Frame rejeitado pelo RabbitMQ
				p.nacksCount++
				log.Printf("⚠️  NACK recebido! Frame rejeitado pelo RabbitMQ (delivery tag: %d)", confirm.DeliveryTag)
			}
			p.mu.Unlock()

			// Tracking para profiling
			TrackPublishConfirm(confirm.Ack)
		}
	}
}

// monitorConnection monitora e reconecta automaticamente
func (p *Publisher) monitorConnection() {
	for {
		select {
		case <-p.done:
			return

		case err := <-p.notifyClose:
			if err != nil {
				log.Printf("🛑 Conexão RabbitMQ perdida: %v", err)
				p.mu.Lock()
				p.connected = false
				p.mu.Unlock()

				p.reconnect()
			}
		}
	}
}

// reconnect tenta reconectar indefinidamente
func (p *Publisher) reconnect() {
	p.mu.Lock()
	if p.reconnecting {
		p.mu.Unlock()
		return
	}
	p.reconnecting = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.reconnecting = false
		p.mu.Unlock()
	}()

	delay := 1 * time.Second

	for {
		select {
		case <-p.done:
			return
		default:
		}

		log.Printf("🔄 Tentando reconectar ao RabbitMQ...")

		// Fecha conexão antiga se existir
		if p.channel != nil {
			p.channel.Close()
		}
		if p.conn != nil {
			p.conn.Close()
		}

		// Tenta reconectar
		err := p.connect()
		if err == nil {
			p.mu.Lock()
			p.connected = true
			p.mu.Unlock()
			log.Printf("✓ Reconectado ao RabbitMQ com sucesso!")
			return
		}

		log.Printf("⚠ Reconexão falhou: %v. Retry em %v...", err, delay)
		time.Sleep(delay)

		// Backoff exponencial: 1s, 2s, 5s, 10s (max 10s)
		if delay < 2*time.Second {
			delay = 2 * time.Second
		} else if delay < 5*time.Second {
			delay = 5 * time.Second
		} else {
			delay = 10 * time.Second
		}
	}
}

// Publish publica um frame no RabbitMQ com retry
func (p *Publisher) Publish(cameraID string, frameData []byte, timestamp time.Time) error {
	// CRÍTICO: Todo o processo de publicação deve ser ATÔMICO
	// Adquire AMBOS os locks no início para evitar race conditions
	p.publishMu.Lock()
	defer p.publishMu.Unlock()

	p.mu.Lock()

	// Se não conectado, retorna erro
	if !p.connected {
		p.publishErrors++
		p.mu.Unlock()
		return fmt.Errorf("não conectado ao RabbitMQ")
	}

	// USA A ROUTING KEY FIXA DO PUBLISHER (já configurada por câmera)
	routingKey := p.routingKey

	// CRÍTICO: Captura o channel DENTRO do lock para evitar race condition
	channel := p.channel

	// DEBUG: Log detalhado de publicação (primeiros 18 frames)
	if p.publishCount < 18 { // 3 frames x 6 cameras = 18 frames
		log.Printf("[PUBLISH DEBUG] Camera: %s, RoutingKey: %s, Size: %d bytes, Header[camera_id]: %s",
			cameraID, routingKey, len(frameData), cameraID)
	}

	p.mu.Unlock()

	// CRÍTICO: FAZ CÓPIA DEFENSIVA ANTES DE PASSAR PARA AMQP
	// A biblioteca streadway/amqp pode manter referência ao slice internamente!
	// Esta é a ÚLTIMA linha de defesa contra race conditions
	frameDataCopy := make([]byte, len(frameData))
	copy(frameDataCopy, frameData)

	// Tenta publicar com a CÓPIA DEFENSIVA
	// IMPORTANTE: Serializado pelo publishMu (defer unlock no topo)
	err := channel.Publish(
		p.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         frameDataCopy, // USA CÓPIA DEFENSIVA!
			Timestamp:    timestamp,
			DeliveryMode: amqp.Transient, // Não persiste (mais rápido)
			Headers: amqp.Table{
				"camera_id": cameraID,
			},
		},
	)

	// Re-adquire lock para atualizar contadores
	p.mu.Lock()

	if err != nil {
		p.publishErrors++
		p.connected = false // Marca como desconectado
		p.mu.Unlock()

		// Trigger reconexão
		go p.reconnect()

		return fmt.Errorf("falha ao publicar: %w", err)
	}

	p.publishCount++
	p.mu.Unlock()
	return nil
}

// Close fecha a conexão
func (p *Publisher) Close() error {
	close(p.done)

	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// Stats retorna estatísticas
func (p *Publisher) Stats() (uint64, uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.publishCount, p.publishErrors
}

// IsConnected retorna se está conectado
func (p *Publisher) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.connected
}

// ConfirmStats retorna estatísticas de confirmações
func (p *Publisher) ConfirmStats() (acks uint64, nacks uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.confirmsCount, p.nacksCount
}
