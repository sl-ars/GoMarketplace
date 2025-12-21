package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger представляет структурированный логгер
type Logger struct {
	*logrus.Logger
}

// Config конфигурация логгера
type Config struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"` // json или text
}

// New создает новый экземпляр логгера
func New(config Config) *Logger {
	logger := logrus.New()

	// Устанавливаем уровень логирования
	level, err := logrus.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Устанавливаем формат вывода
	if config.Format == "text" {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	// Устанавливаем вывод в stdout
	logger.SetOutput(os.Stdout)

	return &Logger{Logger: logger}
}

// WithFields создает логгер с дополнительными полями
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithError создает логгер с полем ошибки
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithRequest создает логгер с полями HTTP запроса
func (l *Logger) WithRequest(method, path, userAgent string, userID interface{}) *logrus.Entry {
	fields := logrus.Fields{
		"method":    method,
		"path":      path,
		"userAgent": userAgent,
	}
	if userID != nil {
		fields["userID"] = userID
	}
	return l.Logger.WithFields(fields)
}

// WithUser создает логгер с полем пользователя
func (l *Logger) WithUser(userID interface{}) *logrus.Entry {
	return l.Logger.WithField("userID", userID)
}

// WithOperation создает логгер с полем операции
func (l *Logger) WithOperation(operation string) *logrus.Entry {
	return l.Logger.WithField("operation", operation)
}
