# Edge Video V2 - Testing Checklist

## ✅ Funcionalidades Implementadas

### 1. Auto-Reconnect AMQP ⚠️ PENDENTE VALIDAÇÃO
**Status**: Implementado, não testado
**O que faz**: Reconecta automaticamente ao RabbitMQ se a conexão cair
**Como testar**:
- Rodar o sistema normalmente
- Derrubar o RabbitMQ (service restart)
- Observar logs de reconexão: 🛑 → 🔄 → ✓
- Verificar que frames continuam sendo publicados após reconexão

**Validação**: ⏳ PENDENTE (não conseguimos derrubar RabbitMQ para testar)

---

## ✅ Funcionalidades Testadas e Funcionando

### 1. Captura FFmpeg Stream Contínuo ✅ FUNCIONA
**Status**: Testado e aprovado
**O que faz**: Mantém stream FFmpeg aberto, captura frames continuamente
**Resultado**: Funciona perfeitamente, sem travamentos

### 2. Latest Frame Policy ✅ FUNCIONA
**Status**: Testado e aprovado
**O que faz**: Sempre publica o frame mais recente, descarta acumulados
**Resultado**: Sincronização perfeita entre câmeras, FPS estável

### 3. Dual-Goroutine Architecture ✅ FUNCIONA
**Status**: Testado e aprovado
**O que faz**: Separa leitura FFmpeg (readFrames) da publicação (publishLoop)
**Result**: Alta performance, sem bloqueios

### 4. HEVC/H.265 Codec Support ✅ FUNCIONA
**Status**: Testado e aprovado
**Câmeras**: Super Carlão (191.7.178.101:8554)
**Resultado**: Decodifica HEVC perfeitamente (V1.6 falhava!)

---

## ❌ Problemas Conhecidos (V1.6 - ABANDONADO)

### 1. HEVC Decoding Crash ❌
**Versão**: V1.6
**Erro**: "Could not find ref with POC X", memory leak 26GB
**Status**: NÃO SERÁ CORRIGIDO - V1.6 abandonado

### 2. Frame Desynchronization ❌
**Versão**: V1.6
**Erro**: Câmeras desincronizadas, worker pool gargalo
**Status**: RESOLVIDO NA V2 com Latest Frame Policy

---

## 🔄 Próximas Features (Ordem de Implementação)

### 2. Shutdown Statistics Report ⏳ EM PROGRESSO
**Prioridade**: ALTA
**O que faz**: Ao dar Ctrl+C, mostra relatório completo:
- FPS médio por câmera
- Total de frames capturados/publicados
- Throughput (KB/s, MB/s)
- Uptime do sistema
- Taxa de erro por câmera
- Latência média de publicação

### 3. Circuit Breaker para Câmeras 📋 PENDENTE
**Prioridade**: ALTA
**O que faz**: Se câmera falhar muito, entra em backoff exponencial
**Benefício**: Evita flood de erros, melhora performance

### 4. Frame Pooling (sync.Pool) 📋 PENDENTE
**Prioridade**: MÉDIA
**O que faz**: Reutiliza buffers de frames, reduz GC
**Benefício**: Menor uso de memória, menos pausas GC

### 5. Memory Controller 📋 PENDENTE
**Prioridade**: MÉDIA
**O que faz**: Monitora uso de RAM, faz throttle se necessário
**Benefício**: Evita crashes por falta de memória

### 6. Prometheus Metrics 📋 PENDENTE
**Prioridade**: BAIXA
**O que faz**: Exporta métricas para Prometheus/Grafana
**Benefício**: Observabilidade em produção

---

## 📊 Ambiente de Teste

**Câmeras**: 5x Super Carlão (191.7.178.101:8554, canais 1-5)
**Codec**: HEVC (H.265)
**RabbitMQ**: 34.71.212.239:5672
**FPS Target**: 15
**Quality**: 5 (JPEG)

**Máquinas Testadas**:
- ✅ Windows (desenvolvimento)
- ⏳ Ubuntu (Docker) - deployado mas não testado extensivamente
- ⏳ Linux direto - não testado

---

## 🎯 Critérios de Sucesso

Para cada feature ser considerada "pronta para produção":

1. ✅ Implementada e compilada sem erros
2. ✅ Testada em ambiente real (5 câmeras HEVC)
3. ✅ Rodando por pelo menos 1 hora sem crashes
4. ✅ Logs claros e informativos
5. ✅ Performance aceitável (CPU < 50%, RAM < 2GB)
6. ✅ Aprovada pelo usuário

---

**Última atualização**: 2025-12-05
**Versão**: V2.0
**Status Geral**: 🟢 ESTÁVEL (core features funcionando)
