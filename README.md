## URL Shortener (Go)

Минималистичное приложение для укорачивания ссылок: Go + вшитый фронтенд (HTML/CSS).

### Запуск (локально)

```bash
go run .
```

Откроется браузер: `http://localhost:8080`

### Сборка бинарника

```bash
go build -o url-shortener .
./url-shortener
```

### Конфигурация через флаги

```bash
./url-shortener -addr=":9090" -base-url="https://mysite.com" -rate-limit=50
```

Доступные флаги:
- `-addr` — адрес сервера (по умолчанию `:8080`)
- `-base-url` — базовый URL для коротких ссылок
- `-shutdown-timeout` — таймаут graceful shutdown (по умолчанию `10s`)
- `-rate-limit` — лимит запросов в минуту (по умолчанию `100`)
- `-rate-window` — окно для rate limiting (по умолчанию `1m`)

### Docker

```bash
docker build -t url-shortener .
docker run --rm -p 8080:8080 url-shortener
```

### Использование

- Вставьте длинный URL в поле и нажмите «Укоротить».
- Получите короткую ссылку вида `http://localhost:8080/s/xxxxxxx`.
- Переход по короткой ссылке выполнит редирект на исходный адрес.

### Go-приколюхи

- **JSON-логирование** — структурированные логи с `slog`
- **Метрики Prometheus** — счетчики запросов, время ответа, активные соединения (`/metrics`)
- **Rate Limiting** — защита от спама с настраиваемыми лимитами
- **CORS** — поддержка кросс-доменных запросов
- **Health Check** — эндпоинт `/health` для мониторинга
- **Graceful Shutdown** — корректное завершение с таймаутом
- **Middleware Chain** — композиция middleware для логирования, метрик, rate limiting

### Особенности

- Валидация и нормализация URL, авто-добавление схемы.
- Криптографически случайные короткие ID, защита от коллизий.
- Потокобезопасное in-memory хранилище.
- Вшитая статика (go:embed), авто-открытие браузера.

### Мониторинг

- Метрики Prometheus: `http://localhost:8080/metrics`
- Health check: `http://localhost:8080/health`
- JSON-логи в stdout




