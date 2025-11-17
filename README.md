# GoTrain Marketplace

Онлайн-маркетплейс для продажи товаров с поддержкой множественных продавцов, корзины покупок, заказов и интеграции платежной системы Stripe.

## Возможности

- ✅ Clean Architecture
- ✅ PostgreSQL + sqlx
- ✅ Миграции через `golang-migrate`
- ✅ Валидация входных данных с `go-playground/validator`
- ✅ Swagger документация API
- ✅ Единый формат JSON ответов (success/error)
- ✅ Регистрация и аутентификация пользователей (JWT)
- ✅ Роли пользователей (admin, seller, customer)
- ✅ Управление продуктами и предложениями
- ✅ Корзина покупок
- ✅ Обработка заказов
- ✅ Интеграция со Stripe для платежей
- ✅ Обработка возвратов (refunds)
- ✅ Webhook обработка от Stripe

## Требования

- Go 1.23.4 или выше
- PostgreSQL 15 или выше
- Docker и Docker Compose (для запуска через Docker)
- Stripe аккаунт (для работы с платежами)

## Быстрый старт с Docker Compose

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd GoTrain
```

### 2. Настройка переменных окружения

Создайте файл `.env` в корне проекта или установите переменные окружения:

```env
HTTP_PORT=8080
DB_DSN=postgres://gotrain:gotrain_password@postgres:5432/gotrain_db?sslmode=disable
JWT_SECRET=your-secret-jwt-key-change-in-production
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret
LOG_LEVEL=info
```

**Важно**: Замените `JWT_SECRET`, `STRIPE_SECRET_KEY` и `STRIPE_WEBHOOK_SECRET` на реальные значения.

### 3. Запуск приложения

```bash
docker-compose up -d
```

Этот команда:
- Запустит PostgreSQL базу данных
- Соберет и запустит приложение
- Автоматически применит миграции базы данных

### 4. Проверка работы

```bash
# Проверка healthcheck
curl http://localhost:8080/api/health

# Просмотр логов
docker-compose logs -f app
```

### 5. Остановка приложения

```bash
docker-compose down

# С удалением данных БД
docker-compose down -v
```

## Локальный запуск (без Docker)

### 1. Установка зависимостей

```bash
go mod download
```

### 2. Установка golang-migrate

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate

# Windows
# Скачайте с https://github.com/golang-migrate/migrate/releases
```

### 3. Настройка PostgreSQL

Создайте базу данных:

```sql
CREATE DATABASE gotrain_db;
CREATE USER gotrain WITH PASSWORD 'gotrain_password';
GRANT ALL PRIVILEGES ON DATABASE gotrain_db TO gotrain;
```

### 4. Настройка переменных окружения

Создайте файл `configs/.env`:

```env
HTTP_PORT=8080
DB_DSN=postgres://gotrain:gotrain_password@localhost:5432/gotrain_db?sslmode=disable
JWT_SECRET=your-secret-jwt-key-change-in-production
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret
LOG_LEVEL=info
```

### 5. Применение миграций

```bash
migrate -path ./migrations -database "postgres://gotrain:gotrain_password@localhost:5432/gotrain_db?sslmode=disable" up
```

### 6. Запуск приложения

```bash
go run cmd/app/main.go -config ./configs/.env
```

Или скомпилируйте и запустите:

```bash
go build -o bin/app cmd/app/main.go
./bin/app -config ./configs/.env
```

## Генерация Swagger документации

```bash
# Установка swag
go install github.com/swaggo/swag/cmd/swag@latest

# Генерация документации
swag init -g cmd/app/main.go -o docs
```

После генерации документация будет доступна по адресу: `http://localhost:8080/swagger/index.html`

## Структура проекта

```
GoTrain/
├── cmd/
│   ├── app/              # Точка входа приложения
│   └── seed_users/       # Утилита для создания тестовых пользователей
├── configs/              # Файлы конфигурации
├── docs/                 # Swagger документация
├── internal/
│   ├── app/              # Инициализация приложения
│   ├── deliveries/       # HTTP handlers и роутинг
│   ├── middleware/       # Middleware (auth, logging, permissions)
│   ├── repositories/     # Репозитории для работы с БД
│   ├── services/         # Бизнес-логика
│   └── usecases/         # Use cases
├── migrations/           # SQL миграции
├── pkg/
│   ├── auth/             # JWT аутентификация
│   ├── domain/           # Доменные модели
│   ├── hash/             # Хеширование паролей
│   ├── httpx/            # HTTP утилиты
│   ├── logger/           # Логирование
│   └── reqresp/          # DTO для запросов/ответов
└── docker-compose.yml    # Docker Compose конфигурация
```

## API Endpoints

Основные эндпоинты API:

- `GET /api/health` - Healthcheck
- `POST /api/register` - Регистрация пользователя
- `POST /api/login` - Вход в систему
- `GET /api/products` - Список продуктов
- `POST /api/products` - Создание продукта (требует роль seller)
- `GET /api/cart` - Получить корзину
- `POST /api/cart/add` - Добавить товар в корзину
- `POST /api/orders` - Создать заказ
- `POST /api/webhook/stripe` - Webhook от Stripe

Полная документация API доступна в Swagger: `http://localhost:8080/swagger/index.html`

## Роли пользователей

- **admin** - Администратор системы
- **seller** - Продавец (может создавать продукты и предложения)
- **customer** - Покупатель (может делать заказы)

## Переменные окружения

| Переменная | Описание | Обязательная | По умолчанию |
|------------|----------|--------------|--------------|
| `HTTP_PORT` | Порт HTTP сервера | Нет | `8080` |
| `DB_DSN` | DSN для подключения к PostgreSQL | Да | - |
| `JWT_SECRET` | Секретный ключ для JWT токенов | Да | - |
| `STRIPE_SECRET_KEY` | Секретный ключ Stripe API | Да | - |
| `STRIPE_WEBHOOK_SECRET` | Секретный ключ для верификации webhook от Stripe | Да | - |
| `LOG_LEVEL` | Уровень логирования | Нет | `info` |

## Разработка

### Запуск тестов

```bash
go test ./...
```

### Форматирование кода

```bash
go fmt ./...
```

### Линтинг

```bash
golangci-lint run
```

## Документация

- [Архитектура приложения](./ARCHITECTURE.md) - Подробное описание архитектуры системы
- [Swagger документация](./docs/swagger.yaml) - API документация

## Лицензия

[Укажите лицензию]

## Авторы

[Укажите авторов]

---
