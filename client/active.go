package client

import (
	"context"
	v1 "github.com/webook-project-go/webook-apis/gen/go/apis/active/v1"
)

type ActiveClient struct {
	client v1.ActiveServiceClient
}

func NewActivveClient(client v1.ActiveServiceClient) *ActiveClient {
	return &ActiveClient{client: client}
}

func (a *ActiveClient) IsActive(ctx context.Context, uid int64) (bool, error) {
	resp, err := a.client.IsActive(ctx, &v1.IsActiveRequest{Uid: uid})
	if err != nil {
		return false, err
	}
	return resp.GetActive(), nil
}
func (a *ActiveClient) LastActiveAt(ctx context.Context, uid int64) (int64, error) {
	resp, err := a.client.GetLastActiveAt(ctx, &v1.GetLastActiveAtRequest{Uid: uid})
	if err != nil {
		return 0, err
	}
	return resp.GetLastActiveAt(), nil
}

func (a *ActiveClient) ActiveFilters(ctx context.Context, uids []int64) ([]int64, error) {
	resp, err := a.client.ActiveFilters(ctx, &v1.ActiveFiltersRequest{Uids: uids})
	if err != nil {
		return nil, err
	}
	return resp.GetActiveUids(), nil
}
