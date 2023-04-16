package redislock

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"testing"
)

// func Testexample(t *testing.T) {
// 	example()
// }

func TestTryLock(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0, // use default db
	})

	disLock, err := NewDisClock(client, "lock resource")
	if err != nil {
		log.Fatal(err)
	}
	succ, err := disLock.TryLock(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
	if succ {
		t.Logf("lock success")
		defer disLock.UnLock(context.Background())
	}
}
