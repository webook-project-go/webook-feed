package client

import (
	"context"
	v1 "github.com/webook-project-go/webook-apis/gen/go/apis/relation/v1"
	"github.com/webook-project-go/webook-feed/repository/cache"
)

type RelationClient struct {
	client v1.RelationServiceClient
	cache  cache.Cache
}

func NewRelationClient(cache cache.Cache, client v1.RelationServiceClient) *RelationClient {
	return &RelationClient{
		client: client,
		cache:  cache,
	}
}
func (r *RelationClient) GetFollowerCnt(ctx context.Context, uid int64) (uint32, error) {
	resp, err := r.client.GetFollowerCount(ctx, &v1.GetFollowerCountReq{Uid: uid})
	if err != nil {
		return 0, err
	}
	return resp.GetCount(), nil
}
func (r *RelationClient) GetFollowees(ctx context.Context, uid, lastId int64, limit int) ([]int64, error) {
	res, err := r.cache.GetFollowees(ctx, uid)
	if err == nil && len(res) > 0 {
		return res, nil
	}
	resp, err := r.client.GetFollowees(ctx, &v1.GetFolloweesReq{
		Uid:    uid,
		LastID: lastId,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}
	go r.cache.SetFollowees(ctx, uid, resp.GetUids())
	return resp.GetUids(), nil
}
func (r *RelationClient) GetFollowers(ctx context.Context, uid, lastId int64, limit int) ([]int64, error) {
	res, err := r.cache.GetFollowers(ctx, uid)
	if err == nil && len(res) > 0 {
		return res, nil
	}
	resp, err := r.client.GetFollowers(ctx, &v1.GetFollowersReq{
		Uid:    uid,
		LastID: lastId,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}
	go r.cache.SetFollowers(ctx, uid, resp.GetUids())
	return resp.GetUids(), nil
}
