package sync

import (
	"errors"
)

type IScopeSync interface {
	Pre() (after func(), err error)
	GetSourceAllDeptUser() ([]IDeptInfo, []IUserInfo, error)
	GetTargetDeptTree() (IDeptTree, error)
	DeleteTargetDept(id interface{}) error
}

type ScopeSyncer struct {
	scopeSyncer IScopeSync
	deptSyncer  IDeptSync
	userSyncer  IUserSync
	deptMapping DeptMapping
	userMapping UserMapping
	deptErrors  []DeptSyncErr
}

func NewScopeSyncer(syncer IScopeSync, deptSyncer IDeptSync,
	userSyncer IUserSync, deptMapping DeptMapping, userMapping UserMapping) *ScopeSyncer {
	return &ScopeSyncer{
		scopeSyncer: syncer,
		deptSyncer:  deptSyncer,
		userSyncer:  userSyncer,
		deptMapping: deptMapping,
		userMapping: userMapping,
	}
}

func (t *ScopeSyncer) Do() error {
	after, err := t.scopeSyncer.Pre()
	if err != nil {
		return err
	}

	defer after()

	sourceDepts, ehrUsers, err := t.scopeSyncer.GetSourceAllDeptUser()
	if err != nil {
		return err
	}
	var deptNameOccupied []func() (interface{}, error)

	deptSyncer := NewDeptSyncer(t.deptSyncer, t.deptMapping)
	deptMap := make(map[interface{}]struct{}, len(sourceDepts))
	for _, dept := range sourceDepts {
		id, err := deptSyncer.SaveDept(dept.GetID())
		deptMap[id] = struct{}{} // 发生错误也记录id，这是因为在更新时，即使发生错误也将id返回来
		if err != nil {
			if errors.Is(err, ErrDeptNameOccupied) {
				deptNameOccupied = append(deptNameOccupied, func() (interface{}, error) {
					id, err := deptSyncer.SaveDept(dept.GetID())
					return id, err
				})
			} else {
				t.deptErrors = append(t.deptErrors, NewDeptSyncErr(id, err))
			}
		}
	}
	userMap := make(map[any]struct{}, len(ehrUsers))

	userSyncer := NewUserSyncer(t.userSyncer, t.deptSyncer, t.userMapping, t.deptMapping)
	var waitSyncedUser []IUserInfo
	for _, user := range ehrUsers {
		userMap[user.GetID()] = struct{}{}
		if user.IsLeft() {
			err = userSyncer.leaveUser(user)
			if err != nil {
				return err
			}
			continue
		}
		_, exist, err := t.userMapping.GetTargetUserID(user.GetID())
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
		if !exist {
			waitSyncedUser = append(waitSyncedUser, user)
			continue
		}
		err = userSyncer.syncByID(user)
		if err == nil {
			continue
		}
		if errors.Is(err, errNotFound) {
			waitSyncedUser = append(waitSyncedUser, user)
			continue
		}
		return err
	}
	var waitCreateUser []IUserInfo
	for _, user := range waitSyncedUser {
		if user.GetMobile() == "" {
			waitCreateUser = append(waitCreateUser, user)
			continue
		}
		err = userSyncer.syncByMobile(user)
		if err == nil {
			continue
		}
		if errors.Is(err, errNotFound) {
			waitCreateUser = append(waitCreateUser, user)
			continue
		}
		return err
	}

	for _, user := range waitCreateUser {
		if id, err := t.userSyncer.CreateUser(user); err != nil {
			return err
		} else {
			t.userMapping.BindUser(user.GetID(), id)
		}
	}

	// 同步部门leader
	for _, v := range sourceDepts {
		sourceDept, err := t.deptSyncer.GetSourceDept(v.GetID())
		if err != nil {
			return err
		}
		targetDeptID, exist, err := t.deptMapping.GetTargetDeptID(sourceDept.GetID())
		if err != nil {
			return err
		}
		if !exist {
			continue
		}
		targetDept, err := t.deptSyncer.GetTargetDept(targetDeptID)
		if err != nil {
			return err
		}

		if t.userSyncer.NeedSyncLeader(sourceDept, targetDept) {
			if err = t.userSyncer.SyncLeader(sourceDept, targetDept); err != nil {
				return err
			}
		}
	}

	targetDeptTree, err := t.scopeSyncer.GetTargetDeptTree()
	if err != nil {
		return err
	}
	if err = ReverseForeachDeptTree(targetDeptTree, func(deptTree IDeptTree) error {
		if _, ok := deptMap[deptTree.GetID()]; !ok {
			if err = t.scopeSyncer.DeleteTargetDept(deptTree.GetID()); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	// 在删除部门后，对更新时发生名称冲突的部门再次更新
	for _, f := range deptNameOccupied {
		if id, err := f(); err != nil {
			t.deptErrors = append(t.deptErrors, NewDeptSyncErr(id, err))
		}
	}
	return nil
}
