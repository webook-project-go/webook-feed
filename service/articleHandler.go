package service

import (
	"context"
	"errors"
	"github.com/robfig/cron/v3"
	"github.com/webook-project-go/webook-feed/client"
	"github.com/webook-project-go/webook-feed/domain"
	"github.com/webook-project-go/webook-feed/repository"
	"github.com/webook-project-go/webook-pkgs/logger"
	"time"
)

type ArticleHandler struct {
	active     *client.ActiveClient
	relation   *client.RelationClient
	repo       repository.Repository
	threshold  int
	cronClient cron.Cron
	l          logger.Logger
}

func (a *ArticleHandler) CreateEvent(ctx context.Context, event domain.FeedEvent) error {
	followerCount, err := a.relation.GetFollowerCnt(ctx, event.ActorID)
	if err != nil {
		return a.repo.CreatePullEvent(ctx, event)
	}
	if followerCount < uint32(a.threshold) {
		uids, err := a.relation.GetFollowers(ctx, event.ActorID, 0, a.threshold)
		if err != nil {
			return err
		}
		lovers, err := a.active.ActiveFilters(ctx, uids)
		if err != nil {
			return err
		}
		go func() {
			timeout := 0
			for i := 0; i < len(lovers); i++ {
				event.TargetID = lovers[i]
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				err := a.repo.CreatePushEvent(ctx, event)
				cancel()
				if err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						timeout++
						if timeout >= 3 {
							return
						}
						continue
					}
					return
				}
			}
		}()
	}
	err = a.repo.CreatePullEvent(ctx, event)
	if err != nil {
		return err
	}
	go func() {
		err := a.repo.AddPushTask(ctx, event)
		if err != nil {
			a.l.Error("add push task failed", logger.Error(err))
		}
	}()
	return nil
}

func (a *ArticleHandler) GetEvents(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	events, err := a.repo.GetPushEvent(ctx, domain.ArticleEvent, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}

	pull := limit - len(events)
	if pull == 0 {
		return events, nil
	}
	followers, err := a.relation.GetFollowers(ctx, uid, 0, 1000)
	if err != nil {
		return events, nil
	}

	activeFollowers, err := a.active.ActiveFilters(ctx, followers)
	if err != nil {
		return events, nil
	}

	moreEvents, err := a.repo.GetPullEvent(ctx, domain.ArticleEvent, activeFollowers, timestamp, pull)
	if err != nil {
		return events, nil
	}

	events = append(events, moreEvents...)
	return events, nil
}
