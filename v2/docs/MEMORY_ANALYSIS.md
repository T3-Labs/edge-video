# 📊 Análise de Consumo de Memória - Edge Video V2

**Data**: 2025-12-05
**Versão**: V2.2 com Memory Controller
**Configuração**: 6 câmeras (5 funcionando, 1 com Circuit Breaker OPEN)

---

## 🎯 Resumo Executivo

| Métrica | Valor |
|---------|-------|
| **Câmeras ativas** | 5 (cam1, cam2, cam3, cam4, cam5) |
| **FPS target** | 15 |
| **JPEG Quality** | 5 |
| **Consumo estimado** | ~558 MB |
| **Max memory configurado** | 2048 MB (ATUAL) → **1024 MB (RECOMENDADO)** |
| **Throughput total** | ~11.6 MB/s |

---

## 📸 Análise por Câmera

### Tamanhos de Frame (baseado em logs reais)

| Câmera | Protocolo | Tamanho Médio/Frame | Comentários |
|--------|-----------|---------------------|-------------|
| **cam1** | RTMP | ~320 KB | Maior frame (maior resolução?) |
| **cam2** | RTSP | ~64 KB | Menor frame (compressão eficiente) |
| **cam3** | RTSP | ~180 KB | Frame médio |
| **cam4** | RTSP | ~115 KB | Frame médio |
| **cam5** | RTSP | ~97 KB | Frame médio-pequeno |
| **cam6** | RTSP | N/A | **Circuit Breaker OPEN** (canal inválido: `channel=banana`) |

**Média ponderada**: ~155 KB/frame (considerando apenas as 5 câmeras funcionando)

---

## 💾 Breakdown de Consumo de Memória

### 1. FFmpeg Processes (MAIOR CONSUMO)
Cada processo FFmpeg consome entre 50-100 MB de RAM para:
- Decodificação de vídeo
- Buffers internos de I/O
- Codec state

**Estimativa**: 6 processos × 75 MB médio = **450 MB**

### 2. Frame Buffers
Cada câmera tem um `frameChan` com buffer de 5 frames:
- cam1: 5 × 320 KB = 1.6 MB
- cam2: 5 × 64 KB = 0.32 MB
- cam3: 5 × 180 KB = 0.9 MB
- cam4: 5 × 115 KB = 0.58 MB
- cam5: 5 × 97 KB = 0.49 MB
- cam6: 0 MB (não ativa)

**Total frame buffers**: **~4 MB**

### 3. Local Buffer Pools
Cada câmera tem pool de 10 buffers de 512 KB cada:
- 6 câmeras × 10 buffers × 512 KB = **~30 MB**

Porém, com sync.Pool, nem todos buffers são alocados simultaneamente.

**Estimativa real**: **~10 MB**

### 4. AMQP Channels
Cada câmera tem seu próprio channel dedicado:
- Connection overhead: ~10 MB
- 6 channels × 5 MB médio = **~30 MB**

**Total AMQP**: **~42 MB**

### 5. Go Runtime Overhead
- Goroutines (12 total: 2 por câmera × 6): ~96 KB
- Maps, structs, Circuit Breakers: ~5 MB
- GC metadata: ~10 MB
- Stack frames: ~15 MB
- Misc runtime: ~20 MB

**Total runtime**: **~50 MB**

### 6. Memory Controller (novo em V2.2)
- Structs e maps: ~500 KB
- Callbacks: ~200 KB
- Stats tracking: ~300 KB

**Total Memory Controller**: **~1 MB** (negligível)

---

## 📊 Consumo Total Estimado

```
FFmpeg processes:       450 MB
Frame buffers:            4 MB
Local buffer pools:      10 MB
AMQP channels:           42 MB
Go runtime:              50 MB
Memory Controller:        1 MB
System overhead:          1 MB
─────────────────────────────
TOTAL:                  558 MB
```

**Com margem de segurança (+20%)**: **~670 MB**

---

## ⚙️ Configuração Atual vs Recomendada

### ❌ Configuração ATUAL (config.yaml)

```yaml
memory_controller:
  enabled: true
  max_memory_mb: 2048          # 2 GB
  warning_percent: 60.0        # 1229 MB
  critical_percent: 75.0       # 1536 MB
  emergency_percent: 85.0      # 1741 MB
  gc_trigger_percent: 70.0     # 1434 MB
```

**Problemas**:
- Max memory muito alto (2 GB) para consumo esperado de ~558 MB
- WARNING só dispara em 1229 MB (220% do esperado!) - muito tarde
- GC só dispara em 1434 MB (257% do esperado!) - muito tarde
- Desperdiça RAM do sistema

### ✅ Configuração RECOMENDADA (config.recommended.yaml)

```yaml
memory_controller:
  enabled: true
  max_memory_mb: 1024          # 1 GB
  warning_percent: 50.0        # 512 MB
  critical_percent: 70.0       # 716 MB
  emergency_percent: 85.0      # 870 MB
  gc_trigger_percent: 60.0     # 614 MB
```

**Vantagens**:
- Max memory apropriado (1 GB) com margem de ~80% acima do esperado
- WARNING aos 512 MB (92% do esperado) = detecção precoce
- GC aos 614 MB (110% do esperado) = proativo
- CRITICAL aos 716 MB (128% do esperado) = problema detectável
- EMERGENCY aos 870 MB (156% do esperado) = situação grave
- Mais eficiente para sistema com 6 câmeras

---

## 📈 Throughput Analysis

### Bandwidth por Câmera @ 15 FPS

| Câmera | Frame Size | Throughput |
|--------|------------|------------|
| cam1 | 320 KB | 4.8 MB/s |
| cam2 | 64 KB | 0.96 MB/s |
| cam3 | 180 KB | 2.7 MB/s |
| cam4 | 115 KB | 1.73 MB/s |
| cam5 | 97 KB | 1.46 MB/s |
| **TOTAL** | **776 KB** | **11.6 MB/s** |

### Network Bandwidth
- **Downstream** (câmeras → Edge Video): ~11.6 MB/s = ~93 Mbps
- **Upstream** (Edge Video → RabbitMQ): ~11.6 MB/s = ~93 Mbps

**Total**: ~186 Mbps bidirectional

---

## 🔍 Observações dos Logs

### 1. Circuit Breaker - cam6
```
[cam6] CB_OPEN - Frames: 0, Último: 2562047h47m16.854775807s atrás | CB: OPEN
```

**Problema**: cam6 nunca funcionou (URL inválida: `channel=banana`)
**Solução**: Corrigir URL ou remover câmera do config.yaml

### 2. Latência de Publicação
```
cam1: 548µs (excelente)
cam2: 0s - 542µs (excelente)
cam3: 559µs - 5.9s (!!! PICOS ALTOS)
cam4: 0s - 4.8s (!!! PICOS ALTOS)
cam5: 0s - 567µs (excelente)
```

**Problema identificado anteriormente**: publishMu serialization bottleneck
**Status**: Deferred para análise futura (conforme solicitação do usuário)

### 3. Reconexões Frequentes
```
[cam1] ERRO ao ler: EOF
[cam1] Tentando reconectar FFmpeg (estado: CLOSED)...
```

Todas as câmeras reconectam simultaneamente às 09:35:07.

**Possíveis causas**:
- Timeout de rede
- Instabilidade do stream
- Limite de conexões do servidor RTSP

**Circuit Breaker está protegendo corretamente** ✅

---

## 🎯 Recomendações

### 1. Ajustar Memory Controller (PRIORIDADE ALTA)
```bash
# Copiar config recomendada
cp config.recommended.yaml config.yaml
```

Ou editar manualmente:
```yaml
memory_controller:
  max_memory_mb: 1024       # Mudar de 2048 para 1024
  warning_percent: 50.0     # Mudar de 60.0 para 50.0
  gc_trigger_percent: 60.0  # Mudar de 70.0 para 60.0
```

### 2. Corrigir cam6 (PRIORIDADE MÉDIA)
Opção A: Corrigir URL
```yaml
- id: "cam6"
  url: "rtsp://pixforce:pixforce1234@186.193.228.105:12554/cam/realmonitor?channel=5&subtype=0"
```

Opção B: Remover câmera
```yaml
# Comentar ou deletar seção cam6
```

### 3. Monitorar Consumo Real (PRIORIDADE BAIXA)
```bash
# Executar por 5 minutos
.\bin\edge-video-v2.exe -config config.yaml

# Observar no relatório de stats:
# "Memory (Go Runtime): Alloc: XX MB"
# "Sistema (Processo): RAM Usage: XX MB"
```

Comparar valores reais com estimativa de 558 MB.

### 4. Ajustar JPEG Quality se necessário (OPCIONAL)
Se memória ainda alta:
```yaml
quality: 10  # Aumentar de 5 para 10 (frames menores)
```

**Impacto**:
- Reduz tamanho de frames em ~50-60%
- Reduz bandwidth de 11.6 MB/s para ~5-6 MB/s
- Reduz consumo de memória em ~100 MB

---

## 📝 Próximos Passos

1. ✅ Memory Controller implementado e testado
2. ⏳ Ajustar configuração conforme recomendado
3. ⏳ Testar consumo real por 5-10 minutos
4. ⏳ Verificar logs de WARNING/CRITICAL
5. ⏳ Ajustar thresholds se necessário
6. ⏳ (Futuro) Resolver bottleneck de publicação

---

**Conclusão**: Com **6 câmeras** (5 ativas), o consumo esperado é **~558 MB**. A configuração atual de `max_memory_mb: 2048` está **superestimada**. Recomenda-se ajustar para **1024 MB** para melhor eficiência.

**Arquivo criado**: `v2/config.recommended.yaml` com valores otimizados.
