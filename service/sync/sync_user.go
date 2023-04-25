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

	CreateUser(sourceUser IUserInfo) (interface{}, error)
	NeedUpdateUser(sourceUser IUserInfo, targetUser IUserInfo) bool
	UpdateUser(sourceUser IUserInfo, targetUser IUserInfo) error

	DeptSyncer() IDeptSync

	NeedSyncDelete() bool
	LeaveUser(targetUserID interface{}) error

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
	targetUser, err := c.userSyncer.GetTargetUser(targetID)
	if err != nil {
		return err
	}
	if !targetUser.IsExist() || targetUser.IsLeft() {
		return nil
	}
	if err = c.userSyncer.LeaveUser(targetUser.GetID()); err != nil {
		return err
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

	targetUser, err := c.userSyncer.GetTargetUser(targetID)
	if err != nil {
		return err
	}
	if targetUser.IsExist() {
		if targetUser.IsLeft() {
			return nil
		}
		if !c.userSyncer.NeedUpdateUser(sourceUser, targetUser) {
			return nil
		}
		if err = c.userSyncer.UpdateUser(sourceUser, targetUser); err != nil {
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

	targetUser, err := c.userSyncer.GetTargetUserByMobile(sourceUser.GetMobile())
	if err != nil {
		return err
	}
	if targetUser.IsExist() {
		if targetUser.IsLeft() {
			return nil
		}

		occupied, err := c.isIDOccupied(sourceUser, targetUser)
		if err != nil {
			return err
		}
		if occupied {
			return IDOccupied
		}

		c.userMapping.BindUser(sourceUser.GetID(), targetUser.GetID())
		if !c.userSyncer.NeedUpdateUser(sourceUser, targetUser) {
			return nil
		}
		if err = c.userSyncer.UpdateUser(sourceUser, targetUser); err != nil {
			return err
		}
		return nil
	}
	return errNotFound
}

func (c *UserSyncer) isIDOccupied(sourceUser IUserInfo, targetUser IUserInfo) (bool, error) {
	sourceID, exist, err := c.userMapping.GetSourceUserID(targetUser.GetID())
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
