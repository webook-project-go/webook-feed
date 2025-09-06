package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRecordNotFound = errors.New("no record")
	ErrTaskOccupied   = errors.New("task occupied")
)

type Dao interface {
	AddPushTask(ctx context.Context, task PushTask) error
	TryClaimTask(ctx context.Context) (*PushTask, bool, error)

	AddToInbox(ctx context.Context, inbox Inbox) error
	AddToOutBox(ctx context.Context, outbox OutBox) error
	GetInbox(ctx context.Context, targetID int64, lastID int64, typ uint8, limit int) ([]Inbox, error)
	GetOutbox(ctx context.Context, uids []int64, timeLine int64, typ uint8, limit int) ([]OutBox, error)
}
type dao struct {
	db *gorm.DB
}

func (d *dao) AddToInbox(ctx context.Context, inbox Inbox) error {
	return d.db.WithContext(ctx).Create(&inbox).Error
}

func (d *dao) AddToOutBox(ctx context.Context, outbox OutBox) error {
	return d.db.WithContext(ctx).Create(&outbox).Error
}

func NewDao(db *gorm.DB) Dao {
	return &dao{
		db: db,
	}
}
func (d *dao) GetInbox(ctx context.Context, targetID int64, lastID int64, typ uint8, limit int) ([]Inbox, error) {
	var inbox []Inbox
	err := d.db.WithContext(ctx).Where("id > ? AND target_id = ? AND typ = ?", lastID, targetID, typ).
		Limit(limit).Find(&inbox).Error
	return inbox, err
}
func (d *dao) GetOutbox(ctx context.Context, actorIds []int64, timeLine int64, typ uint8, limit int) ([]OutBox, error) {
	var outbox []OutBox
	err := d.db.WithContext(ctx).Where(" actor_id in ? AND ctime > ? AND typ = ?", actorIds, timeLine, typ).
		Limit(limit).Find(&outbox).Error
	return outbox, err
}
func (d *dao) AddPushTask(ctx context.Context, task PushTask) error {
	return d.db.WithContext(ctx).Create(&task).Error
}

func (d *dao) TryClaimTask(ctx context.Context) (*PushTask, bool, error) {
	var res PushTask
	err := d.db.WithContext(ctx).Where("execute_at < Now() AND status = 0").First(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, ErrRecordNotFound
		}
		return nil, false, err
	}
	result := d.db.WithContext(ctx).Where("id = ? AND status = ?", res.ID, res.Status).Updates(clause.Assignments(map[string]any{
		"status": 1,
	}))
	if result.RowsAffected == 0 {
		return nil, false, ErrTaskOccupied
	}
	return &res, true, nil
}
