package redismessage

import "fmt"

type Msg struct {
	ID        int //消息编号
	Topic     string
	Body      []byte
	Partition int // 分区
}

func (msg *Msg) PartitionTopic() string {
	return fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
}
