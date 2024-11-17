package ioc

import (
	"github.com/IBM/sarama"
	events2 "github.com/WeiXinao/basic-go/webook/interactive/events"
	"github.com/WeiXinao/basic-go/webook/internal/events"
	"github.com/spf13/viper"
)

func InitSaramaClient() sarama.Client {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(cfg.Addr, scfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitConsumers(c1 *events2.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
