package grpc

import (
	"context"
	"github.com/kisara71/GoTemplate/slice"
	v1 "github.com/webook-project-go/webook-apis/gen/go/apis/feed/v1"
	"github.com/webook-project-go/webook-feed/domain"
	"github.com/webook-project-go/webook-feed/service"
)

type Service struct {
	svc service.Service
	v1.UnimplementedFeedServiceServer
}

func NewService(svc service.Service) *Service {
	return &Service{
		svc: svc,
	}
}

func (s *Service) toProto(event domain.FeedEvent) *v1.FeedEvent {
	return &v1.FeedEvent{
		Id:       event.ID,
		TargetId: event.TargetID,
		ActorId:  event.ActorID,
		Type:     v1.EventType(int32(event.Type)),
		Metadata: event.MetaData,
		Ctime:    event.Ctime,
	}
}
func (s *Service) GetFeedEvent(ctx context.Context, req *v1.GetFeedEventRequest) (*v1.GetFeedEventResponse, error) {
	res, err := s.svc.GetFeedEvent(ctx, req.GetUid(), req.GetUid(), int(req.GetLimit()), domain.TypeFromUint8(uint8(req.GetType())))
	if err != nil {
		return nil, err
	}
	data, err := slice.Map(0, len(res), res, s.toProto)
	if err != nil {
		return nil, err
	}
	return &v1.GetFeedEventResponse{Events: data}, nil
}
