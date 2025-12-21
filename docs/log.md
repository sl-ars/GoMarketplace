# Логирование в Go Marketplace API

## Обзор

В проекте реализовано структурированное логирование с использованием библиотеки `logrus`. Логирование интегрировано на всех уровнях приложения: HTTP handlers, сервисы, репозитории и middleware.

## Конфигурация

### Переменные окружения

```bash
# Уровень логирования (debug, info, warn, error, fatal, panic)
LOG_LEVEL=info

# Формат вывода (json или text)
LOG_FORMAT=json
```

### Уровни логирования

- **debug** - Детальная информация для отладки
- **info** - Общая информация о работе приложения
- **warn** - Предупреждения о потенциальных проблемах
- **error** - Ошибки, которые не останавливают работу приложения
- **fatal** - Критические ошибки, после которых приложение завершается
- **panic** - Критические ошибки с паникой

## Структура логов

### JSON формат (по умолчанию)

```json
{
  "level": "info",
  "msg": "User registered successfully",
  "operation": "user_register",
  "time": "2024-01-15T10:30:45.123Z",
  "userID": 123,
  "email": "user@example.com",
  "username": "john_doe"
}
```

### Text формат

```
2024-01-15 10:30:45 INFO[user_register] User registered successfully userID=123 email=user@example.com username=john_doe
```

## Использование в коде

### Базовое логирование

```go
import "go-app-marketplace/pkg/logger"

// Простое сообщение
log.Info("Application started")

// С ошибкой
log.WithError(err).Error("Failed to connect to database")

// С дополнительными полями
log.WithFields(logger.Fields{
    "userID": 123,
    "email": "user@example.com",
}).Info("User action completed")
```

### Специализированные методы

```go
// Логирование с операцией
log.WithOperation("user_register").Info("Starting registration")

// Логирование с пользователем
log.WithUser(userID).Info("User action")

// Логирование HTTP запроса
log.WithRequest("POST", "/api/register", "Mozilla/5.0", userID).Info("Request processed")
```

## Интеграция по уровням

### HTTP Handlers

Логирование запросов, валидации и ответов:

```go
func RegisterHandler(service *services.UserService, log *logger.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.WithOperation("user_register").Info("Starting user registration")
        
        // ... обработка запроса ...
        
        if err != nil {
            log.WithError(err).WithOperation("user_register").Error("Registration failed")
            return
        }
        
        log.WithFields(logger.Fields{
            "userID": res.ID,
            "email": req.Email,
        }).WithOperation("user_register").Info("User registered successfully")
    }
}
```

### Сервисы

Логирование бизнес-логики:

```go
func (s *UserService) Register(ctx context.Context, req *reqresp.RegisterUserRequest) (*reqresp.RegisterUserResponse, error) {
    s.logger.WithFields(logger.Fields{
        "email": req.Email,
        "username": req.Username,
    }).WithOperation("user_register").Info("Processing user registration")
    
    // ... бизнес-логика ...
    
    if err != nil {
        s.logger.WithError(err).WithOperation("user_register").Error("Failed to register user")
        return nil, err
    }
    
    return response, nil
}
```

### Репозитории

Логирование операций с базой данных:

```go
func (r *UserPostgresRepo) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
    r.logger.WithFields(logger.Fields{
        "email": user.Email,
        "username": user.Username,
    }).WithOperation("create_user").Info("Creating user in database")
    
    // ... SQL операция ...
    
    if err != nil {
        r.logger.WithError(err).WithOperation("create_user").Error("Failed to create user in database")
        return 0, err
    }
    
    r.logger.WithFields(logger.Fields{
        "userID": id,
        "email": user.Email,
    }).WithOperation("create_user").Info("User created successfully in database")
    
    return id, nil
}
```

### Middleware

Логирование HTTP запросов и аутентификации:

```go
func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // ... обработка запроса ...
            
            log.WithFields(logger.Fields{
                "method": r.Method,
                "path": r.URL.Path,
                "status": statusCode,
                "duration": time.Since(start).String(),
            }).Info("HTTP request processed")
        })
    }
}
```

## Мониторинг и анализ

### Рекомендуемые поля для мониторинга

- `operation` - Тип операции для группировки
- `userID` - Идентификатор пользователя для трассировки
- `duration` - Время выполнения операций
- `status` - HTTP статус код
- `error` - Детали ошибок

### Примеры запросов для анализа

```bash
# Все ошибки
grep '"level":"error"' logs/app.log

# Операции конкретного пользователя
grep '"userID":123' logs/app.log

# Медленные запросы (>1 секунды)
grep '"duration":"1' logs/app.log

# Операции регистрации
grep '"operation":"user_register"' logs/app.log
```

