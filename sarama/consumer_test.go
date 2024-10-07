package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	//	正常来说，一个消费者都是归属于一个消费者组的
	// 消费者就是你的业务
	consumer, err := sarama.NewConsumerGroup(addrs, "test_group", cfg)
	require.NoError(t, err)

	// 带超时的 context
	start := time.Now()
	//ctx, canal := context.WithTimeout(context.Background(), time.Second*10)
	ctx, cancel := context.WithCancel(context.Background())
	//defer canal()
	time.AfterFunc(time.Minute*10, func() {
		cancel()
	})
	err = consumer.Consume(ctx, []string{"test_topic"}, testConsumerGroupHandler{})
	//	你消费结束，就会到这里
	t.Log(err, time.Since(start).String())
}

type testConsumerGroupHandler struct {
}

func (t testConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	partitions := session.Claims()["test_topic"]

	for _, part := range partitions {
		session.ResetOffset("test_topic", part,
			sarama.OffsetOldest, "")
		//session.ResetOffset("test_topic", part,
		//	sarama.OffsetNewest, "")
		//session.ResetOffset("test_topic", part,
		//	123, "")
	}
	return nil
}

func (t testConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Cleanup")
	return nil
}

func (t testConsumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	//for msg := range msgs {
	//	m1 := msg
	//	go func() {
	//		//	消费 msg
	//		log.Println(string(m1.Value))
	//		session.MarkMessage(m1, "")
	//	}()
	//}
	//	什么情况下会到这里
	//	message 被人关了，也就是退出消费逻辑
	const batchSize = 10
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		var eg errgroup.Group
		var last *sarama.ConsumerMessage
	overLabel:
		for i := 0; i < batchSize; i++ {
			select {
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				last = msg
				eg.Go(func() error {
					// 你在这里重试
					log.Println(string(msg.Value))
					return nil
				})
			case <-ctx.Done():
				break overLabel
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			//	这边能怎么办？
			//	记录日志
			continue
		}
		if last != nil {
			session.MarkMessage(last, "")
		}
	}
}

func (t testConsumerGroupHandler) ConsumeClaimV1(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		//var bizMsg MyBizMsg
		//err := json.Unmarshal(msg.Value, &bizMsg)
		//if err != nil {
		//	//	这就是消费消息出错
		//	//	大多数时候就是重试
		//	//	记录日志
		//	continue
		//}
		log.Println(string(msg.Value))
		session.MarkMessage(msg, "")
	}
	//	什么情况下会到这里
	//	message 被人关了，也就是退出消费逻辑
	return nil
}

type MyBizMsg struct {
	Name string
}

// 返回只读的 channel
func ChannelV1() <-chan struct{} {
	panic("implement me")
}

// 返回可读可写的 channel
func ChannelV2(t *testing.T) chan struct{} {
	panic("implement me")
}

// 返回只写 channel
func ChannelV3() chan<- struct{} {
	panic("implement me")
}
