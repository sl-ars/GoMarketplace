package redisdb

import (
	"context"
	"encoding/json"
	"time"
)

func CacheGetOrSet[T any](ctx context.Context, key string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	var zero T

	val, err := Rdb.Get(ctx, key).Result()
	if err == nil {
		var result T
		if json.Unmarshal([]byte(val), &result) == nil {
			//fmt.Println("⚡️ Взято из кэша:", key)
			return result, nil
		}
	}

	//fmt.Println("💾 Берём из БД и записываем в Redis:", key)
	data, err := fetch()
	if err != nil {
		return zero, err
	}

	jsonData, _ := json.Marshal(data)
	Rdb.Set(ctx, key, jsonData, ttl)
	return data, nil
}
