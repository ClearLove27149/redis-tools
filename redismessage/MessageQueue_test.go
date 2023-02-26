package redismessage

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func IDealMsg(msg *Msg) error {
	fmt.Printf("I Deal Msg : topic=%s, part=%d, body=%s\n", msg.Topic, msg.Partition, msg.Body)
	return nil
}
func YouDealMsg(msg *Msg) error {
	fmt.Printf("You Deal Msg : topic=%s, part=%d, body=%s\n", msg.Topic, msg.Partition, msg.Body)
	return nil
}

func Test1(t *testing.T) {
	Mq := NewMessageQueue()
	defer Mq.Del()

	allgroups := NewAllGroups()

	group1 := NewConsumerGroup(100)
	for i := 101; i < 105; i++ {
		consumer := NewConsumer(i, IDealMsg)
		group1.Add(consumer)
	}
	group2 := NewConsumerGroup(200)
	for i := 201; i < 205; i++ {
		consumer := NewConsumer(i, YouDealMsg)
		group2.Add(consumer)
	}

	allgroups.add(group1)
	allgroups.add(group2)

	Mq.Register("libin", 100)
	Mq.Register("hexiaolin", 200)
	Mq.SetPartiton()

	for i := 401; i < 405; i++ {
		go Mq.MsgPush(context.Background(), &Msg{
			ID:        i,
			Topic:     "libin",
			Partition: i - 400,
			Body:      []byte{'h', 'e', 'l', 'l', 'o'},
		})
	}
	for i := 501; i < 505; i++ {
		go Mq.MsgPush(context.Background(), &Msg{
			ID:        i,
			Topic:     "hexiaolin",
			Partition: i - 500,
			Body:      []byte{'l', 'l', 'l', 'l', 'o'},
		})
	}

	Mq.MsgPull(context.Background(), allgroups, 1)
	time.Sleep(10000 * time.Millisecond)
}
