# Edge Video - Sistema de Captura e Distribuição de Vídeo


![Go Tests](https://github.com/T3-Labs/edge-video/actions/workflows/go-test.yml/badge.svg)
![Docker Build](https://github.com/T3-Labs/edge-video/actions/workflows/build-and-push.yml/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

---

## Sobre o Edge Video

O **Edge Video** é uma plataforma distribuída para captura, processamento e distribuição de vídeo de câmeras RTSP/IP, projetada para ambientes de edge computing, multi-tenant e integração com sistemas de IA, monitoramento e automação.

---

## Principais Features

- **Multi-Câmera RTSP/IP**: Captura simultânea de múltiplas câmeras.
- **Isolamento Multi-Tenant (RabbitMQ vhost)**: Cada cliente tem seu próprio namespace, sem colisão de dados.
- **Chave Redis Otimizada (Unix Nanoseconds)**: Chaves compactas, ordenáveis e com queries temporais eficientes.
- **Distribuição via RabbitMQ (AMQP) e MQTT**: Flexibilidade para diferentes integrações.
- **Buffer Circular, Worker Pool e Circuit Breaker**: Controle de memória, fila de processamento e proteção contra overflow/falhas.
- **Publicação de Metadados**: Eventos JSON leves para consumidores, detalhando cada frame.
- **Armazenamento Opcional em Redis**: TTL configurável, queries rápidas e compatibilidade multi-tenant.
- **Configuração Flexível via TOML/YAML**: Adição/remoção de câmeras, tuning de parâmetros, ativação de recursos.
- **Instalador Windows (NSIS)**: Instalação como serviço, auto-start, logs, gerenciamento via Services.msc e CLI.
- **Containerização Completa (Docker/Docker Compose)**: Deploy simplificado, integração com RabbitMQ e Redis.
- **Consumer Python com OpenCV**: Visualização em grid, integração fácil para IA e monitoramento.
- **Changelog Automatizado (Towncrier)**: Fragments, changelog por release, integração com pre-commit.
- **Pre-commit Hooks**: Lint, formatação, validação de configs e commits semânticos.
- **Documentação Detalhada**: Arquitetura, exemplos, troubleshooting, guias de migração e multi-tenancy.

---

## Como Usar

### 1. Configuração
Edite `config.toml` ou `config.yaml` para suas câmeras e parâmetros. Exemplo:

```toml
[amqp]
amqp_url = "amqp://user:pass@rabbitmq:5672/meu-cliente"
exchange = "cameras"
routing_key_prefix = "camera."

[redis]
enabled = true
address = "redis:6379"
ttl_seconds = 300
prefix = "frames"

[[cameras]]
id = "cam1"
url = "rtsp://admin:pass@192.168.1.100:554/stream1"
```

### 2. Execução

**Go:**
```bash
go build -o edge-video ./cmd/edge-video
./edge-video --config config.toml
```

**Docker Compose:**
```bash
docker-compose up -d --build
```

**Instalador Windows:**
- Baixe o instalador no GitHub Releases.
- Instale como serviço via assistente ou CLI.
- Gerencie pelo Services.msc ou comandos `net start/stop EdgeVideoService`.

### 3. Monitoramento
- RabbitMQ UI: `http://localhost:15672`
- Logs locais: `logs/` ou Event Viewer (Windows)
- Métricas: Prometheus em `:2112/metrics`

### 4. Integração
- Consuma metadados e frames via Python, Go ou qualquer linguagem compatível com AMQP/MQTT/Redis.
- Exemplo Python:
```python
import pika, redis, json
def callback(ch, method, properties, body):
  metadata = json.loads(body)
  frame = redis_client.get(metadata['redis_key'])
```

---

## Troubleshooting

- Verifique logs locais e Event Viewer.
- Use comandos de serviço para instalar, iniciar, parar e desinstalar.
- Consulte a documentação para migração de chaves Redis e multi-tenancy.

---

## Contribuição

1. Fork, branch, changelog fragment, commit semântico, PR.
2. Use pre-commit hooks para garantir qualidade.

---

## Documentação Avançada

- [docs/windows/README.md](docs/windows/README.md): Instalação e uso no Windows
- [docs/vhost-based-identification.md](docs/vhost-based-identification.md): Multi-tenancy e isolamento
- [docs/features/redis-storage.md](docs/features/redis-storage.md): Detalhes do armazenamento Redis
- [docs/features/message-queue.md](docs/features/message-queue.md): Integração RabbitMQ/MQTT
- [docs/features/metadata.md](docs/features/metadata.md): Estrutura de metadados
- [docs/changelog.md](docs/changelog.md): Histórico de mudanças

---

## Licença

MIT

---

**Desenvolvido por T3 Labs** 🚀

## 📋 Objetivo do Projeto

O **Edge Video** é um sistema distribuído de captura e streaming de câmeras RTSP, projetado para ambientes de edge computing. O sistema captura frames de múltiplas câmeras IP em tempo real, processa-os e distribui através de uma fila de mensagens (RabbitMQ), permitindo que múltiplos consumidores recebam e processem os streams de vídeo de forma escalável e eficiente.

## ⚠️ Breaking Changes - v1.2.0 (Unreleased)

**Migração de Formato de Chaves Redis** - Mudança para Unix Nanoseconds

A partir da versão 1.2.0, o formato de chaves Redis foi otimizado para melhor performance:

**Formato Anterior:** `frames:{vhost}:{cameraID}:{RFC3339_timestamp}:{sequence}`  
**Formato Novo:** `{vhost}:{prefix}:{cameraID}:{unix_nano}:{sequence}`

**Impacto:**
- ⚠️ Chaves antigas no Redis não serão mais compatíveis
- 🔄 **Ação Requerida**: FLUSHDB no Redis, aguardar TTL expirar ou executar script de migração

**Benefícios:**
- ⚡ 36% mais compacto (19 vs 30 caracteres)
- 🚀 10x mais rápido em comparações
- 📊 Sortable naturalmente (ordem cronológica nativa)
- 🔍 Range queries extremamente eficientes

📚 Veja [docs/vhost-based-identification.md](docs/vhost-based-identification.md) para guia de migração completo.

## 🎯 Principais Funcionalidades

- **Captura Multi-Câmera**: Suporta a captura simultânea de múltiplas câmeras RTSP/IP
- **Multi-Tenant (Vhost-Based)**: Isolamento completo de dados por cliente usando RabbitMQ vhosts
- **Processamento em Edge**: Processamento local dos frames antes da transmissão
- **Distribuição via Message Broker**: Utiliza RabbitMQ com protocolo AMQP para distribuição eficiente
- **Cache Redis Otimizado**: Armazenamento de frames com TTL e formato de chave ultra-eficiente
- **Visualização em Grid**: Interface Python para visualização de todas as câmeras em uma única janela
- **Configuração Flexível**: Fácil adição/remoção de câmeras via arquivo TOML
- **Containerizado**: Deploy simplificado com Docker e Docker Compose

## 🏗️ Arquitetura

```
┌─────────────────┐
│  Câmeras RTSP   │
│  (5 câmeras)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Camera         │
│  Collector      │  ← Aplicação Go
│  (FFmpeg)       │
└────────┬────────┘
         │ JPEG Frames
         ▼
┌─────────────────┐
│   RabbitMQ      │
│   (AMQP)        │
│   Exchange:     │
│   cameras       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│                 │
│    Consumer     │  ← Visualização em Grid 2x3
│                 │
└─────────────────┘
```

## � Código Refatorado

Este repositório foi refatorado seguindo as melhores práticas de desenvolvimento Python:

### **Estrutura Refatorada:**
```
src/
├── config/
│   └── config_manager.py      # Gerenciamento de configuração
├── consumer/
│   └── rabbitmq_consumer.py   # Consumidor RabbitMQ
├── display/
│   ├── display_manager.py     # Gerenciador de display OpenCV
│   └── video_processor.py     # Processamento de frames
└── video_consumer_app.py      # Aplicação principal

tests/
├── test_config_manager.py
├── test_rabbitmq_consumer.py
├── test_display_manager.py
├── test_video_processor.py
└── test_video_consumer_app.py
```

### **Principais Melhorias:**
- **Single Responsibility Principle**: Cada classe tem uma responsabilidade específica
- **Separação de Concerns**: Lógica de negócio separada da apresentação
- **Testabilidade**: 100% de cobertura de testes unitários
- **Type Hints**: Tipagem completa para melhor manutenibilidade
- **Performance Otimizada**: Formato de chaves Redis ultra-eficiente com Unix nanoseconds

### **Otimizações de Performance (v1.2.0):**

#### 🚀 Redis Key Format Optimization
O sistema foi otimizado para usar Unix nanoseconds ao invés de RFC3339 timestamps:

**Comparação de Performance:**

| Métrica | RFC3339 | Unix Nano | Melhoria |
|---------|---------|-----------|----------|
| Tamanho da chave | 30 caracteres | 19 dígitos | **36% menor** |
| Tipo de comparação | String parsing | Integer comparison | **10x mais rápido** |
| Sortabilidade | Lexicográfica | Numérica nativa | **Natural** |
| Range queries | Parsing + comparação | `>= start AND <= end` | **Extremamente eficiente** |
| Overhead de memória (1M chaves) | ~30 MB | ~19 MB | **-11 MB** |

**Exemplo de Chave:**
```redis
# Formato Otimizado (Novo)
supermercado_vhost:frames:cam4:1731024000123456789:00001

# Formato Anterior (Deprecated)
frames:supermercado_vhost:cam4:2024-11-07T19:30:00.123456789Z:00001
```

**Benefícios Práticos:**
- ✅ Menor uso de memória Redis em deployments com milhões de chaves
- ✅ Queries temporais (range) executam 10x mais rápido
- ✅ Ordenação cronológica natural sem conversões
- ✅ Compatível com ferramentas de análise de séries temporais
- ✅ Facilita agregações e análises de dados históricos
- **Documentação**: Docstrings detalhadas seguindo padrões Python

### **Como usar o código refatorado:**
```bash
# Instalar dependências
uv sync --dev

# Executar testes
uv run pytest

# Executar aplicação refatorada
uv run python main_refactored.py

# Executar linting
uv run ruff check src/
uv run ruff format src/
```

## �🛠️ Tecnologias Utilizadas

### Backend (Collector)
- **Go 1.24**: Linguagem principal para o collector
- **FFmpeg**: Captura de frames das câmeras RTSP
- **Viper**: Gerenciamento de configuração
- **AMQP (streadway/amqp)**: Cliente RabbitMQ
- **Redis**: Cache de frames com chaves otimizadas (Unix nanoseconds)

### Message Broker & Storage
- **RabbitMQ 3.13**: Sistema de mensageria para distribuição de frames
- **Redis 7.x**: Cache de frames com TTL e multi-tenancy via vhost isolation

### Frontend (Consumer)
- **Python 3.11+**: Linguagem para o consumer
- **OpenCV**: Processamento e visualização de vídeo
- **Pika**: Cliente RabbitMQ para Python
- **NumPy**: Manipulação de arrays para concatenação de frames

### Infraestrutura
- **Docker & Docker Compose**: Containerização e orquestração
- **Alpine Linux**: Imagem base leve para containers
- **GitHub Actions**: CI/CD para testes e builds automatizados

## 📦 Estrutura do Projeto

```
edge-video/
├── config.toml              # Configuração das câmeras e parâmetros
├── docker-compose.yml       # Orquestração dos serviços
├── Dockerfile              # Build da aplicação Go
├── go.mod                  # Dependências Go
├── cmd/
│   └── edge-video/
│       └── main.go         # Entrypoint da aplicação
├── pkg/
│   ├── camera/
│   │   └── camera.go       # Lógica de captura de frames
│   ├── mq/
│   │   ├── publisher.go    # Interface do publisher
│   │   ├── amqp.go         # Implementação AMQP
│   │   └── mqtt.go         # Implementação MQTT (alternativa)
│   ├── config/
│   │   ├── config.go       # Carregamento de configuração
│   │   └── config_test.go  # Testes de configuração
│   └── util/
│       └── compress.go     # Utilitários de compressão
├── internal/
│   ├── storage/
│   │   ├── key_generator.go       # Gerador de chaves Redis otimizado
│   │   └── key_generator_test.go  # Testes do gerador (16 testes)
│   └── metadata/
│       └── publisher.go    # Publisher de metadados
├── docs/
│   ├── changelog.md                    # Changelog do projeto
│   ├── vhost-based-identification.md   # Guia de multi-tenancy
│   └── PRECOMMIT_TOWNCRIER_GUIDE.md   # Guia de contribuição
├── test_consumer.py         # Consumer Python com visualização
└── README.md               # Este arquivo
```

## 🚀 Como Executar

### Pré-requisitos

- Docker e Docker Compose instalados
- Python 3.11+ (para o consumer)
- UV (gerenciador de pacotes Python) ou pip

### 1. Configure as Câmeras

Edite o arquivo `config.toml` e adicione as URLs das suas câmeras:

```toml
[[cameras]]
id = "cam1"
url = "rtsp://user:pass@192.168.1.100:554/stream"

[[cameras]]
id = "cam2"
url = "rtsp://user:pass@192.168.1.101:554/stream"

# ... até 6 câmeras
```

### 2. Executar a Aplicação

#### Usando arquivo de configuração padrão

```bash
# Compilar e executar
go build -o edge-video ./cmd/edge-video
./edge-video

# Ou executar diretamente
go run ./cmd/edge-video
```

#### Usando arquivo de configuração customizado

```bash
# Especificar arquivo via parâmetro --config
./edge-video --config /path/to/custom-config.toml

# Ou com go run
go run ./cmd/edge-video --config config.test.toml
```

#### Validar configuração

```bash
# Validar arquivo de configuração
go run ./cmd/validate-config --config config.toml

# Ver ajuda
./edge-video --help
# Output:
#   -config string
#         Caminho para o arquivo de configuração (default "config.toml")
```

### 3. Inicie os Serviços com Docker

#### Opção A: Usando Docker Compose (Recomendado)

```bash
docker-compose up -d --build
```

Isso iniciará:
- **RabbitMQ**: Porta 5672 (AMQP) e 15672 (Management UI)
- **Camera Collector**: Aplicação Go capturando e publicando frames

#### Opção B: Usando Docker Run (Após Docker Pull)

Se você baixou a imagem do Docker Hub com `docker pull`:

```bash
# 1. Inicie o RabbitMQ primeiro
docker run -d \
  --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=user \
  -e RABBITMQ_DEFAULT_PASS=password \
  -e RABBITMQ_DEFAULT_VHOST=guard_vhost \
  rabbitmq:3.13-management-alpine

# 2. Baixe a imagem do Edge Video (se ainda não tiver)
docker pull t3labs/edge-video:latest

# 3. Execute o Camera Collector com seu config.toml local
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v /path/absoluto/para/seu/config.toml:/app/config.toml \
  t3labs/edge-video:latest
```

**Exemplos de caminhos para o volume:**

```bash
# Exemplo 1: Config.toml na pasta atual
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v $(pwd)/config.toml:/app/config.toml \
  t3labs/edge-video:latest

# Exemplo 2: Config.toml em /etc
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v /etc/edge-video/config.toml:/app/config.toml \
  t3labs/edge-video:latest

# Exemplo 3: Config.toml no home do usuário
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v $HOME/.config/edge-video/config.toml:/app/config.toml \
  t3labs/edge-video:latest

# Exemplo 4: Config.toml em storage montado
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v /mnt/storage/configs/cameras.toml:/app/config.toml \
  t3labs/edge-video:latest
```

**Usando Docker Network (Melhor prática):**

```bash
# 1. Crie uma rede
docker network create edge-video-net

# 2. Inicie o RabbitMQ na rede
docker run -d \
  --name rabbitmq \
  --network edge-video-net \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=user \
  -e RABBITMQ_DEFAULT_PASS=password \
  -e RABBITMQ_DEFAULT_VHOST=guard_vhost \
  rabbitmq:3.13-management-alpine

# 3. Execute o Camera Collector na mesma rede
docker run -d \
  --name camera-collector \
  --network edge-video-net \
  -v /path/para/seu/config.toml:/app/config.toml \
  t3labs/edge-video:latest
```

### 3. Execute o Consumer Python

```bash
# Com UV
uv run test_consumer.py

# Ou com pip
pip install -r requirements.txt
python test_consumer.py
```

### 4. Visualize as Câmeras

Uma janela será aberta mostrando todas as câmeras em uma grade 2x3.

**Pressione 'q' para sair.**

## ⚙️ Configuração

### config.toml

```toml
interval_ms = 500                    # Intervalo entre capturas (ms)
protocol = "amqp"                    # Protocolo: amqp ou mqtt
process_every_n_frames = 3           # Reduz taxa de frames (1 a cada 3)

[amqp]
amqp_url = "amqp://user:password@rabbitmq:5672/guard_vhost"
exchange = "cameras"
routing_key_prefix = "camera"

[compression]
enabled = false                      # Compressão zstd (desabilitada)
level = 3

[[cameras]]
id = "cam1"
url = "rtsp://..."

[[cameras]]
id = "cam2"
url = "rtsp://..."
```

### 🔄 Optional Redis Frame Storage + Metadata

You can enable Redis frame caching and metadata publishing by updating `config.toml`:

```toml
[redis]
enabled = true
address = "redis:6379"
ttl_seconds = 300
prefix = "frames"

[metadata]
enabled = true
exchange = "camera.metadata"
routing_key = "camera.metadata.event"
```

When enabled:

- Frames are stored in Redis with TTL
- Metadata messages are sent asynchronously to RabbitMQ
- Existing video streaming and publishing are unaffected

### 🏢 Isolamento Multi-Cliente (Multi-tenancy)

O Edge Video usa o **vhost do RabbitMQ** como identificador único de cliente, garantindo isolamento automático de dados no Redis.

#### Formato de Chave Redis

```
{vhost}:{prefix}:{cameraID}:{unix_timestamp_nano}:{sequence}
```

**Exemplo:**
```redis
supermercado_vhost:frames:cam4:1731024000123456789:00001
```

#### Como Funciona

1. **Vhost Extraído Automaticamente**: O vhost é extraído da URL AMQP configurada
2. **Unix Nanoseconds**: Timestamps numéricos para sortabilidade e performance
3. **Chaves Redis Isoladas**: Cada cliente possui namespace próprio no Redis
4. **Zero Configuração Adicional**: Não é necessário configurar `instance_id` separadamente

#### Exemplo: Múltiplos Clientes

```toml
# Cliente A (config-client-a.toml)
[amqp]
amqp_url = "amqp://user:pass@rabbitmq:5672/client-a"

# Cliente B (config-client-b.toml) 
[amqp]
amqp_url = "amqp://user:pass@rabbitmq:5672/client-b"
```

**Resultado no Redis:**
```redis
client-a:frames:cam1:1731024000123456789:00001
client-b:frames:cam1:1731024000123456789:00001
```

#### Por que Unix Timestamp?

| Aspecto | RFC3339 | Unix Nano | Vantagem |
|---------|---------|-----------|----------|
| **Tamanho** | 30 chars | 19 dígitos | ✅ 36% menor |
| **Sortable** | String | Numérico | ✅ Natural |
| **Comparação** | Parsing | Inteiro | ✅ 10x mais rápido |
| **Range Query** | Complexo | Simples | ✅ `>= start AND <= end` |

**Benefícios:**
- ✅ Impossível colisão entre clientes diferentes
- ✅ Mesmas câmeras em clientes diferentes não conflitam
- ✅ Timestamps compactos e sortable numericamente
- ✅ Range queries extremamente eficientes
- ✅ Alinhamento com arquitetura AMQP (vhost = multi-tenancy)

📚 **Documentação Completa**: Veja [docs/vhost-based-identification.md](docs/vhost-based-identification.md) para detalhes de implementação, exemplos de deployment e troubleshooting.

## 🔍 Monitoramento

### RabbitMQ Management UI

Acesse: `http://localhost:15672`
- **Usuário**: user
- **Senha**: password

### Logs do Collector

```bash
docker logs camera-collector -f
```

### Métricas do Sistema

Verifique o throughput de mensagens e o uso de recursos no RabbitMQ Management.

## 📊 Casos de Uso

1. **Vigilância e Segurança**: Monitoramento em tempo real de múltiplas câmeras
2. **Análise de Vídeo**: Processamento de frames para detecção de objetos, pessoas, etc.
3. **Edge Computing**: Processamento local antes de envio para a nuvem
4. **Sistemas de Visão Computacional**: Pipeline para aplicações de Computer Vision
5. **Armazenamento Inteligente**: Gravação seletiva baseada em eventos

## 🔧 Desenvolvimento

### Adicionar Nova Câmera

1. Edite `config.toml`
2. Adicione a nova entrada em `[[cameras]]`
3. Reinicie o container: `docker-compose restart camera-collector`

### Modificar Taxa de Frames

Ajuste `interval_ms` no `config.toml` para controlar a taxa de captura.

### Habilitar Compressão

```toml
[compression]
enabled = true
level = 3  # 1-22 (maior = mais compressão)
```

### Habilitar Redis e Metadata

```toml
[redis]
enabled = true
address = "redis:6379"
password = ""  # Opcional
ttl_seconds = 300
prefix = "frames"

[metadata]
enabled = true
exchange = "camera.metadata"
routing_key = "camera.metadata.event"
```

## 🤝 Contribuindo

Este é um projeto da **T3 Labs**. Para contribuir:

1. Fork o repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. **Crie um changelog fragment** para suas mudanças:
   ```bash
   ./scripts/new-changelog.sh feature "Descrição da mudança"
   ```
4. Commit suas mudanças usando [commits semânticos](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "feat: adiciona nova funcionalidade"
   ```
5. Push para a branch (`git push origin feature/nova-funcionalidade`)
6. Abra um Pull Request

### 📝 Sistema de Changelog

Este projeto usa [Towncrier](https://towncrier.readthedocs.io/) para gerenciar o changelog automaticamente.

**Criar um fragment:**
```bash
# Usando o script helper (recomendado)
./scripts/new-changelog.sh feature "Adiciona suporte a PostgreSQL"

# Ou manualmente
echo "Adiciona suporte a PostgreSQL" > changelog.d/$(date +%s).feature.md
```

**Tipos disponíveis:** `feature`, `bugfix`, `docs`, `removal`, `security`, `performance`, `refactor`, `misc`

**Gerar changelog para release:**
```bash
# Preview
./scripts/build-changelog.sh --draft 1.0.0

# Gerar
./scripts/build-changelog.sh 1.0.0
```

Para mais detalhes, veja [docs/PRECOMMIT_TOWNCRIER_GUIDE.md](docs/PRECOMMIT_TOWNCRIER_GUIDE.md)

### 🔍 Pre-commit Hooks

Este projeto usa pre-commit hooks para garantir qualidade:

```bash
# Instalar hooks
pip install pre-commit towncrier
pre-commit install
pre-commit install --hook-type commit-msg

# Executar manualmente
pre-commit run --all-files
```

Os hooks verificam:
- ✅ Formatação de código Go (gofmt, goimports)
- ✅ Lint (go vet, golangci-lint)
- ✅ Changelog fragments (towncrier)
- ✅ Formato de commits (commitizen)
- ✅ Detecção de segredos
- ✅ Validação de YAML/TOML/JSON

## 📝 Licença

Este projeto está sob a licença MIT.

## 🔗 Links

- **Repositório**: https://github.com/T3-Labs/edge-video
- **RabbitMQ**: https://www.rabbitmq.com/
- **FFmpeg**: https://ffmpeg.org/
- **OpenCV**: https://opencv.org/

---

**Desenvolvido por T3 Labs** 🚀

## 📋 Objetivo do Projeto

O **Edge Video** é um sistema distribuído de captura e streaming de câmeras RTSP, projetado para ambientes de edge computing. O sistema captura frames de múltiplas câmeras IP em tempo real, processa-os e distribui através de uma fila de mensagens (RabbitMQ), permitindo que múltiplos consumidores recebam e processem os streams de vídeo de forma escalável e eficiente.

## ⚠️ Breaking Changes - v1.2.0 (Unreleased)

**Migração de Formato de Chaves Redis** - Mudança para Unix Nanoseconds

A partir da versão 1.2.0, o formato de chaves Redis foi otimizado para melhor performance:

**Formato Anterior:** `frames:{vhost}:{cameraID}:{RFC3339_timestamp}:{sequence}`  
**Formato Novo:** `{vhost}:{prefix}:{cameraID}:{unix_nano}:{sequence}`

**Impacto:**
- ⚠️ Chaves antigas no Redis não serão mais compatíveis
- 🔄 **Ação Requerida**: FLUSHDB no Redis, aguardar TTL expirar ou executar script de migração

**Benefícios:**
- ⚡ 36% mais compacto (19 vs 30 caracteres)
- 🚀 10x mais rápido em comparações
- 📊 Sortable naturalmente (ordem cronológica nativa)
- 🔍 Range queries extremamente eficientes

📚 Veja [docs/vhost-based-identification.md](docs/vhost-based-identification.md) para guia de migração completo.

## 🎯 Principais Funcionalidades

- **Captura Multi-Câmera**: Suporta a captura simultânea de múltiplas câmeras RTSP/IP
- **Multi-Tenant (Vhost-Based)**: Isolamento completo de dados por cliente usando RabbitMQ vhosts
- **Processamento em Edge**: Processamento local dos frames antes da transmissão
- **Distribuição via Message Broker**: Utiliza RabbitMQ com protocolo AMQP para distribuição eficiente
- **Cache Redis Otimizado**: Armazenamento de frames com TTL e formato de chave ultra-eficiente
- **Visualização em Grid**: Interface Python para visualização de todas as câmeras em uma única janela
- **Configuração Flexível**: Fácil adição/remoção de câmeras via arquivo TOML
- **Containerizado**: Deploy simplificado com Docker e Docker Compose

## 🏗️ Arquitetura

```
┌─────────────────┐
│  Câmeras RTSP   │
│  (5 câmeras)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Camera         │
│  Collector      │  ← Aplicação Go
│  (FFmpeg)       │
└────────┬────────┘
         │ JPEG Frames
         ▼
┌─────────────────┐
│   RabbitMQ      │
│   (AMQP)        │
│   Exchange:     │
│   cameras       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│                 │
│    Consumer     │  ← Visualização em Grid 2x3
│                 │
└─────────────────┘
```

## � Código Refatorado

Este repositório foi refatorado seguindo as melhores práticas de desenvolvimento Python:

### **Estrutura Refatorada:**
```
src/
├── config/
│   └── config_manager.py      # Gerenciamento de configuração
├── consumer/
│   └── rabbitmq_consumer.py   # Consumidor RabbitMQ
├── display/
│   ├── display_manager.py     # Gerenciador de display OpenCV
│   └── video_processor.py     # Processamento de frames
└── video_consumer_app.py      # Aplicação principal

tests/
├── test_config_manager.py
├── test_rabbitmq_consumer.py
├── test_display_manager.py
├── test_video_processor.py
└── test_video_consumer_app.py
```

### **Principais Melhorias:**
- **Single Responsibility Principle**: Cada classe tem uma responsabilidade específica
- **Separação de Concerns**: Lógica de negócio separada da apresentação
- **Testabilidade**: 100% de cobertura de testes unitários
- **Type Hints**: Tipagem completa para melhor manutenibilidade
- **Performance Otimizada**: Formato de chaves Redis ultra-eficiente com Unix nanoseconds

### **Otimizações de Performance (v1.2.0):**

#### 🚀 Redis Key Format Optimization
O sistema foi otimizado para usar Unix nanoseconds ao invés de RFC3339 timestamps:

**Comparação de Performance:**

| Métrica | RFC3339 | Unix Nano | Melhoria |
|---------|---------|-----------|----------|
| Tamanho da chave | 30 caracteres | 19 dígitos | **36% menor** |
| Tipo de comparação | String parsing | Integer comparison | **10x mais rápido** |
| Sortabilidade | Lexicográfica | Numérica nativa | **Natural** |
| Range queries | Parsing + comparação | `>= start AND <= end` | **Extremamente eficiente** |
| Overhead de memória (1M chaves) | ~30 MB | ~19 MB | **-11 MB** |

**Exemplo de Chave:**
```redis
# Formato Otimizado (Novo)
supermercado_vhost:frames:cam4:1731024000123456789:00001

# Formato Anterior (Deprecated)
frames:supermercado_vhost:cam4:2024-11-07T19:30:00.123456789Z:00001
```

**Benefícios Práticos:**
- ✅ Menor uso de memória Redis em deployments com milhões de chaves
- ✅ Queries temporais (range) executam 10x mais rápido
- ✅ Ordenação cronológica natural sem conversões
- ✅ Compatível com ferramentas de análise de séries temporais
- ✅ Facilita agregações e análises de dados históricos
- **Documentação**: Docstrings detalhadas seguindo padrões Python

### **Como usar o código refatorado:**
```bash
# Instalar dependências
uv sync --dev

# Executar testes
uv run pytest

# Executar aplicação refatorada
uv run python main_refactored.py

# Executar linting
uv run ruff check src/
uv run ruff format src/
```

## �🛠️ Tecnologias Utilizadas

### Backend (Collector)
- **Go 1.24**: Linguagem principal para o collector
- **FFmpeg**: Captura de frames das câmeras RTSP
- **Viper**: Gerenciamento de configuração
- **AMQP (streadway/amqp)**: Cliente RabbitMQ
- **Redis**: Cache de frames com chaves otimizadas (Unix nanoseconds)

### Message Broker & Storage
- **RabbitMQ 3.13**: Sistema de mensageria para distribuição de frames
- **Redis 7.x**: Cache de frames com TTL e multi-tenancy via vhost isolation

### Frontend (Consumer)
- **Python 3.11+**: Linguagem para o consumer
- **OpenCV**: Processamento e visualização de vídeo
- **Pika**: Cliente RabbitMQ para Python
- **NumPy**: Manipulação de arrays para concatenação de frames

### Infraestrutura
- **Docker & Docker Compose**: Containerização e orquestração
- **Alpine Linux**: Imagem base leve para containers
- **GitHub Actions**: CI/CD para testes e builds automatizados

## 📦 Estrutura do Projeto

```
edge-video/
├── config.toml              # Configuração das câmeras e parâmetros
├── docker-compose.yml       # Orquestração dos serviços
├── Dockerfile              # Build da aplicação Go
├── go.mod                  # Dependências Go
├── cmd/
│   └── edge-video/
│       └── main.go         # Entrypoint da aplicação
├── pkg/
│   ├── camera/
│   │   └── camera.go       # Lógica de captura de frames
│   ├── mq/
│   │   ├── publisher.go    # Interface do publisher
│   │   ├── amqp.go         # Implementação AMQP
│   │   └── mqtt.go         # Implementação MQTT (alternativa)
│   ├── config/
│   │   ├── config.go       # Carregamento de configuração
│   │   └── config_test.go  # Testes de configuração
│   └── util/
│       └── compress.go     # Utilitários de compressão
├── internal/
│   ├── storage/
│   │   ├── key_generator.go       # Gerador de chaves Redis otimizado
│   │   └── key_generator_test.go  # Testes do gerador (16 testes)
│   └── metadata/
│       └── publisher.go    # Publisher de metadados
├── docs/
│   ├── changelog.md                    # Changelog do projeto
│   ├── vhost-based-identification.md   # Guia de multi-tenancy
│   └── PRECOMMIT_TOWNCRIER_GUIDE.md   # Guia de contribuição
├── test_consumer.py         # Consumer Python com visualização
└── README.md               # Este arquivo
```

## 🚀 Como Executar

### Pré-requisitos

- Docker e Docker Compose instalados
- Python 3.11+ (para o consumer)
- UV (gerenciador de pacotes Python) ou pip

### 1. Configure as Câmeras

Edite o arquivo `config.toml` e adicione as URLs das suas câmeras:

```toml
[[cameras]]
id = "cam1"
url = "rtsp://user:pass@192.168.1.100:554/stream"

[[cameras]]
id = "cam2"
url = "rtsp://user:pass@192.168.1.101:554/stream"

# ... até 6 câmeras
```

### 2. Executar a Aplicação

#### Usando arquivo de configuração padrão

```bash
# Compilar e executar
go build -o edge-video ./cmd/edge-video
./edge-video

# Ou executar diretamente
go run ./cmd/edge-video
```

#### Usando arquivo de configuração customizado

```bash
# Especificar arquivo via parâmetro --config
./edge-video --config /path/to/custom-config.toml

# Ou com go run
go run ./cmd/edge-video --config config.test.toml
```

#### Validar configuração

```bash
# Validar arquivo de configuração
go run ./cmd/validate-config --config config.toml

# Ver ajuda
./edge-video --help
# Output:
#   -config string
#         Caminho para o arquivo de configuração (default "config.toml")
```

### 3. Inicie os Serviços com Docker

#### Opção A: Usando Docker Compose (Recomendado)

```bash
docker-compose up -d --build
```

Isso iniciará:
- **RabbitMQ**: Porta 5672 (AMQP) e 15672 (Management UI)
- **Camera Collector**: Aplicação Go capturando e publicando frames

#### Opção B: Usando Docker Run (Após Docker Pull)

Se você baixou a imagem do Docker Hub com `docker pull`:

```bash
# 1. Inicie o RabbitMQ primeiro
docker run -d \
  --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=user \
  -e RABBITMQ_DEFAULT_PASS=password \
  -e RABBITMQ_DEFAULT_VHOST=guard_vhost \
  rabbitmq:3.13-management-alpine

# 2. Baixe a imagem do Edge Video (se ainda não tiver)
docker pull t3labs/edge-video:latest

# 3. Execute o Camera Collector com seu config.toml local
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v /path/absoluto/para/seu/config.toml:/app/config.toml \
  t3labs/edge-video:latest
```

**Exemplos de caminhos para o volume:**

```bash
# Exemplo 1: Config.toml na pasta atual
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v $(pwd)/config.toml:/app/config.toml \
  t3labs/edge-video:latest

# Exemplo 2: Config.toml em /etc
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v /etc/edge-video/config.toml:/app/config.toml \
  t3labs/edge-video:latest

# Exemplo 3: Config.toml no home do usuário
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v $HOME/.config/edge-video/config.toml:/app/config.toml \
  t3labs/edge-video:latest

# Exemplo 4: Config.toml em storage montado
docker run -d \
  --name camera-collector \
  --link rabbitmq:rabbitmq \
  -v /mnt/storage/configs/cameras.toml:/app/config.toml \
  t3labs/edge-video:latest
```

**Usando Docker Network (Melhor prática):**

```bash
# 1. Crie uma rede
docker network create edge-video-net

# 2. Inicie o RabbitMQ na rede
docker run -d \
  --name rabbitmq \
  --network edge-video-net \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=user \
  -e RABBITMQ_DEFAULT_PASS=password \
  -e RABBITMQ_DEFAULT_VHOST=guard_vhost \
  rabbitmq:3.13-management-alpine

# 3. Execute o Camera Collector na mesma rede
docker run -d \
  --name camera-collector \
  --network edge-video-net \
  -v /path/para/seu/config.toml:/app/config.toml \
  t3labs/edge-video:latest
```

### 3. Execute o Consumer Python

```bash
# Com UV
uv run test_consumer.py

# Ou com pip
pip install -r requirements.txt
python test_consumer.py
```

### 4. Visualize as Câmeras

Uma janela será aberta mostrando todas as câmeras em uma grade 2x3.

**Pressione 'q' para sair.**

## ⚙️ Configuração

### config.toml

```toml
interval_ms = 500                    # Intervalo entre capturas (ms)
protocol = "amqp"                    # Protocolo: amqp ou mqtt
process_every_n_frames = 3           # Reduz taxa de frames (1 a cada 3)

[amqp]
amqp_url = "amqp://user:password@rabbitmq:5672/guard_vhost"
exchange = "cameras"
routing_key_prefix = "camera"

[compression]
enabled = false                      # Compressão zstd (desabilitada)
level = 3

[[cameras]]
id = "cam1"
url = "rtsp://..."

[[cameras]]
id = "cam2"
url = "rtsp://..."
```

### 🔄 Optional Redis Frame Storage + Metadata

You can enable Redis frame caching and metadata publishing by updating `config.toml`:

```toml
[redis]
enabled = true
address = "redis:6379"
ttl_seconds = 300
prefix = "frames"

[metadata]
enabled = true
exchange = "camera.metadata"
routing_key = "camera.metadata.event"
```

When enabled:

- Frames are stored in Redis with TTL
- Metadata messages are sent asynchronously to RabbitMQ
- Existing video streaming and publishing are unaffected

### 🏢 Isolamento Multi-Cliente (Multi-tenancy)

O Edge Video usa o **vhost do RabbitMQ** como identificador único de cliente, garantindo isolamento automático de dados no Redis.

#### Formato de Chave Redis

```
{vhost}:{prefix}:{cameraID}:{unix_timestamp_nano}:{sequence}
```

**Exemplo:**
```redis
supermercado_vhost:frames:cam4:1731024000123456789:00001
```

#### Como Funciona

1. **Vhost Extraído Automaticamente**: O vhost é extraído da URL AMQP configurada
2. **Unix Nanoseconds**: Timestamps numéricos para sortabilidade e performance
3. **Chaves Redis Isoladas**: Cada cliente possui namespace próprio no Redis
4. **Zero Configuração Adicional**: Não é necessário configurar `instance_id` separadamente

#### Exemplo: Múltiplos Clientes

```toml
# Cliente A (config-client-a.toml)
[amqp]
amqp_url = "amqp://user:pass@rabbitmq:5672/client-a"

# Cliente B (config-client-b.toml) 
[amqp]
amqp_url = "amqp://user:pass@rabbitmq:5672/client-b"
```

**Resultado no Redis:**
```redis
client-a:frames:cam1:1731024000123456789:00001
client-b:frames:cam1:1731024000123456789:00001
```

#### Por que Unix Timestamp?

| Aspecto | RFC3339 | Unix Nano | Vantagem |
|---------|---------|-----------|----------|
| **Tamanho** | 30 chars | 19 dígitos | ✅ 36% menor |
| **Sortable** | String | Numérico | ✅ Natural |
| **Comparação** | Parsing | Inteiro | ✅ 10x mais rápido |
| **Range Query** | Complexo | Simples | ✅ `>= start AND <= end` |

**Benefícios:**
- ✅ Impossível colisão entre clientes diferentes
- ✅ Mesmas câmeras em clientes diferentes não conflitam
- ✅ Timestamps compactos e sortable numericamente
- ✅ Range queries extremamente eficientes
- ✅ Alinhamento com arquitetura AMQP (vhost = multi-tenancy)

📚 **Documentação Completa**: Veja [docs/vhost-based-identification.md](docs/vhost-based-identification.md) para detalhes de implementação, exemplos de deployment e troubleshooting.

## 🔍 Monitoramento

### RabbitMQ Management UI

Acesse: `http://localhost:15672`
- **Usuário**: user
- **Senha**: password

### Logs do Collector

```bash
docker logs camera-collector -f
```

### Métricas do Sistema

Verifique o throughput de mensagens e o uso de recursos no RabbitMQ Management.

## 📊 Casos de Uso

1. **Vigilância e Segurança**: Monitoramento em tempo real de múltiplas câmeras
2. **Análise de Vídeo**: Processamento de frames para detecção de objetos, pessoas, etc.
3. **Edge Computing**: Processamento local antes de envio para a nuvem
4. **Sistemas de Visão Computacional**: Pipeline para aplicações de Computer Vision
5. **Armazenamento Inteligente**: Gravação seletiva baseada em eventos

## 🔧 Desenvolvimento

### Adicionar Nova Câmera

1. Edite `config.toml`
2. Adicione a nova entrada em `[[cameras]]`
3. Reinicie o container: `docker-compose restart camera-collector`

### Modificar Taxa de Frames

Ajuste `interval_ms` no `config.toml` para controlar a taxa de captura.

### Habilitar Compressão

```toml
[compression]
enabled = true
level = 3  # 1-22 (maior = mais compressão)
```

### Habilitar Redis e Metadata

```toml
[redis]
enabled = true
address = "redis:6379"
password = ""  # Opcional
ttl_seconds = 300
prefix = "frames"

[metadata]
enabled = true
exchange = "camera.metadata"
routing_key = "camera.metadata.event"
```

## Uso no Windows (Executável)

**Instalação:**
- Baixe o instalador `EdgeVideoSetup-X.X.X.exe` no [GitHub Releases](https://github.com/T3-Labs/edge-video/releases).
- Execute como Administrador e siga o assistente de instalação.
- O serviço será instalado e iniciado automaticamente.

**Configuração:**
- Edite as câmeras e parâmetros em `C:\Program Files\T3Labs\EdgeVideo\config\config.toml`.

**Gerenciamento do Serviço:**
- Pelo Services.msc (Interface Gráfica):
  - Win + R → services.msc → "Edge Video Camera Capture Service"
- Pela linha de comando:
  ```cmd
  # Instalar serviço manualmente
  edge-video-service.exe install

  # Iniciar serviço
  net start EdgeVideoService
  # ou
  edge-video-service.exe start

  # Parar serviço
  net stop EdgeVideoService
  # ou
  edge-video-service.exe stop

  # Desinstalar serviço
  edge-video-service.exe uninstall
  ```
- Para troubleshooting, rode em modo console:
  ```cmd
  edge-video-service.exe console
  ```
- Logs podem ser visualizados em `C:\Program Files\T3Labs\EdgeVideo\logs\` ou pelo Event Viewer (Application → EdgeVideoService).

---

**Componentes:**
- `supermercado_vhost` - Identificador do cliente (extraído do AMQP vhost)
- `frames` - Prefixo configurável
- `cam4` - ID da câmera
- `1731024000123456789` - Unix timestamp em nanosegundos
- `00001` - Sequência anti-colisão

#### Como Funciona

1. **Vhost Extraído Automaticamente**: O vhost é extraído da URL AMQP configurada
2. **Unix Nanoseconds**: Timestamps numéricos para sortabilidade e performance
3. **Chaves Redis Isoladas**: Cada cliente possui namespace próprio no Redis
4. **Zero Configuração Adicional**: Não é necessário configurar `instance_id` separadamente

#### Exemplo: Múltiplos Clientes

```toml
# Cliente A (config-client-a.toml)
[amqp]
amqp_url = "amqp://user:pass@rabbitmq:5672/client-a"

# Cliente B (config-client-b.toml) 
[amqp]
amqp_url = "amqp://user:pass@rabbitmq:5672/client-b"
```

**Resultado no Redis:**
```redis
client-a:frames:cam1:1731024000123456789:00001
client-b:frames:cam1:1731024000123456789:00001
```

#### Por que Unix Timestamp?

| Aspecto | RFC3339 | Unix Nano | Vantagem |
|---------|---------|-----------|----------|
| **Tamanho** | 30 chars | 19 dígitos | ✅ 36% menor |
| **Sortable** | String | Numérico | ✅ Natural |
| **Comparação** | Parsing | Inteiro | ✅ 10x mais rápido |
| **Range Query** | Complexo | Simples | ✅ `>= start AND <= end` |

**Benefícios:**
- ✅ Impossível colisão entre clientes diferentes
- ✅ Mesmas câmeras em clientes diferentes não conflitam
- ✅ Timestamps compactos e sortable numericamente
- ✅ Range queries extremamente eficientes
- ✅ Alinhamento com arquitetura AMQP (vhost = multi-tenancy)

📚 **Documentação Completa**: Veja [docs/vhost-based-identification.md](docs/vhost-based-identification.md) para detalhes de implementação, exemplos de deployment e troubleshooting.

## 🔍 Monitoramento

### RabbitMQ Management UI

Acesse: `http://localhost:15672`
- **Usuário**: user
- **Senha**: password

### Logs do Collector

```bash
docker logs camera-collector -f
```

### Métricas do Sistema

Verifique o throughput de mensagens e o uso de recursos no RabbitMQ Management.

## 📊 Casos de Uso

1. **Vigilância e Segurança**: Monitoramento em tempo real de múltiplas câmeras
2. **Análise de Vídeo**: Processamento de frames para detecção de objetos, pessoas, etc.
3. **Edge Computing**: Processamento local antes de envio para a nuvem
4. **Sistemas de Visão Computacional**: Pipeline para aplicações de Computer Vision
5. **Armazenamento Inteligente**: Gravação seletiva baseada em eventos

## 🔧 Desenvolvimento

### Adicionar Nova Câmera

1. Edite `config.toml`
2. Adicione a nova entrada em `[[cameras]]`
3. Reinicie o container: `docker-compose restart camera-collector`

### Modificar Taxa de Frames

Ajuste `interval_ms` no `config.toml` para controlar a taxa de captura.

### Habilitar Compressão

```toml
[compression]
enabled = true
level = 3  # 1-22 (maior = mais compressão)
```

### Habilitar Redis e Metadata

```toml
[redis]
enabled = true
address = "redis:6379"
password = ""  # Opcional
ttl_seconds = 300
prefix = "frames"

[metadata]
enabled = true
exchange = "camera.metadata"
routing_key = "camera.metadata.event"
```

## Uso no Windows (Executável)

**Instalação:**
- Baixe o instalador `EdgeVideoSetup-X.X.X.exe` no [GitHub Releases](https://github.com/T3-Labs/edge-video/releases).
- Execute como Administrador e siga o assistente de instalação.
- O serviço será instalado e iniciado automaticamente.

**Configuração:**
- Edite as câmeras e parâmetros em `C:\Program Files\T3Labs\EdgeVideo\config\config.toml`.

**Gerenciamento do Serviço:**
- Pelo Services.msc (Interface Gráfica):
  - Win + R → services.msc → "Edge Video Camera Capture Service"
- Pela linha de comando:
  ```cmd
  # Instalar serviço manualmente
  edge-video-service.exe install

  # Iniciar serviço
  net start EdgeVideoService
  # ou
  edge-video-service.exe start

  # Parar serviço
  net stop EdgeVideoService
  # ou
  edge-video-service.exe stop

  # Desinstalar serviço
  edge-video-service.exe uninstall
  ```
- Para troubleshooting, rode em modo console:
  ```cmd
  edge-video-service.exe console
  ```
- Logs podem ser visualizados em `C:\Program Files\T3Labs\EdgeVideo\logs\` ou pelo Event Viewer (Application → EdgeVideoService).
