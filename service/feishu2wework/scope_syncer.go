package feishu2wework

import (
	"context"
	"fmt"
	"sync_roster/service/do"
	"sync_roster/service/sync"
)

type ScopeSyncer struct {
	ctx context.Context

	deptSyncer *DeptSyncer
	userSyncer *UserSyncer

	feishuConfig  do.FeishuConfig
	weworkConfig  do.WeworkConfig
	feishuAdaptor IFeishu
	weworkAdaptor IWework
}

func (s *ScopeSyncer) Pre() (after func(), err error) {
	fmt.Println("pre one")
	sourceDeptTree, sourceUsers, err := s.feishuAdaptor.GetAllDeptUser(s.ctx, s.feishuConfig)
	if err != nil {
		return nil, err
	}
	targetDeptTree, targetUsers, err := s.weworkAdaptor.GetAllDeptUser(s.ctx, s.weworkConfig)
	if err != nil {
		return nil, err
	}

	s.deptSyncer.sourceDeptMap = do.NewFeishuDeptMap(sourceDeptTree)
	s.deptSyncer.targetDeptMap = do.NewWeworkDeptMap(targetDeptTree)
	sourceUserMap := make(map[interface{}]do.FeishuUser, len(sourceUsers))
	for _, v := range sourceUsers {
		sourceUserMap[v.UserID] = v
	}
	s.userSyncer.sourceUserMap = sourceUserMap
	targetUserMap := make(map[interface{}]do.WeworkUser, len(targetUsers))
	for _, v := range targetUsers {
		targetUserMap[v.UserID] = v
	}
	s.userSyncer.targetUserMap = targetUserMap
	return func() {
		fmt.Println("after one")
	}, nil
}

func (s *ScopeSyncer) GetSourceAllDeptUser() (depts []sync.IDeptInfo, users []sync.IUserInfo, err error) {
	root, err := s.deptSyncer.GetSourceDept(do.FeishuRootDeptID)
	if err != nil {
		return nil, nil, err
	}
	if err = sync.ForeachDeptTree(root.(*do.FeishuDeptTree), func(deptTree sync.IDeptTree) error {
		depts = append(depts, deptTree)
		return nil
	}); err != nil {
		return nil, nil, err
	}
	for _, user := range s.userSyncer.sourceUserMap {
		users = append(users, user)
	}
	return
}

func (s *ScopeSyncer) GetTargetDeptTree() (sync.IDeptTree, error) {
	dept, err := s.deptSyncer.GetTargetDept(do.WeworkRootDeptID)
	if err != nil {
		return nil, err
	}
	return dept.(*do.WeworkDeptTree), nil
}

func (s *ScopeSyncer) DeleteTargetDept(id interface{}) error {
	return s.deptSyncer.DeleteDept(id)
}
