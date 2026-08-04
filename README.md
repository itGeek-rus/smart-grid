# Smart Grid

Сервисы алгоритмов умных сетей: приём телеметрии с IoT, обработка событий, API и наблюдаемость.

## Архитектура

```
Devices → MQTT → Ingestion → Kafka (raw.telemetry)
                              ↓
                         Processor → TimescaleDB / Redis
                              ↓
                    processed.events / alerts / DLQ
                              ↓
                             API (REST)
                              ↓
                    Prometheus ← /metrics → Grafana
```

| Сервис | Порт | Назначение |
|--------|------|------------|
| **ingestion** | `:8081` | MQTT → валидация → Kafka `raw.telemetry` / DLQ |
| **processor** | `:8080` | Kafka → алгоритмы → Timescale/Redis → `processed.events` / `alerts` |
| **api** | `:8082` | REST чтение устройств/телеметрии/алертов; команды → Kafka `commands` |

Алгоритмы processor (MVP): `voltage_out_of_range` (200–250 В), `power_spike` (>1.8× среднего окна).

## Стек

Go, MQTT (Mosquitto), Kafka, TimescaleDB, Redis, Prometheus, Grafana, Docker Compose. Далее: K8s/Helm, OTel.

## Структура

```
cmd/ingestion|processor|api
internal/domain|usecase|service|repository|transport
internal/pkg/metrics|logger
internal/repository/postgres/migrations
deployments/docker|observability
```

## Быстрый старт

### 1. Инфра

```bash
docker compose --env-file .env. up -d
```

Поднимаются: TimescaleDB, Redis, Kafka, Mosquitto, Prometheus (`:9090`), Grafana (`:3000`, admin/admin).

### 2. Kafka-топики

```bash
for t in raw.telemetry raw.telemetry.dlq processed.events alerts commands; do
  docker exec smartgrid-kafka /opt/kafka/bin/kafka-topics.sh \
    --bootstrap-server localhost:9092 --create --if-not-exists \
    --topic "$t" --partitions 1 --replication-factor 1
done
```

### 3. Миграции

```bash
set -a && source .env.processor && set +a
go run github.com/pressly/goose/v3/cmd/goose@latest \
  -dir internal/repository/postgres/migrations \
  postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" up
```

### 4. Запуск сервисов (3 терминала)

```bash
set -a && source .env.processor && set +a && task run:processor
set -a && source .env.ingestion && set +a && task run:ingestion
set -a && source .env.api && set +a && task run:api
```

### 5. Health / ready / metrics

```bash
curl -sf http://localhost:8080/healthz
curl -sf http://localhost:8080/readyz
curl -sf http://localhost:8081/healthz
curl -sf http://localhost:8082/healthz
curl -sf http://localhost:8082/readyz
curl -s http://localhost:8082/metrics | head
```

`/healthz` — процесс жив. `/readyz` (api/processor) — ping Postgres + Redis (503 если недоступны).

### 6. Пайплайн + API

```bash
docker exec smartgrid-mosquitto mosquitto_pub -h localhost \
  -t 'smartmeter/zone1/dev-001' -m \
  '{"voltage":220.1,"current":5.2,"power":1144.5,"frequency":50,"measured_at":"2026-08-02T14:00:00Z"}'

curl -s 'http://localhost:8082/api/v1/devices'
curl -s 'http://localhost:8082/api/v1/devices/dev-001/telemetry/latest'
curl -s 'http://localhost:8082/api/v1/devices/dev-001/alerts'
curl -s -X POST 'http://localhost:8082/api/v1/devices/dev-001/commands' \
  -H 'Content-Type: application/json' \
  -d '{"command":"shed_load","params":{"percent":"10"}}'
```

Postman: те же URL/методы.

### 7. Observability

- Prometheus: http://localhost:9090 (targets api/processor/ingestion)
- Grafana: http://localhost:3000 (admin/admin)

### 8. Код и hooks

```bash
task cleancode
brew install lefthook && lefthook install   # pre-commit → fmt/lint/test/build
```

## Task

| Команда | Назначение |
|---------|------------|
| `task docker:up` / `docker:down` | инфра |
| `task migrate` | миграции |
| `task run:processor` | processor `:8080` |
| `task run:ingestion` | ingestion `:8081` |
| `task run:api` | api `:8082` |
| `task cleancode` | fmt + lint + test + build |

## Kafka-топики

| Топик | Назначение |
|-------|------------|
| `raw.telemetry` | сырые показания |
| `raw.telemetry.dlq` | битые сообщения |
| `processed.events` | результат обработки |
| `alerts` | алерты |
| `commands` | команды на устройства |

