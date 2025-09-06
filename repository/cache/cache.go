package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type Cache interface {
	SetFollowees(ctx context.Context, uid int64, followees []int64) error
	GetFollowees(ctx context.Context, uid int64) ([]int64, error)
	GetFollowers(ctx context.Context, uid int64) ([]int64, error)
	SetFollowers(ctx context.Context, uid int64, followers []int64) error
}
type cache struct {
	cmd redis.Cmdable
}

func NewCache(cmd redis.Cmdable) Cache {
	return &cache{cmd: cmd}
}
func (c *cache) followeeKey(uid int64) string {
	return fmt.Sprintf("user:followee:%d", uid)
}
func (c *cache) followerKey(uid int64) string {
	return fmt.Sprintf("user:follower:%d", uid)
}
func (c *cache) SetFollowees(ctx context.Context, uid int64, followees []int64) error {
	data, err := json.Marshal(&followees)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, c.followeeKey(uid), data, time.Minute*10).Err()
}
func (c *cache) GetFollowees(ctx context.Context, uid int64) ([]int64, error) {
	key := c.followeeKey(uid)
	res, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var followees []int64
	err = json.Unmarshal([]byte(res), &followees)
	if err != nil {
		return nil, err
	}
	return followees, nil
}
func (c *cache) SetFollowers(ctx context.Context, uid int64, followers []int64) error {
	data, err := json.Marshal(&followers)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, c.followerKey(uid), data, time.Minute*10).Err()
}
func (c *cache) GetFollowers(ctx context.Context, uid int64) ([]int64, error) {
	key := c.followerKey(uid)
	res, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var followers []int64
	err = json.Unmarshal([]byte(res), &followers)
	if err != nil {
		return nil, err
	}
	return followers, nil
}
