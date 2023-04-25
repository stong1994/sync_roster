package feishu2wework

import (
	"context"
	"sync_roster/service/do"
)

type IFeishu interface {
	GetUser(ctx context.Context, config do.FeishuConfig, id interface{}) (do.FeishuUser, error)
}

type IWework interface {
	GetUser(ctx context.Context, config do.WeworkConfig, id interface{}) (do.WeworkUser, error)
	GetUsrByMobile(ctx context.Context, config do.WeworkConfig, mobile string) (do.WeworkUser, error)
	UpdateUser(ctx context.Context, config do.WeworkConfig, user do.WeworkUser) error
	LeaveUser(ctx context.Context, config do.WeworkConfig, id string) error
	CrateUser(ctx context.Context, config do.WeworkConfig, user do.WeworkUser) (do.WeworkUser, error)
}
