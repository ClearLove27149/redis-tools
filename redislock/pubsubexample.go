package redislock

import (
	"fmt"

	"github.com/go-redis/redis"
)

func example() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pubsub := client.Subscribe("mychannel")
	defer pubsub.Close()

	// 在另一个goroutine中进行订阅操作
	ch := pubsub.Channel()
	go func() {
		for msg := range ch {
			fmt.Println(msg.Channel, msg.Payload)
		}
	}()

	// 发布消息
	client.Publish("mychannel", "hello world")

	// 让main函数等待消息
	_, err := pubsub.Receive()
	if err != nil {
		panic(err)
	}
}
