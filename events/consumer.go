package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/kisara71/GoTemplate/slice"
	"github.com/webook-project-go/webook-feed/domain"
	"github.com/webook-project-go/webook-feed/service"
	"github.com/webook-project-go/webook-pkgs/logger"
	"github.com/webook-project-go/webook-pkgs/saramax"
	"time"
)

type kafkaConsumer struct {
	svc    service.Service
	client sarama.Client
	l      logger.Logger
}
type FeedEvent struct {
	TargetID int64             `json:"target_id"`
	ActorID  int64             `json:"actor_id"`
	Type     uint8             `json:"type"`
	MetaData map[string]string `json:"meta_data"`
	Ctime    int64             `json:"ctime"`
}

func NewKafkaConsumer(svc service.Service, client sarama.Client, l logger.Logger) *kafkaConsumer {
	return &kafkaConsumer{client: client, svc: svc, l: l}
}

func (c *kafkaConsumer) Start() {
	consumer, err := sarama.NewConsumerGroupFromClient("feed_event", c.client)
	if err != nil {
		panic(err)
	}
	go func() {
		hdl := saramax.NewBatchHandler(100, time.Millisecond*100, c.Consume, c.l)
		err := consumer.Consume(context.Background(), []string{"feed_event"}, hdl)
		if err != nil {
			c.l.Error("start consume failed", logger.Error(err))
		}
	}()
	return
}
func (c *kafkaConsumer) toDomainEvent(event FeedEvent) domain.FeedEvent {
	return domain.FeedEvent{
		TargetID: event.TargetID,
		ActorID:  event.ActorID,
		Type:     domain.TypeFromUint8(event.Type),
		MetaData: event.MetaData,
		Ctime:    event.Ctime,
	}
}
func (c *kafkaConsumer) Consume(messages []*sarama.ConsumerMessage, ts []FeedEvent) error {
	events, err := slice.Map(0, len(ts), ts, c.toDomainEvent)
	if err != nil {
		c.l.Error("convert events failed", logger.Error(err))
		return nil
	}
	err = c.svc.CreatFeedEvent(context.Background(), events)
	if err != nil {
		c.l.Error("create feed event failed", logger.Error(err))
		return err
	}
	return nil
}
