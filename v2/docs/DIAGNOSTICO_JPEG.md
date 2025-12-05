# Diagnóstico: Falhas de Decodificação JPEG

## 🎯 Descoberta Crítica

Após análise profunda do problema de decodificação, descobri que:

**O problema NÃO é o stripping de metadata nem o formato JPEG em si!**

### Evidências

✅ **Teste offline bem-sucedido:**
- PIL/Pillow decodifica JPEG original: **SUCCESS**
- PIL/Pillow decodifica JPEG stripped: **SUCCESS**
- OpenCV decodifica JPEG original: **SUCCESS**
- OpenCV decodifica JPEG stripped: **SUCCESS**

❌ **Problema em tempo real:**
- OpenCV falha em **46.9%** dos frames durante consumo do RabbitMQ
- Todos os frames têm metadata FFmpeg (100%)
- Zero vazamentos de routing key (0%)

### Conclusão

**Os frames estão chegando CORROMPIDOS ou INCOMPLETOS do RabbitMQ!**

Possíveis causas:
1. 📡 Perda de pacotes de rede (Edge Video → RabbitMQ ou RabbitMQ → Viewer)
2. ⚡ Race condition no publisher (frames publicados antes de estarem completos)
3. 📦 Limite de tamanho de mensagem do RabbitMQ causando truncamento
4. 🎥 FFmpeg gerando alguns frames malformados

## 🔬 Próximo Passo: Verificar Integridade

Atualizei o `viewer_cam1_sync.py` para verificar se os JPEGs estão chegando **completos** (com marcador EOF `FFD9`).

### Executar Teste

```bash
cd D:\Users\rafa2\OneDrive\Desktop\edge-video\v2
python viewer_cam1_sync.py cam2
```

### O que esperar

O viewer agora mostra para os primeiros 5 frames:

```
[INTEGRITY #1] Size=58575, ✓ EOF OK, EOF em -2 bytes (0 bytes padding), Last4=12ffd9
[INTEGRITY #2] Size=57925, ✗ SEM EOF, EOF NÃO ENCONTRADO nos últimos 20 bytes, Last4=00000000
```

Isso nos dirá se:
- ✅ **EOF OK**: JPEG está completo (problema pode ser outro)
- ❌ **SEM EOF**: JPEG está truncado/corrompido (problema de transmissão)

## 📊 Resultados Esperados

**Se frames têm EOF mas falham:**
→ Problema é compatibilidade OpenCV vs FFmpeg MJPEG
→ **Solução**: Usar PIL/Pillow para decodificação

**Se frames NÃO têm EOF:**
→ Problema é corrupção/truncamento na transmissão
→ **Soluções possíveis:**
  - Aumentar `max_message_size` no RabbitMQ
  - Adicionar flush do pipe FFmpeg antes de publicar
  - Adicionar verificação de integridade no publisher
  - Usar protocolo com checksums (AMQP message properties)

## 🚀 Implementação Alternativa: Decoder PIL

Se o problema for OpenCV, posso implementar fallback para PIL:

```python
# Tenta OpenCV primeiro (mais rápido)
img = cv2.imdecode(np_arr, cv2.IMREAD_COLOR)

# Se falhar, usa PIL (mais permissivo)
if img is None:
    from PIL import Image
    import io
    img_pil = Image.open(io.BytesIO(cleaned_body))
    img = np.array(img_pil)
    img = cv2.cvtColor(img, cv2.COLOR_RGB2BGR)
```

Isso daria **100% de taxa de decodificação** se PIL consegue ler o que OpenCV rejeita.

---

**Status**: Aguardando resultado do teste de integridade para definir próxima ação.
