# 📹 test_all_cameras.bat - Guia de Uso

## 🎯 Propósito

Script batch para Windows que abre **6 viewers simultaneamente**, cada um em seu próprio terminal, para testar todas as câmeras ao mesmo tempo.

**Perfeito para**:
- ✅ Validar que **não há contaminação entre câmeras**
- ✅ Stress test com múltiplas câmeras
- ✅ Verificar isolamento completo de buffers
- ✅ Monitorar performance com carga máxima

---

## 🚀 Como Usar

### 1. Pré-requisitos

- ✅ `edge-video-v2.exe` rodando
- ✅ Python instalado e no PATH
- ✅ `pika` e `opencv-python` instalados:
  ```bash
  pip install pika opencv-python pillow
  ```
- ✅ Conexão com RabbitMQ funcionando

### 2. Executar o Script

**Opção 1**: Duplo clique no arquivo `test_all_cameras.bat`

**Opção 2**: Via terminal:
```bash
cd D:\Users\rafa2\OneDrive\Desktop\edge-video\v2
.\test_all_cameras.bat
```

### 3. Resultado Esperado

O script abrirá **6 janelas de terminal**, uma para cada câmera:

```
┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐
│   CAM1 Viewer       │  │   CAM2 Viewer       │  │   CAM3 Viewer       │
│                     │  │                     │  │                     │
│ [Frame display]     │  │ [Frame display]     │  │ [Frame display]     │
│                     │  │                     │  │                     │
│ Stats: FPS, Size... │  │ Stats: FPS, Size... │  │ Stats: FPS, Size... │
└─────────────────────┘  └─────────────────────┘  └─────────────────────┘

┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐
│   CAM4 Viewer       │  │   CAM5 Viewer       │  │   CAM6 Viewer       │
│                     │  │                     │  │                     │
│ [Frame display]     │  │ [Frame display]     │  │ [Frame display]     │
│                     │  │                     │  │                     │
│ Stats: FPS, Size... │  │ Stats: FPS, Size... │  │ Stats: FPS, Size... │
└─────────────────────┘  └─────────────────────┘  └─────────────────────┘
```

Cada janela mostra:
- Nome da câmera no título da janela
- Frame da câmera sendo exibido
- Estatísticas em tempo real

---

## ✅ Validação de Sucesso

**O que verificar**:

1. **Cada viewer mostra APENAS sua câmera**:
   - cam1 → viewer CAM1
   - cam2 → viewer CAM2
   - ...

2. **Zero vazamentos nos logs**:
   ```
   [VAZAMENTO ROUTING] → NENHUM ✅
   [VAZAMENTO HEADER] → NENHUM ✅
   [RESOLUÇÃO INVÁLIDA] → NENHUM ✅
   ```

3. **FPS estável em todos os viewers** (~10-15 FPS)

4. **Sem erros de decodificação** (ou < 5%)

**Se TODOS os itens acima estiverem OK: Sistema está 100% funcionando!** 🎉

---

## 🛑 Como Fechar os Viewers

**Opção 1**: Fechar cada janela individualmente (clique no X)

**Opção 2**: Em cada terminal, pressione `Ctrl+C`

**Opção 3**: Task Manager → Finalizar todos os processos `python.exe` relacionados

---

## 🔧 Troubleshooting

### Erro: "python não é reconhecido"

**Solução**: Python não está no PATH

```bash
# Edite test_all_cameras.bat na linha 11:
set PYTHON=C:\Python39\python.exe  # Ajuste para seu caminho
```

### Erro: "No module named 'pika'"

**Solução**: Instalar dependências

```bash
pip install pika opencv-python pillow
```

### Viewers não abrem

**Causa**: `edge-video-v2.exe` não está rodando

**Solução**: Em outro terminal, execute:
```bash
.\edge-video-v2.exe
```

### Apenas algumas câmeras aparecem

**Causa**: Algumas câmeras podem estar offline ou com URL errado

**Solução**: Verifique logs do `edge-video-v2.exe` para ver quais câmeras falharam ao conectar

### Janelas abrem mas ficam em branco

**Causa**: RabbitMQ não está publicando frames

**Verificar**:
1. `edge-video-v2.exe` está rodando?
2. Conexão com RabbitMQ está OK?
3. Câmeras estão conectadas? (veja logs)

---

## 📊 Dicas de Performance

### Organizar as Janelas

Windows 11/10 permite organizar janelas em grade:

1. Arraste uma janela para canto superior esquerdo
2. Arraste outra para canto superior direito
3. Repita para as outras 4 em baixo
4. Resultado: Grade 3x2 perfeita para monitorar todas!

### Monitorar Performance

Deixe rodando por 5-10 minutos e observe:

- **FPS médio**: Deve ficar em ~10-15 FPS
- **Taxa de erro**: Deve ser < 5%
- **Contaminação**: Deve ser ZERO ✅

### Stress Test de Longo Prazo

Para teste de estabilidade, deixe rodando por 30+ minutos:

```bash
# Deixe edge-video-v2.exe rodando
# Deixe test_all_cameras.bat rodando
# Monitore memory usage no Task Manager
# Memory deve ficar estável (~120MB)
```

---

## 🎓 O Que Este Teste Valida

✅ **Isolamento de Buffers**: Cada câmera tem seu próprio buffer pool

✅ **Zero Race Conditions**: Não há compartilhamento de memória entre câmeras

✅ **Routing Correto**: RabbitMQ entrega cada frame para o viewer correto

✅ **Headers AMQP**: Metadados corretos em todas as mensagens

✅ **Performance Multi-Camera**: Sistema escala para 6+ câmeras

✅ **Memory Safety**: Sem leaks, uso estável de RAM

---

## 📝 Logs de Exemplo

**Saída esperada em cada viewer**:

```
================================================================================
VISUALIZADOR SINCRONIZADO - cam1
================================================================================
RabbitMQ: 34.71.212.239:5672
VHost: supercarlao_rj_mercado
Exchange: supercarlao_rj_mercado.exchange
Routing Key: supercarlao_rj_mercado.cam1
Queue: supercarlao_rj_mercado.cam1.viewer
================================================================================

[RECV #1] RoutingKey=supercarlao_rj_mercado.cam1, Header[camera_id]=cam1, Size=329563
[RECV #2] RoutingKey=supercarlao_rj_mercado.cam1, Header[camera_id]=cam1, Size=331245
[RECV #3] RoutingKey=supercarlao_rj_mercado.cam1, Header[camera_id]=cam1, Size=328891
...

Estatísticas (10s):
  - Frames recebidos: 142
  - Frames com metadata removido: 142 (100.0%)
  - Erros de decodificação: 6 (4.2%)
  - FPS médio: 14.2
  - Taxa de descarte: 0.0%
```

**Nenhum vazamento ou erro de validação = SUCESSO!** ✅

---

## 🚀 Para Produção

Este script é apenas para **testes e validação**.

Em produção, você terá:
- Viewers rodando em máquinas separadas
- Load balancing de viewers
- Monitoring centralizado
- Alertas automáticos

Mas o **princípio é o mesmo**: cada viewer consome de uma routing key dedicada e valida isolamento completo.

---

**Última atualização**: Dezembro 2024
**Versão**: V2.1 (Post-Bug-Fix)
**Status**: ✅ TESTADO E FUNCIONANDO
