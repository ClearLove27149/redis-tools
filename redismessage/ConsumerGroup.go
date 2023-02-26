package redismessage

import "fmt"

type MsgHandler func(msg *Msg) error

type Consumer struct {
	ID      int
	group   *ConsumerGroup // 属于哪个消费者组
	handler MsgHandler     // 消费者的处理逻辑
}

type IntArray []int

type ConsumerGroup struct {
	GroupID   int
	CurGroup  []*Consumer
	TopicInfo map[string]IntArray
}

func NewConsumer(id int, handle MsgHandler) *Consumer {
	consumer := Consumer{
		ID:      id,
		handler: handle,
	}
	return &consumer
}

func (consumer *Consumer) Consume(topic string, part int, body []byte, tag int) (bool, error) {
	fmt.Println(topic, part)
	msg := &Msg{
		Topic:     topic,
		Partition: part,
		Body:      body,
	}
	err := consumer.handler(msg)

	if err != nil {
		return false, err
	}
	return true, nil
}

func NewConsumerGroup(id int) *ConsumerGroup {
	return &ConsumerGroup{
		GroupID: id,
	}
}

//add
func (group *ConsumerGroup) Add(consumer *Consumer) bool {
	group.CurGroup = append(group.CurGroup, consumer)
	consumer.group = group
	return true
}

//del
func (group *ConsumerGroup) Del(consumer *Consumer) bool {
	for i, con := range group.CurGroup {
		if con.ID == consumer.ID {
			group.CurGroup = append(group.CurGroup[:i], group.CurGroup[i+1:]...)
			consumer.group = nil
			return true
		}
	}
	return false
}

// group consume
func (group *ConsumerGroup) GroupConsume(topic string, part int, body []byte, tag int) bool {
	// TODO
	// tag 不同的消息分发方式
	for _, consumer := range group.CurGroup {
		go consumer.Consume(topic, part, body, tag)
	}
	return true
}
