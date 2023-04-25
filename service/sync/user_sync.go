package sync

import (
	"errors"
	"fmt"
)

type IUserSync interface {
	Pre() (after func(), err error)
	After(sourceUser IUserInfo) error

	IsNeedSync(sourceUserID interface{}) (bool, error)
	GetSourceUser(sourceUserID interface{}) (IUserInfo, error)
	GetTargetUser(targetUserID interface{}) (IUserInfo, error)
	GetTargetUserByMobile(mobile string) (IUserInfo, error)
	GetNeedSyncDeptList(sourceUserID interface{}) ([]interface{}, error)

	CreateUser(user IUserInfo) (interface{}, error)
	NeedUpdateUser(sourceUser IUserInfo, thirdUser IUserInfo) bool
	UpdateUser(sourceUser IUserInfo, thirdUser IUserInfo) error

	DeptSyncer() IDeptSync

	NeedSyncDelete() bool
	LeaveUser(userID interface{}) error
	IsLeaveNow(userID interface{}) bool
	DelayLeave(userID interface{}) error

	NeedSyncLeader(source, target IDeptInfo) bool
	SyncLeader(source, target IDeptInfo) error
}

type UserSyncer struct {
	userSyncer  IUserSync
	deptSyncer  IDeptSync
	userMapping UserMapping
	deptMapping DeptMapping
}

func NewUserSyncer(syncer IUserSync, deptSyncer IDeptSync, userMapping UserMapping, deptMapping DeptMapping) *UserSyncer {
	return &UserSyncer{
		userSyncer:  syncer,
		deptSyncer:  deptSyncer,
		userMapping: userMapping,
		deptMapping: deptMapping,
	}
}

func (c *UserSyncer) SaveUsers(userIDs []interface{}) error {
	after, err := c.userSyncer.Pre()
	if err != nil {
		return err
	}
	defer after()

	for _, userID := range userIDs {
		if err = c.saveUser(userID); err != nil {
			return err
		}
	}
	return nil
}

func (c *UserSyncer) saveUser(userID interface{}) error {
	isNeedSync, err := c.userSyncer.IsNeedSync(userID)
	if err != nil || !isNeedSync {
		return err
	}

	sourceUser, err := c.userSyncer.GetSourceUser(userID)
	if err != nil {
		return err
	}
	if !sourceUser.IsExist() {
		return fmt.Errorf("not exist source user")
	}
	defer c.userSyncer.After(sourceUser)

	if sourceUser.IsLeft() {
		return c.leaveUser(sourceUser)
	}

	_, exist, err := c.userMapping.GetTargetUserID(sourceUser.GetID())
	if err != nil {
		return err
	}
	if exist {
		err = c.syncByID(sourceUser)
		if err != nil {
			if !errors.Is(err, errNotFound) {
				return err
			}
		} else {
			return nil
		}
	}
	if sourceUser.GetMobile() != "" {
		err = c.syncByMobile(sourceUser)
		if err != nil {
			if !errors.Is(err, errNotFound) {
				return err
			}
		} else {
			return nil
		}
	}
	_, err = c.userSyncer.CreateUser(sourceUser)
	return err
}

func (c *UserSyncer) leaveUser(sourceUser IUserInfo) error {
	targetID, exist, err := c.userMapping.GetTargetUserID(sourceUser.GetID())
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}

	if !c.userSyncer.NeedSyncDelete() {
		return nil
	}
	thirdUser, err := c.userSyncer.GetTargetUser(targetID)
	if err != nil {
		return err
	}
	if !thirdUser.IsExist() || thirdUser.IsLeft() {
		return nil
	}
	if c.userSyncer.IsLeaveNow(sourceUser.GetID()) {
		if err = c.userSyncer.LeaveUser(sourceUser.GetID()); err != nil {
			return err
		}
	} else {
		c.userSyncer.DelayLeave(sourceUser.GetID())
	}
	return nil
}

func (c *UserSyncer) syncByID(sourceUser IUserInfo) error {
	userDepts, err := c.userSyncer.GetNeedSyncDeptList(sourceUser.GetID())
	if err != nil {
		return err
	}
	for _, deptID := range userDepts {
		_, err = NewDeptSyncer(c.deptSyncer, c.deptMapping).SaveDept(deptID)
		if err != nil {
			return err
		}
	}
	targetID, exist, err := c.userMapping.GetTargetUserID(sourceUser.GetID())
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}

	thirdUser, err := c.userSyncer.GetTargetUser(targetID)
	if err != nil {
		return err
	}
	if thirdUser.IsExist() {
		if thirdUser.IsLeft() {
			return nil
		}
		if !c.userSyncer.NeedUpdateUser(sourceUser, thirdUser) {
			return nil
		}
		if err = c.userSyncer.UpdateUser(sourceUser, thirdUser); err != nil {
			return err
		}
		return nil
	}
	return errNotFound
}

func (c *UserSyncer) syncByMobile(sourceUser IUserInfo) error {
	userDepts, err := c.userSyncer.GetNeedSyncDeptList(sourceUser.GetID())
	if err != nil {
		return err
	}
	for _, deptID := range userDepts {
		_, err = NewDeptSyncer(c.deptSyncer, c.deptMapping).SaveDept(deptID)
		if err != nil {
			return err
		}
	}

	thirdUser, err := c.userSyncer.GetTargetUserByMobile(sourceUser.GetMobile())
	if err != nil {
		return err
	}
	if thirdUser.IsExist() {
		if thirdUser.IsLeft() {
			return nil
		}

		occupied, err := c.isIDOccupied(sourceUser, thirdUser)
		if err != nil {
			return err
		}
		if occupied {
			return IDOccupied
		}

		c.userMapping.BindUser(sourceUser.GetID(), thirdUser.GetID())
		if !c.userSyncer.NeedUpdateUser(sourceUser, thirdUser) {
			return nil
		}
		if err = c.userSyncer.UpdateUser(sourceUser, thirdUser); err != nil {
			return err
		}
		return nil
	}
	return errNotFound
}

func (c *UserSyncer) isIDOccupied(sourceUser IUserInfo, thirdUser IUserInfo) (bool, error) {
	sourceID, exist, err := c.userMapping.GetSourceUserID(thirdUser.GetID())
	if err != nil {
		return false, err
	}
	if exist {
		bindUser, err := c.userSyncer.GetSourceUser(sourceID)
		if err != nil {
			return false, err
		}
		if bindUser.IsExist() && bindUser.GetID() != sourceUser.GetID() {
			return true, nil
		}
	}
	return false, nil
}
