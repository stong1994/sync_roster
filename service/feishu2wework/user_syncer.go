package feishu2wework

import (
	"context"
	"fmt"
	"sync_roster/service/do"
	"sync_roster/service/sync"
)

type UserSyncer struct {
	ctx           context.Context
	sourceUserMap map[interface{}]do.FeishuUser
	targetUserMap map[interface{}]do.WeworkUser
	deptSyncer    sync.IDeptSync
	userMapping   sync.UserMapping
	deptMapping   sync.DeptMapping

	feishuConfig  do.FeishuConfig
	weworkConfig  do.WeworkConfig
	feishuAdaptor IFeishu
	weworkAdaptor IWework
}

func (u *UserSyncer) Pre() (after func(), err error) {
	fmt.Println("lock there")
	return func() {}, nil
}

func (u *UserSyncer) After(sourceUser sync.IUserInfo) error {
	fmt.Println("synced user: ", sourceUser.GetID())
	return nil
}

func (u *UserSyncer) IsNeedSync(sourceUserID interface{}) (bool, error) {
	user, err := u.GetSourceUser(sourceUserID)
	if err != nil {
		return false, err
	}
	if !user.IsExist() || user.IsLeft() {
		return false, nil
	}
	return true, nil
}

func (u *UserSyncer) GetSourceUser(sourceUserID interface{}) (sync.IUserInfo, error) {
	if user, ok := u.sourceUserMap[sourceUserID]; ok {
		return user, nil
	}
	user, err := u.feishuAdaptor.GetUser(u.ctx, u.feishuConfig, sourceUserID)
	if err != nil {
		return nil, err
	}
	u.sourceUserMap[sourceUserID] = user
	return user, nil
}

func (u *UserSyncer) GetTargetUser(targetUserID interface{}) (sync.IUserInfo, error) {
	if user, ok := u.targetUserMap[targetUserID]; ok {
		return user, nil
	}
	user, err := u.weworkAdaptor.GetUser(u.ctx, u.weworkConfig, targetUserID)
	if err != nil {
		return nil, err
	}
	u.targetUserMap[targetUserID] = user
	return user, nil
}

func (u *UserSyncer) GetTargetUserByMobile(mobile string) (sync.IUserInfo, error) {
	for _, v := range u.targetUserMap {
		if v.GetMobile() == mobile {
			return v, nil
		}
	}
	user, err := u.weworkAdaptor.GetUsrByMobile(u.ctx, u.weworkConfig, mobile)
	if err != nil {
		return nil, err
	}
	u.targetUserMap[user.GetID()] = user
	return user, nil
}

func (u *UserSyncer) GetNeedSyncDeptList(sourceUserID interface{}) ([]interface{}, error) {
	user, err := u.GetSourceUser(sourceUserID)
	if err != nil {
		return nil, err
	}
	return user.GetDeptID(), nil
}

func (u *UserSyncer) CreateUser(user sync.IUserInfo) (interface{}, error) {
	createInfo, err := u.getCreateUserInfo(user.(do.FeishuUser))
	if err != nil {
		return nil, err
	}
	weworkUser, err := u.weworkAdaptor.CrateUser(u.ctx, u.weworkConfig, createInfo)
	if err != nil {
		return nil, err
	}
	u.targetUserMap[weworkUser.GetID()] = weworkUser
	return weworkUser.GetID(), nil
}

func (u *UserSyncer) getCreateUserInfo(feishuUser do.FeishuUser) (rst do.WeworkUser, err error) {
	var weworkDetps []int
	for _, v := range feishuUser.Department {
		weworkDept, exist, err := u.deptMapping.GetTargetDeptID(v)
		if err != nil {
			return rst, err
		}
		if !exist {
			continue
		}
		weworkDetps = append(weworkDetps, weworkDept.(int))
	}
	// 主部门使用第一个
	if len(weworkDetps) > 0 {
		rst.MainDept = weworkDetps[0]
	}
	rst.Department = weworkDetps
	rst.Name = feishuUser.Name
	rst.Address = feishuUser.Address
	rst.Position = feishuUser.Position
	rst.Sex = feishuUser.Sex
	rst.SelfPhone = feishuUser.SelfPhone
	rst.WorkMail = feishuUser.WorkMail
	rst.Status = do.WeworkEmpStatusActive
	return
}

func (u *UserSyncer) NeedUpdateUser(sourceUser sync.IUserInfo, targetUser sync.IUserInfo) bool {
	fUser, wUser := sourceUser.(do.FeishuUser), targetUser.(do.WeworkUser)
	if fUser.Name != wUser.Name || fUser.Position != wUser.Position || fUser.Sex != wUser.Sex ||
		fUser.Address != wUser.Address || fUser.WorkMail != wUser.WorkMail {
		return true
	}
	if len(fUser.Department) != len(wUser.Department) {
		return true
	}
	wDepts := make(map[int]bool)
	for _, v := range wUser.Department {
		wDepts[v] = true
	}
	for _, v := range fUser.Department {
		id, exist, _ := u.deptMapping.GetTargetDeptID(v)
		if !exist || !wDepts[id.(int)] {
			return true
		}
	}
	return false
}

func (u *UserSyncer) UpdateUser(sourceUser sync.IUserInfo, targetUser sync.IUserInfo) error {
	fUser, wUser := sourceUser.(do.FeishuUser), targetUser.(do.WeworkUser)
	wUser.Name = fUser.Name
	wUser.Position = fUser.Position
	wUser.Sex = fUser.Sex
	wUser.WorkMail = fUser.WorkMail
	wUser.Address = fUser.Address
	wDepts := make(map[int]bool)
	for _, v := range wUser.Department {
		wDepts[v] = true
	}
	for _, v := range fUser.Department {
		id, exist, _ := u.deptMapping.GetTargetDeptID(v)
		if exist && !wDepts[id.(int)] {
			wUser.Department = append(wUser.Department, id.(int))
		}
	}
	// todo handle main dept
	if err := u.weworkAdaptor.UpdateUser(u.ctx, u.weworkConfig, wUser); err != nil {
		return err
	}
	u.targetUserMap[wUser.UserID] = wUser
	return nil
}

func (u *UserSyncer) DeptSyncer() sync.IDeptSync {
	return u.deptSyncer
}

func (u *UserSyncer) NeedSyncDelete() bool {
	return true
}

func (u *UserSyncer) LeaveUser(targetUserID interface{}) error {
	if err := u.weworkAdaptor.LeaveUser(u.ctx, u.weworkConfig, targetUserID.(string)); err != nil {
		return err
	}
	return nil
}

func (u *UserSyncer) NeedSyncLeader(source, target sync.IDeptInfo) bool {
	fDept, wDept := source.(do.FeishuDeptInfo), target.(do.WeworkDeptInfo)
	wUser, exist, _ := u.userMapping.GetTargetUserID(fDept.LeaderID)
	if !exist {
		return false
	}
	for _, v := range wDept.Leaders {
		if v == wUser {
			return false
		}
	}
	return true
}

func (u *UserSyncer) SyncLeader(source, target sync.IDeptInfo) error {
	//fDept, wDept := source.(do.FeishuDeptInfo), target.(do.WeworkDeptInfo)
	//wUser, _, _ := u.userMapping.GetTargetUserID(fDept.LeaderID)
	//wDept.Leaders = append(wDept.Leaders, wUser.(string))
	//if err := u.weworkAdaptor.UpdateDept(u.ctx, u.weworkConfig, wDept); err != nil {
	//	return err
	//}
	//u.DeptSyncer().update
	// todo
	return nil
}
