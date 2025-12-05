# Guia de Refatoração do Repositório Edge Video

## 🎯 Objetivo
Organizar o repositório para facilitar navegação, manutenção e onboarding de novos desenvolvedores.

---

## 📁 Estrutura Proposta

```
edge-video/
├── .github/                     # GitHub Actions workflows
├── cmd/                         # Aplicações Go
│   ├── edge-video/             # Aplicação principal
│   └── edge-video-service/     # Windows service wrapper
├── pkg/                         # Pacotes públicos reutilizáveis
│   ├── buffer/
│   ├── camera/
│   ├── circuit/
│   ├── config/
│   ├── logger/
│   ├── memcontrol/             # ✨ Novo: Controle de memória
│   ├── metrics/
│   ├── mq/
│   ├── registration/
│   ├── util/
│   └── worker/
├── internal/                    # Pacotes privados internos
│   ├── metadata/
│   └── storage/
├── configs/                     # ✨ Novo: Consolidar todas as configs
│   ├── config.example.toml
│   ├── config.memory-control.toml
│   ├── config.test.toml
│   └── docker-compose/
│       ├── docker-compose.yml
│       └── docker-compose.test.yml
├── examples/                    # ✨ Novo: Exemplos de uso
│   ├── python/
│   │   ├── consumer_basic.py
│   │   ├── consumer_with_opencv.py
│   │   └── consumer_status_monitor.py
│   └── go/
│       └── validate-config/
├── scripts/                     # Scripts de build e deploy
│   ├── build-windows.sh
│   ├── run-docker.sh
│   ├── build-changelog.sh
│   └── new-changelog.sh
├── installer/                   # Instalador Windows
│   └── windows/
├── docs/                        # Documentação MkDocs
│   ├── getting-started/
│   ├── features/
│   ├── guides/
│   ├── architecture/
│   ├── api/
│   ├── windows/
│   └── development/
├── .github/                     # CI/CD workflows
├── tests/                       # ✨ Novo: Testes integrados (opcional)
├── Dockerfile
├── README.md
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md
├── go.mod
├── go.sum
├── mkdocs.yml
└── .gitignore
```

---

## 🧹 Passo 1: Remover Arquivos Temporários

Execute os seguintes comandos:

```bash
# Remover binários compilados (já no .gitignore)
rm -f edge-video edge-video-test

# Remover coverage reports
rm -f coverage.out

# Remover arquivo XML de saída
rm -f repomix-output.xml

# Limpar diretórios de build
rm -rf site/
rm -rf dist/ (se não estiver sendo usado para releases)
```

---

## 📦 Passo 2: Reorganizar Configurações

```bash
# Criar diretório de configs
mkdir -p configs/docker-compose

# Mover arquivos de configuração
mv config.toml configs/config.example.toml
mv config-with-memory-control.toml configs/config.memory-control.toml
mv config.test.toml configs/config.test.toml

# Mover docker-compose
mv docker-compose.yml configs/docker-compose/
mv docker-compose.test.yml configs/docker-compose/

# Criar symlinks na raiz para compatibilidade (opcional)
ln -s configs/config.example.toml config.toml
ln -s configs/docker-compose/docker-compose.yml docker-compose.yml
```

---

## 🐍 Passo 3: Reorganizar Exemplos Python

```bash
# Criar diretório de exemplos
mkdir -p examples/python

# Mover consumers
mv test_camera_redis_amqp.py examples/python/consumer_basic.py
mv test_consumer_status.py examples/python/consumer_status_monitor.py

# Se test_consumer.py existir e for útil
# mv test_consumer.py examples/python/consumer_legacy.py

# Mover validate-config para examples/go
mkdir -p examples/go
mv cmd/validate-config examples/go/
```

---

## 📜 Passo 4: Organizar Scripts

```bash
# Todos os scripts já estão em scripts/ - verificar se há duplicatas
cd scripts/

# Garantir que todos os scripts sejam executáveis
chmod +x *.sh
```

---

## 📚 Passo 5: Consolidar Documentação

```bash
cd docs/

# Mover documentos da raiz para docs/guides/
mkdir -p guides

# Mover apenas se não houver conflito
# VHOST_IMPLEMENTATION.md -> docs/guides/vhost-implementation.md
# IMPLEMENTATION-SUMMARY.md -> docs/guides/implementation-summary.md

# Documentos que devem permanecer na raiz:
# - README.md (entrada principal)
# - CHANGELOG.md (histórico de versões)
# - CONTRIBUTING.md (guia de contribuição)
# - LICENSE (licença do projeto)
```

---

## 🔧 Passo 6: Atualizar .gitignore

Adicione ao final do `.gitignore`:

```gitignore
# Built binaries
edge-video
edge-video-test
edge-video-service
edge-video-service.exe

# Temporary files
*.tmp
*.log
*.swp
*.swo
*~
.DS_Store

# Build artifacts
dist/
build/
bin/

# XML output files
repomix-output.xml
*.xml.bak

# IDE specific (simplificado)
.vscode/
.idea/

# Test artifacts
*.test
coverage.*

# OS specific
Thumbs.db
```

---

## 📝 Passo 7: Atualizar Referências

Após mover arquivos, atualize as referências nos seguintes locais:

### README.md
```markdown
# Atualizar seção de configuração
- config.toml -> configs/config.example.toml

# Atualizar seção de Docker
- docker-compose.yml -> configs/docker-compose/docker-compose.yml

# Atualizar exemplos Python
- test_camera_redis_amqp.py -> examples/python/consumer_basic.py
```

### Dockerfile
```dockerfile
# Se houver referência a config.toml, atualizar para:
COPY configs/config.example.toml /app/config.toml
```

### docker-compose.yml (novo caminho)
```yaml
# Verificar volumes e paths
volumes:
  - ./configs/config.example.toml:/app/config.toml
```

### GitHub Actions (`.github/workflows/`)
```yaml
# Atualizar paths de configs se necessário
- run: go test -v ./...
  # Garantir que tests encontrem configs em configs/
```

### mkdocs.yml
```yaml
# Verificar se paths de documentos estão corretos
nav:
  - Home: index.md
  - Guides:
    - Implementation Summary: guides/implementation-summary.md
    - Vhost Implementation: guides/vhost-implementation.md
```

---

## 🧪 Passo 8: Validar Estrutura

Execute os seguintes testes:

```bash
# 1. Compilar o projeto
go build -o edge-video ./cmd/edge-video

# 2. Executar testes
go test ./...

# 3. Validar config de exemplo
./edge-video --config configs/config.example.toml --validate

# 4. Verificar docker-compose
cd configs/docker-compose
docker-compose config

# 5. Verificar documentação
mkdocs serve
# Acesse http://localhost:8000 e verifique links
```

---

## 📋 Passo 9: Atualizar Documentação de Onboarding

Crie ou atualize `docs/getting-started/quick-start.md`:

```markdown
# Quick Start

## Estrutura do Projeto

- `cmd/` - Aplicações executáveis
- `pkg/` - Pacotes reutilizáveis
- `internal/` - Código interno privado
- `configs/` - Arquivos de configuração
- `examples/` - Exemplos de uso
- `docs/` - Documentação completa
- `scripts/` - Scripts de build/deploy
- `installer/` - Instalador Windows

## Configuração Rápida

1. Copie a configuração de exemplo:
   ```bash
   cp configs/config.example.toml config.toml
   ```

2. Edite `config.toml` com suas câmeras

3. Execute:
   ```bash
   go build -o edge-video ./cmd/edge-video
   ./edge-video --config config.toml
   ```

## Docker Compose

```bash
cd configs/docker-compose
docker-compose up -d
```

## Exemplos Python

```bash
cd examples/python
python consumer_basic.py
```
```

---

## 🎨 Passo 10: Melhorias no README.md

Simplifique o README principal para ser uma **página de entrada**:

```markdown
# Edge Video

> Sistema distribuído de captura e processamento de vídeo para edge computing

[![Go Tests](badge)](link)
[![License](badge)](link)

## ✨ Features

- Multi-câmera RTSP/IP
- Isolamento multi-tenant (RabbitMQ vhost)
- Controle de memória (previne travamento do SO)
- Armazenamento Redis com TTL
- Distribuição via AMQP/MQTT
- Instalador Windows como serviço
- Consumer Python com OpenCV

## 🚀 Quick Start

### Local
```bash
cp configs/config.example.toml config.toml
# Edite config.toml
go build -o edge-video ./cmd/edge-video
./edge-video --config config.toml
```

### Docker
```bash
cd configs/docker-compose
docker-compose up -d
```

### Windows Installer
Baixe no [GitHub Releases](link)

## 📚 Documentação

- [Documentação Completa](https://t3-labs.github.io/edge-video/)
- [Getting Started](docs/getting-started/)
- [Controle de Memória](docs/MEMORY-CONTROL.md)
- [Guias](docs/guides/)
- [API Reference](docs/api/)

## 🛠️ Desenvolvimento

```bash
# Testes
go test ./...

# Lint
golangci-lint run

# Build para Windows
./scripts/build-windows.sh
```

Ver [CONTRIBUTING.md](CONTRIBUTING.md)

## 📄 Licença

MIT License - ver [LICENSE](LICENSE)
```

---

## ✅ Checklist Final

Após completar a refatoração, verifique:

- [ ] Todos os testes passam: `go test ./...`
- [ ] Projeto compila: `go build ./cmd/edge-video`
- [ ] Docker compose funciona: `docker-compose up`
- [ ] Documentação renderiza: `mkdocs serve`
- [ ] Links no README estão corretos
- [ ] Exemplos Python funcionam
- [ ] .gitignore atualizado
- [ ] Nenhum arquivo temporário commitado
- [ ] README simplificado e claro
- [ ] Estrutura de diretórios lógica

---

## 📊 Antes vs Depois

### ❌ Antes (Raiz Desorganizada)
```
edge-video/
├── config.toml
├── config.test.toml
├── config-with-memory-control.toml
├── docker-compose.yml
├── docker-compose.test.yml
├── test_camera_redis_amqp.py
├── test_consumer.py
├── test_consumer_status.py
├── VHOST_IMPLEMENTATION.md
├── IMPLEMENTATION-SUMMARY.md
├── edge-video (binário)
├── edge-video-test (binário)
├── coverage.out
├── repomix-output.xml
└── 40+ arquivos na raiz
```

### ✅ Depois (Raiz Limpa)
```
edge-video/
├── cmd/
├── pkg/
├── internal/
├── configs/
├── examples/
├── docs/
├── scripts/
├── installer/
├── README.md
├── CHANGELOG.md
├── CONTRIBUTING.md
├── LICENSE
├── Dockerfile
├── go.mod
├── go.sum
├── mkdocs.yml
└── .gitignore

Total: ~15 itens na raiz (todos essenciais)
```

---

## 🎯 Benefícios da Refatoração

1. **Navegação Intuitiva**
   - Arquivos agrupados por função
   - Estrutura clara e previsível

2. **Manutenção Simplificada**
   - Fácil localizar configs
   - Exemplos separados do código principal

3. **Onboarding Rápido**
   - Novo desenvolvedor encontra tudo facilmente
   - README direto ao ponto

4. **Build & Deploy**
   - Scripts organizados
   - Configs isoladas

5. **Documentação**
   - Estrutura lógica em docs/
   - Links corretos e mantíveis

---

## 🚀 Execução do Plano

Execute os comandos nesta ordem:

```bash
# 1. Backup (segurança)
git stash
git checkout -b refactor/organize-repo

# 2. Criar nova estrutura
mkdir -p configs/docker-compose examples/python examples/go

# 3. Mover arquivos (usar os comandos dos passos acima)

# 4. Atualizar referências

# 5. Testar
go test ./...
go build ./cmd/edge-video

# 6. Commit
git add -A
git commit -m "refactor: Reorganize repository structure

- Move configs to configs/ directory
- Move Python examples to examples/python/
- Move Go examples to examples/go/
- Update .gitignore for temporary files
- Simplify root directory
- Update documentation references

Improves:
- Navigation and discoverability
- Onboarding experience
- Maintainability
- Build and deployment workflows"

# 7. Merge (após review)
git checkout main
git merge refactor/organize-repo
```

---

## 📞 Dúvidas?

Consulte:
- `docs/development/contributing.md` para guias de contribuição
- `CONTRIBUTING.md` para processo de desenvolvimento
- GitHub Issues para reportar problemas

---

**Data de Criação**: 2024-11-25
**Versão**: 1.0.0
**Status**: 📋 Pronto para execução
