package feishu2wework

import (
	"context"
	"sync_roster/service/do"
)

type IFeishu interface {
	GetUser(ctx context.Context, config do.FeishuConfig, id interface{}) (do.FeishuUser, error)
	GetDept(ctx context.Context, config do.FeishuConfig, id string) (*do.FeishuDeptTree, error)
	GetAllDeptUser(ctx context.Context, config do.FeishuConfig) (*do.FeishuDeptTree, []do.FeishuUser, error)
}

type IWework interface {
	GetUser(ctx context.Context, config do.WeworkConfig, id interface{}) (do.WeworkUser, error)
	GetUsrByMobile(ctx context.Context, config do.WeworkConfig, mobile string) (do.WeworkUser, error)
	UpdateUser(ctx context.Context, config do.WeworkConfig, user do.WeworkUser) error
	LeaveUser(ctx context.Context, config do.WeworkConfig, id string) error
	CrateUser(ctx context.Context, config do.WeworkConfig, user do.WeworkUser) (do.WeworkUser, error)
	CrateDept(ctx context.Context, config do.WeworkConfig, dept do.WeworkDeptCreateInfo) (do.WeworkDeptInfo, error)
	UpdateDept(ctx context.Context, config do.WeworkConfig, updateInfo do.WeworkDeptUpdateInfo) error
	DeleteDept(ctx context.Context, config do.WeworkConfig, id int) error
	GetDept(ctx context.Context, config do.WeworkConfig, id int) (*do.WeworkDeptTree, error)
	GetAllDeptUser(ctx context.Context, config do.WeworkConfig) (*do.WeworkDeptTree, []do.WeworkUser, error)
}
