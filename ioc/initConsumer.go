package ioc

import (
	"github.com/IBM/sarama"
	"github.com/webook-project-go/webook-feed/events"
	"github.com/webook-project-go/webook-feed/service"
	"github.com/webook-project-go/webook-pkgs/logger"
)

func InitConsumer(client sarama.Client, l logger.Logger, svc service.Service) []events.Consumer {
	feedConsumer := events.NewKafkaConsumer(svc, client, l)
	return []events.Consumer{feedConsumer}
}
