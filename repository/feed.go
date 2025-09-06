package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/webook-project-go/webook-feed/domain"
	"github.com/webook-project-go/webook-feed/repository/cache"
	"github.com/webook-project-go/webook-feed/repository/dao"
	"time"
)

var (
	ErrRecordNotFound = dao.ErrRecordNotFound
	ErrTaskOccupied   = dao.ErrTaskOccupied
)

type Repository interface {
	CreatePullEvent(ctx context.Context, event domain.FeedEvent) error
	CreatePushEvent(ctx context.Context, event domain.FeedEvent) error
	CreatePushEventFromTask(ctx context.Context, task domain.FeedEvent, targetID int64) error
	GetPushEvent(ctx context.Context, typ domain.EventType, targetId, lastID int64, limit int) ([]domain.FeedEvent, error)
	GetPullEvent(ctx context.Context, typ domain.EventType, actorIds []int64, timeline int64, limit int) ([]domain.FeedEvent, error)

	AddPushTask(ctx context.Context, event domain.FeedEvent) error
	TryClaimTask(ctx context.Context) (domain.FeedEvent, bool, error)
}

type repository struct {
	cache cache.Cache
	db    dao.Dao
}

func NewRepository(cache cache.Cache, db dao.Dao) Repository {
	return &repository{
		cache: cache,
		db:    db,
	}
}
func (r *repository) CreatePushEventFromTask(ctx context.Context, task domain.FeedEvent, targetID int64) error {
	task.TargetID = targetID
	return r.CreatePushEvent(ctx, task)
}

func (r *repository) taskToDomain(task *dao.PushTask) (domain.FeedEvent, error) {
	var data map[string]string
	err := json.Unmarshal(task.MetaData, &data)
	if err != nil {
		return domain.FeedEvent{}, err
	}
	return domain.FeedEvent{
		ActorID:  task.ActorID,
		Type:     domain.TypeFromUint8(task.Type),
		MetaData: data,
		Ctime:    task.Ctime,
	}, nil
}
func (r *repository) TryClaimTask(ctx context.Context) (domain.FeedEvent, bool, error) {
	res, ok, err := r.db.TryClaimTask(ctx)

	if err != nil {
		if errors.Is(err, dao.ErrTaskOccupied) {
			return domain.FeedEvent{}, false, ErrTaskOccupied
		} else if errors.Is(err, dao.ErrRecordNotFound) {
			return domain.FeedEvent{}, false, ErrRecordNotFound
		}
		return domain.FeedEvent{}, false, err
	}
	event, err := r.taskToDomain(res)
	if err != nil {
		return domain.FeedEvent{}, false, err
	}
	return event, ok, nil
}

func (r *repository) inboxToDomain(inbox dao.Inbox) (domain.FeedEvent, error) {
	var data map[string]string
	err := json.Unmarshal(inbox.MetaData, &data)
	if err != nil {
		return domain.FeedEvent{}, err
	}
	return domain.FeedEvent{
		ID:       inbox.ID,
		TargetID: inbox.TargetID,
		ActorID:  inbox.ActorID,
		Type:     domain.TypeFromUint8(inbox.Type),
		MetaData: data,
		Ctime:    inbox.Ctime,
	}, nil
}
func (r *repository) outboxToDomain(outbox dao.OutBox) (domain.FeedEvent, error) {
	var data map[string]string
	err := json.Unmarshal(outbox.MetaData, &data)
	if err != nil {
		return domain.FeedEvent{}, err
	}
	return domain.FeedEvent{
		ID:       outbox.ID,
		ActorID:  outbox.ActorID,
		Type:     domain.TypeFromUint8(outbox.Type),
		MetaData: data,
		Ctime:    outbox.Ctime,
	}, nil
}
func (r *repository) domainToInbox(event domain.FeedEvent) (dao.Inbox, error) {
	data, err := json.Marshal(&event.MetaData)
	if err != nil {
		return dao.Inbox{}, err
	}
	return dao.Inbox{
		TargetID: event.TargetID,
		ActorID:  event.ActorID,
		MetaData: data,
		Type:     uint8(event.Type),
	}, nil
}
func (r *repository) CreatePushEvent(ctx context.Context, event domain.FeedEvent) error {
	inbox, err := r.domainToInbox(event)
	if err != nil {
		return err
	}
	err = r.db.AddToInbox(ctx, inbox)
	return err
}

func (r *repository) CreatePullEvent(ctx context.Context, event domain.FeedEvent) error {
	outbox, err := r.domainToOutbox(event)
	if err != nil {
		return err
	}
	return r.db.AddToOutBox(ctx, outbox)
}

func (r *repository) GetPushEvent(ctx context.Context, typ domain.EventType, targetId, lastID int64, limit int) ([]domain.FeedEvent, error) {
	data, err := r.db.GetInbox(ctx, targetId, lastID, uint8(typ), limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.FeedEvent, 0, len(data))
	for i := 0; i < len(data); i++ {
		d, err := r.inboxToDomain(data[i])
		if err != nil {
			continue
		}
		res = append(res, d)
	}
	return res, nil
}

func (r *repository) GetPullEvent(ctx context.Context, typ domain.EventType, actorIds []int64, timeline int64, limit int) ([]domain.FeedEvent, error) {
	data, err := r.db.GetOutbox(ctx, actorIds, timeline, uint8(typ), limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.FeedEvent, 0, len(data))
	for i := 0; i < len(data); i++ {
		d, err := r.outboxToDomain(data[i])
		if err != nil {
			continue
		}
		res = append(res, d)
	}
	return res, nil
}

func (r *repository) toTask(event domain.FeedEvent) (dao.PushTask, error) {
	data, err := json.Marshal(&event.MetaData)
	if err != nil {
		return dao.PushTask{}, err
	}
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)

	executeAt := time.Date(
		tomorrow.Year(),
		tomorrow.Month(),
		tomorrow.Day(),
		2, 55, 0, 0,
		now.Location(),
	)
	return dao.PushTask{
		ActorID:   event.ActorID,
		MetaData:  data,
		Type:      uint8(event.Type),
		ExecuteAt: &executeAt,
		Status:    0,
	}, nil
}
func (r *repository) AddPushTask(ctx context.Context, event domain.FeedEvent) error {
	en, err := r.toTask(event)
	if err != nil {
		return err
	}
	return r.db.AddPushTask(ctx, en)
}

func (r *repository) domainToOutbox(event domain.FeedEvent) (dao.OutBox, error) {
	data, err := json.Marshal(&event.MetaData)
	if err != nil {
		return dao.OutBox{}, err
	}
	return dao.OutBox{
		ActorID:  event.ActorID,
		MetaData: data,
		Type:     uint8(event.Type),
	}, nil
}
