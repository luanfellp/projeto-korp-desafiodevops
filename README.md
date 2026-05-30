# http-server-projeto-korp

Serviço HTTP em Go com observabilidade via Prometheus/Grafana e automação completa com Ansible.

---

## Estrutura do Projeto

```
projeto-korp/
├── app/
│   ├── main.go                  # Servidor HTTP Go
│   ├── go.mod                   # Módulo Go + dependências
│   └── Dockerfile               # Build multi-stage
├── nginx/
│   └── conf.d/
│       └── http-server-projeto-korp.conf   # Proxy reverso
├── prometheus/
│   └── prometheus.yml           # Configuração de scraping
├── grafana/
│   └── provisioning/
│       ├── datasources/
│       │   └── datasources.yml  # Datasource Prometheus
│       └── dashboards/
│           ├── dashboards.yml   # Provider de dashboards
│           └── http-server-projeto-korp-dashboard.json
├── ansible/
│   ├── inventory.ini            # Inventário (localhost por padrão)
│   └── playbook.yml             # Playbook completo
├── docker-compose.yml
└── README.md
```

---

## Parte 1 — Serviço e Infraestrutura

### Rodar manualmente (sem Ansible)

```bash
# Build e start de todos os containers
docker compose up --build -d

# Testar o endpoint
curl http://localhost:80/projeto-korp
```

**Resposta esperada:**
```json
{"nome":"Projeto Korp","horario":"2024-05-01T12:00:00Z"}
```

### Arquitetura de rede

- A aplicação Go **não expõe portas ao host** — acessível apenas via NGINX.
- Todos os containers compartilham a rede bridge `projeto-korp-net`.
- NGINX faz proxy reverso de `:80` → `http-server-projeto-korp:8080`.

---

## Parte 2 — Monitoramento e Observabilidade

| Serviço    | URL                                        |
|------------|--------------------------------------------|
| NGINX      | http://localhost:80/projeto-korp           |
| Prometheus | http://localhost:9090                      |
| Grafana    | http://localhost:3000  (admin / admin)     |

### Métricas expostas (`/metrics`)

| Métrica               | Tipo    | Descrição                                    |
|-----------------------|---------|----------------------------------------------|
| `service_up`          | Gauge   | 1 = serviço disponível; 0/ausente = indisponível |
| `http_requests_total` | Counter | Volume de requisições (labels: method, path, status_code) |

### Dashboard Grafana

O dashboard `http-server-projeto-korp` é **provisionado automaticamente** e contém:
- Painel de disponibilidade (UP / DOWN)
- Total acumulado de requisições
- Taxa de requisições em req/s (janela de 1 min)
- Histórico de disponibilidade
- Volume de requisições por minuto (barras empilhadas)

---

## Parte 3 — Automação com Ansible

### Pré-requisitos

- Ansible instalado na máquina de controle
- Sistema alvo: Ubuntu 20.04+ / Debian 11+
- Acesso `sudo` no host alvo

```bash
pip install ansible
```

### Execução (único comando)

```bash
cd ansible/
ansible-playbook -i inventory.ini playbook.yml
```

**Para host remoto**, edite `inventory.ini`:
```ini
[projeto_korp]
192.168.1.100 ansible_user=ubuntu ansible_ssh_private_key_file=~/.ssh/id_rsa
```

### O que o playbook faz

1. Remove versões antigas do Docker
2. Instala Docker CE + Compose Plugin via repositório oficial
3. Inicia e habilita o serviço Docker
4. Copia os arquivos do projeto para `/opt/projeto-korp`
5. Cria a rede Docker `projeto-korp-net`
6. Faz o build da imagem `http-server-projeto-korp`
7. Sobe todos os containers (`docker compose up -d`)
8. Aguarda o serviço responder em `http://localhost:80/projeto-korp`
9. Exibe a resposta JSON no console e imprime o resumo do ambiente

---

## Decisões Técnicas

- **Multi-stage build**: imagem final baseada em `alpine:3.19` (~15 MB), sem toolchain Go.
- **Métricas de disponibilidade**: `service_up` como Gauge — enquanto o processo está vivo vale `1`; se o container cair o Prometheus registra ausência da métrica, equivalente a `0`.
- **Provisionamento Grafana**: datasource e dashboard configurados via arquivos `provisioning/`, sem necessidade de configuração manual.
- **Idempotência Ansible**: tasks com `creates:`, `ignore_errors` e checagem prévia garantem que o playbook pode ser reexecutado sem efeitos colaterais.
