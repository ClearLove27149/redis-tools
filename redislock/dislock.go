package redislock

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
)

const (
	defaultExp = 10 * time.Second
	sleepDur   = 10 * time.Millisecond
)

/***
** SetNX
**/
type Disclock struct {
	Client     *MyRedisClient
	Key        string
	uuid       string
	cancelFunc context.CancelFunc
}

/***
使用cancel创建可取消的context，一般用来关闭协程
**/
func NewDisClock(client *MyRedisClient, key string) (*Disclock, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	return &Disclock{
		Client: client,
		Key:    key,
		uuid:   id.String(),
	}, nil
}
func (dClock *Disclock) TryLock(ctx context.Context) (bool, error) {
	ok, err := dClock.Client.SetNX(ctx, dClock.Key, dClock.uuid, defaultExp).Result()
	if err != nil {
		return false, err
	}
	newCtx, cancel := context.WithCancel(ctx)
	dClock.cancelFunc = cancel
	dClock.refresh(newCtx)
	return ok, nil
}

// 一个守护协程，一直给锁续期
func (dClock *Disclock) refresh(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(defaultExp / 4)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dClock.Client.Expire(ctx, dClock.Key, defaultExp)
			}
		}
	}()
}

//spinlock
func (dClock *Disclock) SpinLock(ctx context.Context, retryCounts int) (bool, error) {
	for i := 0; i < retryCounts; i++ {
		resp, err := dClock.TryLock(ctx)
		if err != nil {
			return false, err
		}
		if resp {
			return resp, nil
		}
		time.Sleep(sleepDur)
	}
	return false, nil
}

// unlock
func (dClock *Disclock) UnLock(ctx context.Context) (bool, error) {
	res, err := dClock.Client.Eval(ctx, compareAndDeleteScript, []string{dClock.Key}, dClock.uuid).Result()
	if err != nil {
		return false, err
	}
	if res == success {
		dClock.cancelFunc()
	}
	return true, nil
}

/***
可重入锁
return :
1 : lock
2 : lock repeat
-2 : key不存在
-1 : key存在，但没有设置剩余生存时间
剩余时间：else
***/
func (dClock *Disclock) TryLockRepeat(ctx context.Context) (int64, error) {
	okk, err := dClock.Client.Eval(ctx, repeatLockScript, []string{dClock.Key}, defaultExp, dClock.uuid).Result()
	ok := okk.(int64)
	if err != nil {
		return ok, err
	}
	if ok != 1 && ok != 2 {
		log.Fatalf("try lock fail, return the exist lock expired time *{ok}\n")
		return ok, nil
	}
	newCtx, cancel := context.WithCancel(ctx)
	dClock.cancelFunc = cancel
	dClock.refresh(newCtx)
	return ok, nil
}

// unlock repeat
// 2 : unlock, 1 : 计数器-1，锁还在
func (dClock *Disclock) UnLockRepeat(ctx context.Context) (int64, error) {
	okk, err := dClock.Client.Eval(ctx, repeatUnLockScript, []string{dClock.Key}, defaultExp, dClock.uuid).Result()
	if err != nil {
		return -3, err
	}
	ok := okk.(int64)
	if ok != 1 && ok != 2 {
		log.Fatalf("repeat unlock fail, key or uuid is not exist\n")
		return ok, nil
	}
	if ok == 2 {
		dClock.cancelFunc()
	}
	return ok, nil
}

// 基于订阅发布，不能在包装client，不然缺很多方法
// 基于订阅发布实现，避免尝试轮询加锁
// sub
func (dClock *Disclock) TryLockRepeatSub(ctx context.Context) (int64, error) {
	for {
		// 订阅一个频道topic
		pubsub := dClock.Client.Subscribe(ctx, "pubsub")
		defer pubsub.Close()

		_, err := pubsub.Receive(ctx)
		if err != nil {
			return -3, err
		}

		ch := pubsub.Channel()
		for msg := range ch {
			fmt.Println(msg.Channel, msg.Payload)
			if msg.Payload == "ok" {
				break
			}
		}

		okk, err := dClock.Client.Eval(ctx, repeatLockSubScript, []string{dClock.Key}, defaultExp, dClock.uuid).Result()
		// 加锁失败
		ok := okk.(int64)
		if err != nil {
			return ok, err
		}
		if ok != 1 && ok != 2 {
			log.Fatalf("try lock fail use sub, return the exist lock expired time #{ok}\n")
			return ok, nil
		}
		newCtx, cancel := context.WithCancel(ctx)
		dClock.cancelFunc = cancel
		dClock.refresh(newCtx)
		return ok, nil
	}
}

// cmd key, topic, time, uuid
func (dClock *Disclock) UnLockRepeatPub(ctx context.Context) (int64, error) {
	okk, err := dClock.Client.Eval(ctx, repeatUnLockPubScript, []string{dClock.Key, topic}, defaultExp, dClock.uuid).Result()
	if err != nil {
		return -3, err
	}
	ok := okk.(int64)
	if ok != 1 && ok != 2 {
		log.Fatalf("unlock fail use pub, key or uuid not exist\n")
		return ok, nil
	}
	if ok == 2 {
		dClock.cancelFunc()
	}
	return ok, nil
}
