package article

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

const TopicReadEvent = "article_read"

type Producer interface {
	ProducerReadEvent(evt ReadEvent) error
	ProducerReadEventV1(ctx context.Context, v1 ReadEventV1)
}

type ReadEvent struct {
	Aid int64
	Uid int64
}

type ReadEventV1 struct {
	Uids []int64
	Aids []int64
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

func (s *SaramaSyncProducer) ProducerReadEventV1(ctx context.Context, v1 ReadEventV1) {
}

func (s *SaramaSyncProducer) ProducerReadEvent(evt ReadEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicReadEvent,
		Value: sarama.StringEncoder(val),
	})
	return err
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{producer: producer}
}
