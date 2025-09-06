package service

import (
	"context"
	"errors"
	"github.com/robfig/cron/v3"
	"github.com/webook-project-go/webook-feed/client"
	"github.com/webook-project-go/webook-feed/domain"
	"github.com/webook-project-go/webook-feed/repository"
	"github.com/webook-project-go/webook-pkgs/logger"
	_ "github.com/webook-project-go/webook-pkgs/logger"
	"time"
)

type service struct {
	handler  map[domain.EventType]Handler
	l        logger.Logger
	repo     repository.Repository
	relation *client.RelationClient
	active   *client.ActiveClient
}

func NewService(l logger.Logger, repo repository.Repository, relation *client.RelationClient,
	active *client.ActiveClient) Service {
	s := &service{
		handler:  make(map[domain.EventType]Handler),
		l:        l,
		repo:     repo,
		relation: relation,
		active:   active,
	}
	c := cron.New()
	_, err := c.AddFunc("0 3 * * *", func() {
		s.processPushEvents()
	})
	if err != nil {
		panic(err)
	}
	c.Start()
	return s
}

func (s *service) Register(typ domain.EventType, hdl Handler) {
	s.handler[typ] = hdl
}
func (s *service) CreatFeedEvent(ctx context.Context, events []domain.FeedEvent) error {
	for _, event := range events {
		go func() {
			hdl, ok := s.handler[event.Type]
			if !ok {
				s.l.Error("invalid type of event", logger.Int32("type", int32(event.Type)))
				return
			}
			err := hdl.CreateEvent(ctx, event)
			if err != nil {
				s.l.Error("creat event failed", logger.Error(err))
			}
		}()

	}
	return nil
}

func (s *service) GetFeedEvent(ctx context.Context, uid, timestamp int64, limit int, typ domain.EventType) ([]domain.FeedEvent, error) {
	hdl, ok := s.handler[typ]
	if !ok {
		return nil, errors.New("invalid type of event")
	}
	res, err := hdl.GetEvents(ctx, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (s *service) processPushEvents() {
	for {
		task, ok, err := s.repo.TryClaimTask(context.Background())
		if ok {
			go s.processTask(task)
		} else if errors.Is(err, repository.ErrRecordNotFound) {
			return
		}
		continue
	}
}
func (s *service) Run() {
	s.processPushEvents()
}

type Task struct {
	UID   int64
	Event *domain.FeedEvent
}

func (s *service) startWorkerPool(workerNum int, tasks <-chan Task) {
	for i := 0; i < workerNum; i++ {
		go func(id int) {
			for task := range tasks {
				_ = s.repo.CreatePushEventFromTask(context.Background(), *task.Event, task.UID)
			}
		}(i)
	}
}

func (s *service) processTask(task domain.FeedEvent) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	fanscount, err := s.relation.GetFollowerCnt(ctx, task.ActorID)
	cancel()
	if err != nil {
		return
	}
	var workerNum, taskNum int
	switch {
	case fanscount < 10000:
		workerNum = 10
		taskNum = 200
	case fanscount < 100000:
		workerNum = 100
		taskNum = 2000
	default:
		workerNum = 200
		taskNum = 4000
	}
	tasks := make(chan Task, taskNum)
	s.startWorkerPool(workerNum, tasks)
	lastID, limit := int64(0), 2000
	retry := 0
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		uids, err := s.relation.GetFollowees(ctx, task.ActorID, lastID, limit)
		cancel()
		if err != nil {
			if !errors.Is(err, context.DeadlineExceeded) {
				return
			}
			if retry >= 5 {
				return
			}
			retry++
			continue
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		lover, err := s.active.ActiveFilters(ctx, uids)
		cancel()
		if err != nil {
			if !errors.Is(err, context.DeadlineExceeded) {
				return
			}
			if retry >= 5 {
				return
			}
			retry++
			continue
		}
		for idx := 0; idx < len(lover); idx++ {
			tasks <- Task{
				UID:   lover[idx],
				Event: &task,
			}
		}
		if len(uids) < limit {
			close(tasks)
			return
		}
		lastID = uids[len(uids)-1]
	}
}
