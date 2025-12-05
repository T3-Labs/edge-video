# ✅ Checklist de Refatoração - Edge Video

Use este checklist para acompanhar o progresso da refatoração.

---

## 📋 Preparação

- [ ] Fazer backup do branch atual
- [ ] Criar branch de refatoração: `git checkout -b refactor/organize-repo`
- [ ] Garantir que todos os testes passam antes de começar
- [ ] Fazer stash de mudanças não commitadas

---

## 🗂️ Reorganização de Estrutura

### Criar Diretórios
- [ ] `mkdir -p configs/docker-compose`
- [ ] `mkdir -p examples/python`
- [ ] `mkdir -p examples/go`
- [ ] `mkdir -p tests` (opcional)

### Mover Configurações
- [ ] `config.toml` → `configs/config.example.toml`
- [ ] `config-with-memory-control.toml` → `configs/config.memory-control.toml`
- [ ] `config.test.toml` → `configs/config.test.toml`
- [ ] `docker-compose.yml` → `configs/docker-compose/`
- [ ] `docker-compose.test.yml` → `configs/docker-compose/`

### Mover Exemplos Python
- [ ] `test_camera_redis_amqp.py` → `examples/python/consumer_basic.py`
- [ ] `test_consumer_status.py` → `examples/python/consumer_status_monitor.py`
- [ ] `test_consumer.py` → `examples/python/consumer_legacy.py` (se existir)

### Mover Exemplos Go
- [ ] `cmd/validate-config` → `examples/go/validate-config`

### Mover Documentação
- [ ] Verificar se docs já estão organizados em `docs/`
- [ ] `VHOST_IMPLEMENTATION.md` → manter na raiz OU mover para `docs/guides/`
- [ ] `IMPLEMENTATION-SUMMARY.md` → manter na raiz OU mover para `docs/guides/`

---

## 🧹 Limpeza

### Remover Arquivos Temporários
- [ ] `rm -f edge-video` (binário compilado)
- [ ] `rm -f edge-video-test` (binário de teste)
- [ ] `rm -f coverage.out` (coverage report)
- [ ] `rm -f repomix-output.xml` (output XML)
- [ ] `rm -f *.log` (arquivos de log)

### Remover Diretórios de Build
- [ ] `rm -rf site/` (MkDocs output)
- [ ] `rm -rf dist/` (se não for usado para releases)
- [ ] Verificar e limpar `.venv/`, `.venv-docs/`, `.venv-tools/`

---

## 🔗 Criar Symlinks

Para manter compatibilidade com scripts e workflows existentes:

- [ ] `ln -sf configs/config.example.toml config.toml`
- [ ] `ln -sf configs/docker-compose/docker-compose.yml docker-compose.yml`

---

## 📝 Atualizar Arquivos

### .gitignore
- [ ] Adicionar `edge-video` e `edge-video-test`
- [ ] Adicionar `*.log`, `*.tmp`, `*.swp`
- [ ] Adicionar `repomix-output.xml`
- [ ] Adicionar `coverage.*`
- [ ] Simplificar seção de IDE (apenas `.vscode/` e `.idea/`)

### README.md
- [ ] Atualizar seção "Quick Start" com novos paths
- [ ] Atualizar exemplos de configuração
- [ ] Atualizar estrutura do projeto
- [ ] Atualizar links para documentação
- [ ] Simplificar e tornar mais direto (usar README.NEW.md como base)

### Dockerfile
- [ ] Verificar paths de COPY (se houver)
- [ ] Atualizar referências a `config.toml`
- [ ] Testar build: `docker build -t edge-video:test .`

### docker-compose.yml (novo path)
- [ ] Atualizar volumes para apontar para `configs/`
- [ ] Verificar paths relativos
- [ ] Testar: `cd configs/docker-compose && docker-compose config`

### GitHub Actions (`.github/workflows/`)
- [ ] Verificar paths de configurações de teste
- [ ] Atualizar referências a `config.test.toml` → `configs/config.test.toml`
- [ ] Atualizar paths de docker-compose se necessário

### mkdocs.yml
- [ ] Verificar navegação de documentos
- [ ] Atualizar paths se documentos foram movidos
- [ ] Testar: `mkdocs serve`

---

## 📚 Criar Documentação

### configs/README.md
- [ ] Criar documentação de configurações
- [ ] Explicar cada arquivo de config
- [ ] Adicionar exemplos de uso
- [ ] Linkar para documentação principal

### examples/README.md
- [ ] Documentar exemplos Python
- [ ] Documentar exemplos Go
- [ ] Adicionar instruções de uso
- [ ] Listar requisitos (pip, go modules)

### REFACTORING_SUMMARY.md
- [ ] Criar sumário de mudanças
- [ ] Listar estrutura antes/depois
- [ ] Documentar file movements
- [ ] Listar benefícios

---

## ✅ Validação

### Compilação e Testes
- [ ] `go build -o edge-video ./cmd/edge-video` (deve compilar)
- [ ] `go test ./...` (todos os testes devem passar)
- [ ] `go test -race ./...` (sem race conditions)
- [ ] `golangci-lint run` (sem erros de lint)

### Docker
- [ ] `docker build -t edge-video:test .` (build com sucesso)
- [ ] `cd configs/docker-compose && docker-compose up` (sobe sem erros)
- [ ] `docker-compose logs` (verificar logs)

### Documentação
- [ ] `mkdocs serve` (documentação carrega sem erros)
- [ ] Verificar links quebrados na documentação
- [ ] Verificar imagens e assets carregam

### Exemplos
- [ ] Testar exemplo Python básico
- [ ] Testar consumer de status
- [ ] Testar validate-config Go

### Estrutura
- [ ] Raiz do repositório tem ~15 itens ou menos
- [ ] Todos os arquivos essenciais estão na raiz
- [ ] Todos os arquivos secundários estão organizados em subdiretórios

---

## 📦 Git Operations

### Stage Changes
- [ ] `git add -A` (adicionar todas as mudanças)
- [ ] `git status` (revisar mudanças)

### Commit
- [ ] Criar commit com mensagem descritiva:
```bash
git commit -m "refactor: Reorganize repository structure

- Move configs to configs/ directory
- Move Python examples to examples/python/
- Move Go examples to examples/go/
- Update .gitignore for temporary files
- Simplify root directory (40+ → ~15 items)
- Create documentation for configs and examples
- Update references in README and workflows

Improves:
- Navigation and discoverability
- Onboarding experience
- Maintainability
- Build and deployment workflows

BREAKING CHANGE: File paths have changed. Update your
workflows and scripts to use new paths."
```

### Review
- [ ] `git diff HEAD~1` (revisar mudanças)
- [ ] `git log --oneline -5` (verificar commits)

### Merge
- [ ] Testar branch de refatoração completamente
- [ ] `git checkout main`
- [ ] `git merge refactor/organize-repo`
- [ ] Resolver conflitos se houver
- [ ] Push: `git push origin main`

---

## 🚀 Deploy e Comunicação

### Atualizar CI/CD
- [ ] Verificar se workflows GitHub Actions funcionam
- [ ] Atualizar badges no README se necessário
- [ ] Testar pipeline completo

### Comunicar Mudanças
- [ ] Criar release notes com breaking changes
- [ ] Atualizar documentação de deployment
- [ ] Notificar time sobre mudanças de paths

### Atualizar Ambientes
- [ ] Atualizar servidores de produção
- [ ] Atualizar containers Docker
- [ ] Atualizar instalações Windows

---

## 📊 Métricas de Sucesso

### Antes da Refatoração
- Arquivos na raiz: ~40+
- Tempo para encontrar config: ~2-3 minutos
- Tempo para encontrar exemplos: ~3-5 minutos
- Clareza da estrutura: 5/10

### Depois da Refatoração
- [ ] Arquivos na raiz: ~15 ou menos ✓
- [ ] Tempo para encontrar config: ~30 segundos ✓
- [ ] Tempo para encontrar exemplos: ~1 minuto ✓
- [ ] Clareza da estrutura: 9/10 ✓

---

## 🎯 Benefícios Alcançados

- [ ] ✅ Raiz do repositório limpa e organizada
- [ ] ✅ Configurações centralizadas em `configs/`
- [ ] ✅ Exemplos fáceis de encontrar em `examples/`
- [ ] ✅ Documentação clara de estrutura
- [ ] ✅ Onboarding mais rápido para novos devs
- [ ] ✅ Manutenção facilitada
- [ ] ✅ Build e deploy simplificados
- [ ] ✅ Navegação intuitiva
- [ ] ✅ Compatibilidade mantida via symlinks

---

## 📞 Troubleshooting

### Problema: Testes falhando após mudança
**Solução**: Verificar paths de configurações de teste em `*_test.go`

### Problema: Docker Compose não encontra configs
**Solução**: Atualizar volumes no `docker-compose.yml` para apontar para `configs/`

### Problema: CI/CD falhando
**Solução**: Atualizar paths nos workflows `.github/workflows/`

### Problema: Symlinks não funcionam no Windows
**Solução**: Executar como Admin ou usar cópias ao invés de symlinks

---

## ✨ Status Final

- [ ] ✅ Refatoração completa
- [ ] ✅ Todos os testes passando
- [ ] ✅ Documentação atualizada
- [ ] ✅ Commit criado
- [ ] ✅ Merge realizado
- [ ] ✅ Deploy atualizado

---

**Data de Conclusão**: _______________

**Responsável**: _______________

**Aprovado por**: _______________

---

<p align="center">
  🎉 <b>Parabéns pela refatoração!</b> 🎉
</p>
