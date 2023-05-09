package feishu2wework

import (
	"context"
	"fmt"
	"sync_roster/service/do"
	"sync_roster/service/sync"
)

type DeptSyncer struct {
	ctx           context.Context
	sourceDeptMap *do.FeishuDeptMap
	targetDeptMap *do.WeworkDeptMap
	deptMapping   sync.DeptMapping

	feishuConfig  do.FeishuConfig
	weworkConfig  do.WeworkConfig
	feishuAdaptor IFeishu
	weworkAdaptor IWework
}

func (d *DeptSyncer) Pre() (after func(), err error) {
	fmt.Println("lock there2")
	return func() {
		fmt.Println("unlock there2")
	}, nil
}

func (d *DeptSyncer) IsNeedSync(sourceDeptID interface{}) (bool, error) {
	return true, nil
}

func (d *DeptSyncer) GetSourceDept(sourceDeptID interface{}) (sync.IDeptInfo, error) {
	if dept := d.sourceDeptMap.Get(sourceDeptID.(string)); dept != nil {
		return dept, nil
	}
	dept, err := d.feishuAdaptor.GetDept(d.ctx, d.feishuConfig, sourceDeptID.(string))
	if err != nil {
		return nil, err
	}
	d.sourceDeptMap.Add(dept)
	return dept, nil
}

func (d *DeptSyncer) GetTargetDept(targetDeptID interface{}) (sync.IDeptInfo, error) {
	if dept := d.targetDeptMap.Get(targetDeptID.(int)); dept != nil {
		return dept, nil
	}
	dept, err := d.weworkAdaptor.GetDept(d.ctx, d.weworkConfig, targetDeptID.(int))
	if err != nil {
		return nil, err
	}
	d.targetDeptMap.Add(dept)
	return dept, nil
}

func (d *DeptSyncer) GetMatchedChild(targetParentID interface{}, child sync.IDeptInfo) (sync.IDeptInfo, bool, error) {
	dept, err := d.GetTargetDept(targetParentID)
	if err != nil {
		return nil, false, err
	}
	for _, c := range dept.(*do.WeworkDeptTree).List {
		if c.Name == child.GetName() {
			return c, true, nil
		}
	}
	return sync.NotExistDept, false, nil

}

func (d *DeptSyncer) CreateDept(targetParentID interface{}, dept sync.IDeptInfo) (id interface{}, err error) {
	deptInfo, err := d.weworkAdaptor.CrateDept(d.ctx, d.weworkConfig, do.WeworkDeptCreateInfo{
		Name:     dept.GetName(),
		ParentID: targetParentID.(int),
	})
	if err != nil {
		return nil, err
	}
	d.targetDeptMap.Add(deptInfo.ToTree())
	return deptInfo.ID, nil
}

func (d *DeptSyncer) NeedUpdate(targetParentID interface{}, targetDept sync.IDeptInfo, sourceDept sync.IDeptInfo) (bool, error) {
	if targetDept.GetName() != sourceDept.GetName() {
		return true, nil
	}
	targetParent, err := d.GetTargetDept(targetDept.GetParentID())
	if err != nil {
		return false, err
	}
	if !targetParent.IsExist() {
		return true, nil
	}
	targetParentInMapping, exist, err := d.deptMapping.GetTargetDeptID(sourceDept.GetParentID())
	if err != nil {
		return false, err
	}
	if !exist {
		panic("must have parent dept mapping")
	}
	if targetParent.GetID() != targetParentInMapping {
		return true, nil
	}
	return false, nil
}

func (d *DeptSyncer) UpdateDept(targetDept sync.IDeptInfo, sourceDept sync.IDeptInfo) error {
	targetParentInMapping, exist, err := d.deptMapping.GetTargetDeptID(sourceDept.GetParentID())
	if err != nil {
		return err
	}
	if !exist {
		panic("must have parent dept mapping")
	}
	updateInfo := do.WeworkDeptUpdateInfo{
		Name:     targetDept.GetName(),
		ParentID: targetParentInMapping.(int),
	}
	if err = d.weworkAdaptor.UpdateDept(d.ctx, d.weworkConfig, updateInfo); err != nil {
		return err
	}
	d.targetDeptMap.Update(updateInfo)
	return nil
}

func (d *DeptSyncer) DeleteDept(target interface{}) error {
	if err := d.weworkAdaptor.DeleteDept(d.ctx, d.weworkConfig, target.(int)); err != nil {
		return err
	}
	d.targetDeptMap.Delete(target.(int))
	return nil
}
