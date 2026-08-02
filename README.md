# Smart Grid

Сервис(ы) алгоритмов умных сетей: приём телеметрии с IoT-устройств, обработка событий и хранение временных рядов.

## Что делает система

```
Devices → MQTT → Ingestion → Kafka (raw.telemetry)
                              ↓
                         Processor → TimescaleDB / Redis
                              ↓
                    processed.events / alerts / DLQ
                              ↓
                         API (позже)
```

| Сервис | Назначение |
|--------|------------|
| **ingestion** | Подписывается на MQTT, валидирует показания, пишет в Kafka `raw.telemetry`; битые сообщения → `raw.telemetry.dlq` |
| **processor** | Читает `raw.telemetry`, применяет алгоритмы, пишет в TimescaleDB/Redis и топики `processed.events` / `alerts` |

## Стек

- Go
- MQTT (Eclipse Mosquitto)
- Apache Kafka
- PostgreSQL / TimescaleDB
- Redis
- Docker Compose
- далее: Prometheus, Grafana, Kubernetes / Helm

## Структура

```
cmd/ingestion          — MQTT → Kafka
cmd/processor          — обработка телеметрии
internal/domain        — модели и события
internal/repository    — интерфейсы и адаптеры (MQTT, Kafka, Postgres, Redis)
internal/service       — бизнес-логика
internal/usecase       — use cases
internal/transport     — REST / Kafka / gRPC
internal/repository/postgres/migrations — goose-миграции
deployments/           — docker / helm / k8s
```

## Быстрый старт

### 1. Env

```bash
# инфра + processor
cp .env.example .env   # если нужно

# отдельно для ingestion (пример)
# .env.ingestion — APP_NAME, HTTP_ADDR, MQTT_*, KAFKA_*
```

`.env` не коммитить.

### 2. Инфраструктура

```bash
docker compose --env-file .env up -d
```

Поднимаются: TimescaleDB, Redis, Kafka, Mosquitto.

### 3. Kafka-топики (если auto-create выключен)

```bash
docker exec smartgrid-kafka /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 --create --if-not-exists \
  --topic raw.telemetry --partitions 3 --replication-factor 1

docker exec smartgrid-kafka /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 --create --if-not-exists \
  --topic raw.telemetry.dlq --partitions 1 --replication-factor 1

docker exec smartgrid-kafka /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 --create --if-not-exists \
  --topic processed.events --partitions 3 --replication-factor 1

docker exec smartgrid-kafka /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 --create --if-not-exists \
  --topic alerts --partitions 1 --replication-factor 1
```

### 4. Миграции

```bash
set -a && source .env && set +a
go run github.com/pressly/goose/v3/cmd/goose@latest \
  -dir internal/repository/postgres/migrations \
  postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" up
```

### 5. Запуск сервисов

```bash
# processor (health :8080)
task run

# ingestion (health :8081)
set -a && source .env.ingestion && set +a
task run:ingestion
```

Проверка:

```bash
curl -sf http://localhost:8080/healthz
curl -sf http://localhost:8081/healthz
```

### 6. Проверка пайплайна MQTT → Kafka

```bash
docker exec smartgrid-mosquitto mosquitto_pub -h localhost \
  -t 'smartmeter/zone1/dev-001' -m \
  '{"voltage":220.1,"current":5.2,"power":1144.5,"frequency":50,"measured_at":"2026-07-22T01:00:00Z"}'
```

В логах ingestion ожидается `telemetry ingested`.

Битое сообщение → DLQ:

```bash
docker exec smartgrid-mosquitto mosquitto_pub -h localhost \
  -t 'smartmeter/zone1/dev-001' -m '{"voltage":-1}'
```

### 7. Проверки кода

```bash
task cleancode   # fmt + lint + test + build
```

### 8 Git hooks

```bash
brew install lefthook
lefthook install
```

## Команды Task

| Команда | Назначение |
|---------|------------|
| `task docker:up` | поднять инфру |
| `task docker:down` | остановить инфру |
| `task migrate` | миграции |
| `task run` | processor |
| `task run:ingestion` | ingestion |
| `task smoke` | health processor |
| `task cleancode` | fmt + lint + test + build |

## Kafka-топики

| Топик | Назначение |
|-------|------------|
| `raw.telemetry` | сырые показания |
| `raw.telemetry.dlq` | битые сообщения |
| `processed.events` | результат обработки |
| `alerts` | алерты |
| `commands` | команды на устройства |

## Примечания

- Module: `github.com/itGeek-rus/smart-grid.git`
- Local Mosquitto: `allow_anonymous=true` (только для разработки)
- В DataGrip/GoLand подключайся к БД `smart-grid`, не к `postgres`
