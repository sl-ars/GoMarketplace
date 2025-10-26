package redisdb

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx = context.Background()
	Rdb *redis.Client
)

func Init() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // порт Redis
		Password: "",               // если без пароля
		DB:       0,                // дефолтная база
	})

	if err := Rdb.Ping(Ctx).Err(); err != nil {
		log.Fatalf("❌ Ошибка подключения к Redis: %v", err)
	}

	log.Println("✅ Подключение к Redis успешно")
}
