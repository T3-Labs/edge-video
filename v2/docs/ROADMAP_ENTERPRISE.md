# 🚀 Edge Video V2 - Enterprise Roadmap

**Data**: Dezembro 2024
**Versão Atual**: V2.1 (Stable)
**Status**: Production Ready → Enterprise-Grade

---

## 🎯 Visão Geral

Este documento mapeia **TUDO** que é necessário para transformar Edge Video V2 de um sistema "production ready" para **ENTERPRISE-GRADE**, pronto para escala massiva (100+ câmeras, múltiplos clientes, 24/7/365).

**Metodologia**: Análise baseada em **impacto vs esforço** com foco em **valor real de negócio**.

---

## 📊 Análise de Gaps Críticos

### **Atual (V2.1)**
✅ Core funcional (captura, publicação, isolamento)
✅ Bug-free (frame contamination resolvido)
✅ Auto-reconnect AMQP
✅ Estatísticas básicas
✅ Suporta múltiplas câmeras (testado com 6)

### **Faltando para Enterprise**
❌ Observabilidade (métricas, alertas, dashboards)
❌ Recuperação de falhas (circuit breaker, retry inteligente)
❌ Configuração dinâmica (hot-reload, API de config)
❌ Segurança (autenticação, TLS, secrets management)
❌ Deployment automatizado (CI/CD, containers, K8s)
❌ Performance tuning (batching, compression, rate limiting)
❌ Multi-tenancy (isolamento por cliente)
❌ Disaster recovery (backup, failover, replicação)

---

## 🔥 Features Críticas (MUST HAVE)

### **Tier 1: Observabilidade e Confiabilidade** 🚨

#### **1. Health Check HTTP Endpoint**
**Prioridade**: 🔴 CRÍTICA
**Esforço**: 🟢 Baixo (2-3 horas)
**Impacto**: 🔴 ALTO

**O que é**:
Endpoint HTTP `/health` e `/ready` para monitorar status do sistema.

**Por que é crítico**:
- Load balancers precisam saber se instância está healthy
- Kubernetes/Docker precisam de probes para restart automático
- Alertas automáticos em caso de falha
- SLA compliance (99.9% uptime)

**Implementação**:
```go
// Endpoint /health
{
  "status": "healthy",
  "uptime": "2h34m12s",
  "cameras": {
    "total": 6,
    "active": 6,
    "failed": 0
  },
  "amqp": {
    "connected": true,
    "publish_rate": 84.5
  }
}

// Endpoint /ready
{
  "ready": true,
  "cameras_ready": 6,
  "amqp_ready": true
}
```

**Acceptance Criteria**:
- HTTP server na porta 8080
- `/health` retorna status geral
- `/ready` retorna se pode receber tráfego
- `/metrics` retorna métricas em formato Prometheus
- Responde em < 10ms

---

#### **2. Structured Logging (JSON)**
**Prioridade**: 🔴 CRÍTICA
**Esforço**: 🟡 Médio (4-6 horas)
**Impacto**: 🔴 ALTO

**O que é**:
Logs estruturados em JSON para parsing automático.

**Por que é crítico**:
- ELK/Splunk/DataDog precisam de JSON
- Alertas baseados em logs
- Debugging em produção (buscar por camera_id, error_type, etc.)
- Compliance (auditoria)

**Implementação**:
```go
// ANTES (texto simples):
log.Printf("[cam1] Frame #42 published")

// DEPOIS (JSON estruturado):
{
  "timestamp": "2024-12-05T10:30:45Z",
  "level": "info",
  "camera_id": "cam1",
  "event": "frame_published",
  "frame_number": 42,
  "latency_ms": 11,
  "size_bytes": 329563
}
```

**Biblioteca**: `github.com/rs/zerolog` (zero-allocation, super rápida)

**Acceptance Criteria**:
- Todos os logs em JSON
- Campos padronizados (timestamp, level, camera_id, event, msg)
- Log levels configuráveis (DEBUG, INFO, WARN, ERROR)
- Rotation automático (100MB por arquivo, máx 10 arquivos)

---

#### **3. Prometheus Metrics Exporter**
**Prioridade**: 🟠 ALTA
**Esforço**: 🟡 Médio (6-8 horas)
**Impacto**: 🔴 ALTO

**O que é**:
Exporta métricas para Prometheus/Grafana.

**Por que é importante**:
- Dashboards visuais em tempo real
- Alertas automáticos (PagerDuty, Slack)
- Trending histórico
- Capacity planning

**Métricas Essenciais**:
```
# Câmeras
camera_frames_total{camera_id="cam1"} 15420
camera_frames_dropped{camera_id="cam1"} 0
camera_fps{camera_id="cam1"} 14.2
camera_errors_total{camera_id="cam1"} 3
camera_status{camera_id="cam1",status="active"} 1

# AMQP
amqp_publish_total 15420
amqp_publish_errors 0
amqp_publish_latency_seconds{quantile="0.5"} 0.011
amqp_publish_latency_seconds{quantile="0.99"} 0.045
amqp_connected 1

# Sistema
system_goroutines 15
system_memory_bytes 125829120
system_uptime_seconds 9245
```

**Acceptance Criteria**:
- Endpoint `/metrics` em formato Prometheus
- Métricas por câmera
- Histogramas de latência
- Grafana dashboard pronto para usar

---

#### **4. Circuit Breaker para Câmeras**
**Prioridade**: 🟠 ALTA
**Esforço**: 🟡 Médio (4-6 horas)
**Impacto**: 🟠 MÉDIO

**O que é**:
Se câmera falha muito, entra em backoff exponencial.

**Por que é importante**:
- Evita flood de logs com erros repetitivos
- Poupa CPU/rede (não fica tentando conectar infinitamente)
- Isolamento de falhas (1 câmera ruim não afeta outras)

**Estados**:
```
CLOSED → câmera funcionando (tenta conectar)
OPEN → câmera falhando muito (para de tentar por X tempo)
HALF_OPEN → tenta reconectar (se sucesso → CLOSED, se falha → OPEN)
```

**Lógica**:
```
Falhas consecutivas:
0-5: Retry imediato
6-10: Backoff 10s
11-20: Backoff 30s
21+: Backoff 5min (circuit OPEN)

A cada 5 min, tenta 1x (HALF_OPEN)
Se sucesso: volta ao normal (CLOSED)
Se falha: volta ao backoff 5min (OPEN)
```

**Acceptance Criteria**:
- Máximo 5 retries antes de circuit open
- Log claro de mudança de estado
- Métrica `camera_circuit_breaker_state{camera_id, state}`
- Configurável via config.yaml

---

### **Tier 2: Operações e Deploy** 🛠️

#### **5. Configuração Hot-Reload**
**Prioridade**: 🟡 MÉDIA
**Esforço**: 🟠 Alto (8-12 horas)
**Impacto**: 🟠 MÉDIO

**O que é**:
Recarrega `config.yaml` sem restart do sistema.

**Por que é útil**:
- Adicionar/remover câmeras sem downtime
- Ajustar FPS/quality em runtime
- Mudança de credenciais AMQP sem restart

**Implementação**:
```go
// Watcher de arquivo (fsnotify)
// Ao detectar mudança em config.yaml:
1. Valida novo config
2. Calcula diff (câmeras adicionadas/removidas/modificadas)
3. Para câmeras removidas
4. Inicia câmeras adicionadas
5. Reconecta câmeras modificadas
```

**Acceptance Criteria**:
- Editar config.yaml não requer restart
- Mudanças aplicadas em < 5s
- Validação antes de aplicar (rollback se inválido)
- Log de cada mudança aplicada

---

#### **6. Graceful Shutdown Melhorado**
**Prioridade**: 🟡 MÉDIA
**Esforço**: 🟢 Baixo (2-3 horas)
**Impacto**: 🟡 BAIXO

**O que é**:
Shutdown mais robusto com timeout e flush de buffers.

**Melhorias**:
```go
1. Ao receber SIGTERM/SIGINT:
   - Para de aceitar novos frames
   - Publica frames pendentes (flush)
   - Espera até 30s para completar
   - Fecha conexões AMQP gracefully
   - Salva estatísticas finais em arquivo JSON
2. Se não completar em 30s: força shutdown
```

**Acceptance Criteria**:
- Zero frames perdidos no shutdown
- Estatísticas salvas em `stats_final.json`
- Timeout configurável
- Log de cada etapa do shutdown

---

#### **7. Docker Multi-Stage Build Otimizado**
**Prioridade**: 🟡 MÉDIA
**Esforço**: 🟢 Baixo (3-4 horas)
**Impacto**: 🟠 MÉDIO

**O que é**:
Dockerfile otimizado para produção.

**Melhorias**:
```dockerfile
# Stage 1: Builder
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache ffmpeg-dev
COPY . /src
WORKDIR /src
RUN go build -ldflags="-s -w" -o edge-video-v2

# Stage 2: Runtime
FROM alpine:latest
RUN apk add --no-cache ffmpeg ca-certificates
COPY --from=builder /src/edge-video-v2 /usr/local/bin/
COPY config.yaml /etc/edge-video/
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -q --spider http://localhost:8080/health || exit 1
USER nobody
ENTRYPOINT ["/usr/local/bin/edge-video-v2"]
```

**Resultado**:
- Imagem < 100MB (vs ~500MB atual)
- Scan de vulnerabilidades (Trivy)
- Non-root user (segurança)
- Health check integrado

**Acceptance Criteria**:
- Build em < 2 minutos
- Imagem < 100MB
- Zero vulnerabilidades críticas
- Docker Compose pronto

---

#### **8. Kubernetes Manifests**
**Prioridade**: 🟡 MÉDIA
**Esforço**: 🟡 Médio (6-8 horas)
**Impacto**: 🟠 MÉDIO

**O que é**:
Yamls para deploy em Kubernetes.

**Arquivos**:
```
k8s/
├── deployment.yaml      # Deployment principal
├── service.yaml         # Service (ClusterIP)
├── configmap.yaml       # ConfigMap para config.yaml
├── secret.yaml          # Secret para credenciais
├── hpa.yaml            # HorizontalPodAutoscaler
└── servicemonitor.yaml  # Prometheus ServiceMonitor
```

**Features**:
- Auto-scaling baseado em CPU (50-80% target)
- Rolling updates (zero downtime)
- Resource limits (CPU: 500m-2, Memory: 512Mi-2Gi)
- Liveness/Readiness probes
- ConfigMap para config (hot-reload via sidecar)

**Acceptance Criteria**:
- Deploy com `kubectl apply -f k8s/`
- Auto-scaling funciona
- Zero downtime em updates
- Secrets gerenciados via K8s

---

### **Tier 3: Performance e Escala** ⚡

#### **9. Frame Batching para AMQP**
**Prioridade**: 🟢 BAIXA
**Esforço**: 🟡 Médio (6-8 horas)
**Impacto**: 🟠 MÉDIO

**O que é**:
Agrupa múltiplos frames antes de publicar.

**Por que é útil**:
- Reduz overhead de rede (1 publish com 5 frames vs 5 publishes)
- Melhora throughput (80 FPS → 120+ FPS)
- Reduz latência no RabbitMQ (menos operações)

**Implementação**:
```go
// Acumula até 5 frames OU 100ms (o que vier primeiro)
batch := []Frame{}
timer := time.NewTimer(100 * time.Millisecond)

for {
    select {
    case frame := <-frameChan:
        batch = append(batch, frame)
        if len(batch) >= 5 {
            publishBatch(batch)
            batch = batch[:0]
        }
    case <-timer.C:
        if len(batch) > 0 {
            publishBatch(batch)
            batch = batch[:0]
        }
    }
}
```

**Trade-off**: +100ms latência média, mas +50% throughput

**Acceptance Criteria**:
- Configurável (batch_size, batch_timeout)
- Métricas de batch efficiency
- Latência < 150ms (P99)

---

#### **10. Adaptive Quality Control**
**Prioridade**: 🟢 BAIXA
**Esforço**: 🟠 Alto (10-12 horas)
**Impacto**: 🟠 MÉDIO

**O que é**:
Ajusta qualidade JPEG dinamicamente baseado em carga.

**Lógica**:
```
Se publish_latency > 50ms:
  → Reduz quality (5 → 7 → 10)
  → Frames menores = menos latência

Se publish_latency < 10ms E CPU < 50%:
  → Aumenta quality (10 → 7 → 5)
  → Melhor qualidade quando sistema está ocioso
```

**Benefício**: Mantém FPS estável mesmo sob carga

**Acceptance Criteria**:
- Quality ajusta automaticamente
- Métricas `camera_quality_current{camera_id}`
- Configurável (min_quality, max_quality, latency_threshold)

---

### **Tier 4: Segurança e Compliance** 🔒

#### **11. TLS/mTLS para AMQP**
**Prioridade**: 🟠 ALTA (produção)
**Esforço**: 🟡 Médio (4-6 horas)
**Impacto**: 🔴 ALTO (compliance)

**O que é**:
Conexão criptografada com RabbitMQ.

**Por que é crítico**:
- Compliance (LGPD, GDPR, PCI-DSS)
- Evita man-in-the-middle
- Autenticação mútua (mTLS)

**Implementação**:
```go
tlsConfig := &tls.Config{
    RootCAs: caCertPool,
    Certificates: []tls.Certificate{clientCert},
}
conn, err := amqp.DialTLS("amqps://...", tlsConfig)
```

**Acceptance Criteria**:
- Suporta TLS 1.2+
- Validação de certificados
- mTLS opcional
- Configurável via config.yaml

---

#### **12. Secrets Management**
**Prioridade**: 🟠 ALTA (produção)
**Esforço**: 🟡 Médio (6-8 horas)
**Impacto**: 🟠 MÉDIO

**O que é**:
Credenciais em Vault/AWS Secrets Manager, não em config.yaml.

**Por que é importante**:
- Segurança (não commita senhas no Git)
- Rotação automática de credenciais
- Auditoria de acesso

**Suporte**:
- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Variáveis de ambiente (fallback)

**Acceptance Criteria**:
- Credenciais nunca em plain text
- Suporta múltiplos backends
- Fallback para env vars
- Documentação completa

---

### **Tier 5: Multi-Tenancy e SaaS** 🏢

#### **13. Multi-Tenant Support**
**Prioridade**: 🟢 BAIXA (futuro)
**Esforço**: 🔴 Muito Alto (20-30 horas)
**Impacto**: 🔴 ALTO (novo modelo de negócio)

**O que é**:
Múltiplos clientes na mesma instância, completamente isolados.

**Arquitetura**:
```
edge-video-v2 --config=/etc/tenants/
  ├── tenant1/
  │   ├── config.yaml (cameras do cliente 1)
  │   └── stats/
  ├── tenant2/
  │   ├── config.yaml (cameras do cliente 2)
  │   └── stats/
  ...
```

**Isolamento**:
- Buffers separados por tenant
- Conexões AMQP separadas
- Métricas com label `tenant_id`
- Limites de recursos (CPU, RAM) por tenant

**Acceptance Criteria**:
- Zero cross-contamination entre tenants
- Hot-add/remove de tenants
- Billing metrics (frames_processed_total{tenant_id})

---

## 🎯 Roadmap Priorizado

### **Sprint 1: Observabilidade** (1-2 semanas)
1. ✅ Health Check HTTP Endpoint
2. ✅ Structured Logging (JSON)
3. ✅ Prometheus Metrics Exporter

**Entrega**: Sistema monitorável via Grafana

---

### **Sprint 2: Confiabilidade** (1 semana)
4. ✅ Circuit Breaker para Câmeras
5. ✅ Graceful Shutdown Melhorado

**Entrega**: Sistema resiliente a falhas

---

### **Sprint 3: Deploy e Ops** (1-2 semanas)
6. ✅ Docker Multi-Stage Build
7. ✅ Kubernetes Manifests
8. ✅ CI/CD Pipeline (GitHub Actions)

**Entrega**: Deploy automatizado em K8s

---

### **Sprint 4: Segurança** (1 semana)
9. ✅ TLS/mTLS para AMQP
10. ✅ Secrets Management

**Entrega**: Compliance-ready

---

### **Sprint 5: Performance** (1-2 semanas) - OPCIONAL
11. Frame Batching
12. Adaptive Quality Control
13. Config Hot-Reload

**Entrega**: Sistema otimizado para alta escala

---

### **Sprint 6: SaaS** (2-3 semanas) - FUTURO
14. Multi-Tenant Support
15. API REST para Management
16. Web Dashboard

**Entrega**: Produto SaaS completo

---

## 📊 Matriz de Priorização

| Feature | Impacto | Esforço | ROI | Prioridade |
|---------|---------|---------|-----|------------|
| Health Check | 🔴 Alto | 🟢 Baixo | 🔥 Muito Alto | 1️⃣ |
| Structured Logging | 🔴 Alto | 🟡 Médio | 🔥 Alto | 2️⃣ |
| Prometheus Metrics | 🔴 Alto | 🟡 Médio | 🔥 Alto | 3️⃣ |
| Circuit Breaker | 🟠 Médio | 🟡 Médio | 🟡 Médio | 4️⃣ |
| TLS/mTLS | 🔴 Alto | 🟡 Médio | 🔥 Alto | 5️⃣ |
| Secrets Management | 🟠 Médio | 🟡 Médio | 🟡 Médio | 6️⃣ |
| K8s Manifests | 🟠 Médio | 🟡 Médio | 🟡 Médio | 7️⃣ |
| Docker Optimized | 🟠 Médio | 🟢 Baixo | 🟡 Médio | 8️⃣ |
| Graceful Shutdown | 🟡 Baixo | 🟢 Baixo | 🟢 Baixo | 9️⃣ |
| Config Hot-Reload | 🟠 Médio | 🟠 Alto | 🟢 Baixo | 🔟 |
| Frame Batching | 🟠 Médio | 🟡 Médio | 🟢 Baixo | 1️⃣1️⃣ |
| Adaptive Quality | 🟡 Baixo | 🟠 Alto | 🟢 Baixo | 1️⃣2️⃣ |
| Multi-Tenant | 🔴 Alto | 🔴 Muito Alto | 🟡 Médio | 1️⃣3️⃣ |

---

## 🎓 Recomendação Final

### **FASE 1: MVP Enterprise** (3-4 semanas)
Implementar **Sprints 1-2 + TLS**:
- Health Check
- Structured Logging
- Prometheus Metrics
- Circuit Breaker
- TLS/mTLS

**Resultado**: Sistema pronto para produção enterprise com observabilidade completa.

### **FASE 2: DevOps Excellence** (2-3 semanas)
Implementar **Sprint 3**:
- Docker otimizado
- Kubernetes
- CI/CD

**Resultado**: Deploy automatizado, zero-downtime updates.

### **FASE 3: Otimização** (Opcional, 2-4 semanas)
Implementar features de performance conforme necessidade.

### **FASE 4: SaaS** (Futuro, 4-6 semanas)
Multi-tenancy + API + Dashboard quando escalar para múltiplos clientes.

---

## ✅ Checklist de Decisão

**Responda:**

1. Quantas câmeras em produção? (< 10, 10-50, 50-100, 100+)
2. SLA requerido? (95%, 99%, 99.9%, 99.99%)
3. Orçamento de infra? (baixo, médio, alto)
4. Time de DevOps? (sim/não)
5. Compliance necessário? (LGPD, GDPR, etc.)
6. Múltiplos clientes/tenants? (sim/não - futuro?)
7. Já tem Kubernetes? (sim/não)
8. Já tem Prometheus/Grafana? (sim/não)

**Com base nas respostas, eu monto um roadmap CUSTOMIZADO!**

---

**O QUE VOCÊ ACHA? QUAIS DESSAS FEATURES FAZEM MAIS SENTIDO PARA SEU CASO DE USO?** 🎯
