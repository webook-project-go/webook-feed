package service

import (
	"context"
	"github.com/webook-project-go/webook-feed/domain"
)

type Service interface {
	CreatFeedEvent(ctx context.Context, event []domain.FeedEvent) error
	GetFeedEvent(ctx context.Context, uid, timestamp int64, limit int, typ domain.EventType) ([]domain.FeedEvent, error)
}

type Handler interface {
	CreateEvent(ctx context.Context, event domain.FeedEvent) error
	GetEvents(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error)
}
