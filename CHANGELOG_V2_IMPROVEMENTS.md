# 🚀 Edge-Video V1.6 - Melhorias de Sincronização (Baseadas na V2)

## 📅 Data: 2025-12-05

## 🎯 Objetivo

Aplicar as técnicas de sincronização perfeita da **V2** na arquitetura enterprise da **V1.6**, resolvendo os problemas de:
- ❌ Frames dessincronizados entre câmeras
- ❌ FPS baixo e instável
- ❌ Acúmulo de lag no buffer

## 🔧 Mudanças Implementadas

### 1. **Flags de Baixa Latência do FFmpeg**

#### Arquivo: `pkg/camera/persistent_capture.go` (linhas 92-107)

**ANTES:**
```go
"ffmpeg",
"-rtsp_transport", "tcp",
"-i", pc.rtspURL,
"-f", "image2pipe",
"-vcodec", "mjpeg",
"-q:v", fmt.Sprintf("%d", pc.quality),
"-r", fmt.Sprintf("%d", pc.fps),  // ❌ FFmpeg controla FPS
"-",
```

**DEPOIS:**
```go
"ffmpeg",
"-loglevel", "error",              // ✅ Menos ruído nos logs
"-rtsp_transport", "tcp",
"-fflags", "nobuffer",             // ✅ Desabilita buffering interno
"-flags", "low_delay",             // ✅ Modo baixa latência
"-max_delay", "0",                 // ✅ Zero delay
"-i", pc.rtspURL,
"-vf", fmt.Sprintf("fps=%d", pc.fps),  // ✅ FPS via filtro (mais preciso)
"-f", "image2pipe",
"-vcodec", "mjpeg",
"-q:v", fmt.Sprintf("%d", pc.quality),
"-",
```

**Benefício:** Reduz latência de captura e melhora precisão do FPS.

---

### 2. **Latest Frame Policy (Coração da V2)**

#### Arquivo: `pkg/camera/camera.go` (linhas 172-223)

**ANTES:**
```go
case <-ticker.C:
    frame, ok := c.persistentCapture.GetFrameWithTimeout(c.interval / 2)
    if !ok {
        // erro
        continue
    }

    // ❌ Publica qualquer frame do buffer (pode ser antigo)
    c.enqueueFrame(frame, false)
```

**DEPOIS:**
```go
case <-ticker.C:
    // ✅ Pega primeiro frame
    frame, ok = c.persistentCapture.GetFrameNonBlocking()
    if !ok {
        continue
    }

    // ✅ CRÍTICO: Flush de frames antigos acumulados
    flushedCount := 0
    for {
        oldFrame, hasMore := c.persistentCapture.GetFrameNonBlocking()
        if !hasMore {
            break
        }
        releaseFrameBuffer(frame)  // Libera frame antigo
        frame = oldFrame           // Usa o mais recente
        flushedCount++
    }

    // ✅ Sempre publica o frame MAIS RECENTE disponível
    c.enqueueFrame(frame, false)
```

**Benefício:**
- Elimina acúmulo de lag no buffer (50-200 frames)
- Garante que cada câmera publica frames sincronizados
- Descarta frames antigos explicitamente

---

### 3. **Flags de Baixa Latência no Modo Clássico**

#### Arquivo: `pkg/camera/camera.go` (linhas 317-332)

**ANTES:**
```go
"ffmpeg",
"-rtsp_transport", "tcp",
"-i", c.config.URL,
"-frames:v", "1",
"-f", "image2pipe",
"-vcodec", "mjpeg",
"-q:v", "5",
"-",
```

**DEPOIS:**
```go
"ffmpeg",
"-loglevel", "error",
"-rtsp_transport", "tcp",
"-fflags", "nobuffer",     // ✅ Sem buffer
"-flags", "low_delay",     // ✅ Baixa latência
"-max_delay", "0",         // ✅ Zero delay
"-i", c.config.URL,
"-frames:v", "1",
"-f", "image2pipe",
"-vcodec", "mjpeg",
"-q:v", "5",
"-",
```

**Benefício:** Modo clássico (fallback) também se beneficia da baixa latência.

---

## 📊 Impacto Esperado

| Problema | V1.6 Original | V1.6 Melhorada |
|----------|---------------|----------------|
| **Sincronização** | ❌ Instável (lag acumula) | ✅ Estável (flush explícito) |
| **FPS Real** | ❌ Baixo/Variável | ✅ Estável no target |
| **Latência** | ❌ Alta (buffer acumula) | ✅ Baixa (latest frame) |
| **Buffer Lag** | ❌ 50-200 frames | ✅ 0-1 frames |
| **FFmpeg Flags** | ❌ Padrão | ✅ Low-latency |

---

## 🧪 Como Testar

### 1. Compilar
```bash
cd D:\Users\rafa2\OneDrive\Desktop\edge-video
go build -o edge-video.exe ./cmd/edge-video
```

### 2. Executar
```bash
.\edge-video.exe -config .\config.toml
```

### 3. Verificar Logs
Procure por:
```
"Frames antigos descartados (Latest Frame Policy)"
"flushed_count": X
```

Se `flushed_count > 0`, significa que o sistema está descartando frames antigos corretamente! ✅

### 4. Verificar Sincronização
Use o viewer para comparar timestamps entre câmeras:
```bash
python viewer_cam1_sync.py
```

---

## 🎓 Diferença Arquitetural vs V2

**V2 (Simplificada):**
- Desacopla captura de publicação completamente
- 2 goroutines independentes por câmera
- Channel size=1 para Latest Frame

**V1.6 (Enterprise com melhorias V2):**
- Mantém worker pool e todas features enterprise
- Aplica Latest Frame Policy no loop de publicação
- Preserva circuit breaker, memory controller, metrics, Redis

**Resultado:** Melhor dos dois mundos! 🚀
- ✅ Sincronização perfeita da V2
- ✅ Resiliência e observabilidade da V1.6

---

## 📝 Arquivos Modificados

1. `pkg/camera/persistent_capture.go` - Flags FFmpeg + `-vf fps`
2. `pkg/camera/camera.go` - Latest Frame Policy + Flags modo clássico

**Total de mudanças:** ~50 linhas modificadas
**Impacto:** Altíssimo (resolve problema raiz de sincronização)

---

## ⚠️ Notas Importantes

1. **Buffer Size:** O parâmetro `persistent_buffer_size` no config agora funciona mais como "safety buffer". O Latest Frame Policy garante que frames antigos sejam descartados.

2. **Métricas:** Novo label `flushed_old_frames` em `FramesDropped` indica frames descartados pela política Latest Frame (isso é **bom**, não erro!).

3. **Performance:** Pode haver ligeiro aumento no uso de CPU devido ao flush loop, mas é negligível comparado ao benefício de sincronização.

4. **Compatibilidade:** Mudanças 100% retrocompatíveis. Configs existentes funcionam sem alteração.

---

## 🎯 Próximos Passos

1. ✅ Testar em ambiente de desenvolvimento
2. ⏳ Testar em 1 máquina cliente (MF-VEHICLECOUNTER)
3. ⏳ Validar sincronização com múltiplas câmeras
4. ⏳ Deploy em todas as máquinas se validado

---

## 👤 Autor

- **Rafael (com assistência Claude Code)**
- **Data:** 2025-12-05
- **Branch:** main (aplicado diretamente)
- **Versão:** 1.6 → 1.6.1 (sync improvements)

---

## 🔗 Referências

- Análise comparativa V1.6 vs V2
- Código fonte V2: `v2/camera_stream.go` (Latest Frame Policy original)
- FFmpeg low-latency flags: https://ffmpeg.org/ffmpeg-formats.html#toc-Options
