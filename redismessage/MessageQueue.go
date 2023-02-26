package redismessage

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type AllGroups struct {
	groups map[int]*ConsumerGroup // id->group
}

func NewAllGroups() *AllGroups {
	allgroups := new(AllGroups)
	allgroups.groups = make(map[int]*ConsumerGroup)
	return allgroups
}
func (allgroups *AllGroups) add(group *ConsumerGroup) {
	allgroups.groups[group.GroupID] = group
}

type MessageQueue struct {
	Client *redis.Client
	// consumer group, topic
	Topic2Group map[string]IntArray
	//topic , partition
	Topic2Partition map[string]IntArray
}

func NewMessageQueue() *MessageQueue {
	mq := new(MessageQueue)
	mq.Client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	mq.Topic2Group = make(map[string]IntArray)
	mq.Topic2Partition = make(map[string]IntArray)
	return mq
}

// register
func (mq *MessageQueue) Register(topic string, groupid int) {
	groups, ok := mq.Topic2Group[topic]
	if ok {
		mq.Topic2Group[topic] = append(groups, groupid)
		return
	}
	mq.Topic2Group[topic] = []int{groupid}
	fmt.Printf("group %d register %s\n", groupid, topic)
}

//test ,set partition 4
func (mq *MessageQueue) SetPartiton() {
	for topic, _ := range mq.Topic2Group {
		parts := []int{1, 2, 3, 4}
		mq.Topic2Partition[topic] = parts
	}
}

// push message to redis
func (mq *MessageQueue) MsgPush(ctx context.Context, msg *Msg) error {
	fmt.Printf("push message to topic : %s\n", msg.Topic)
	return mq.Client.LPush(ctx, msg.PartitionTopic(), msg.Body).Err()
}

// 消费，tag是不同的分发标志
func (mq *MessageQueue) MsgPull(ctx context.Context, allgroups *AllGroups, tag int) {
	// 广播：订阅topic的所有消费者，消费所有的分区消息
	// 组播：每个分区消息给每个消费者组的所有消费者
	// 单播：，每个分区消息给每个消费者组的某一个消费者
	// TODO 组播，单播
	// 1.广播，消息分发
	for {
		for topic, partition := range mq.Topic2Partition {
			for _, p := range partition {
				topicPart := fmt.Sprintf("%s:%d", topic, p)
				//fmt.Println(topicPart)
				body, err := mq.Client.LIndex(ctx, topicPart, -1).Bytes()
				if err != nil && !errors.Is(err, redis.Nil) {
					continue
				}
				if err := mq.Client.RPop(ctx, topicPart).Err(); err != nil {
					continue
				}
				if errors.Is(err, redis.Nil) {
					continue
				}
				// 分发消息,
				// TODO 消息消费成功才删除，即ack机制。
				group := mq.Topic2Group[topic]
				for _, gid := range group {
					consumerGroup := allgroups.groups[gid]
					consumerGroup.GroupConsume(topic, p, body, tag)
				}
			}
		}
	}
}

func (mq *MessageQueue) Del() {
	mq.Client.Close()
}
