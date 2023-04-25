package sync

type IDeptSync interface {
	// Pre 用于处理同步之前的逻辑，比如准备配置数据、对同步加锁等
	Pre() (after func(), err error)
	// IsNeedSync 判断来源部门id是否需要同步
	IsNeedSync(sourceDeptID interface{}) (bool, error)
	// GetSourceDept 获取来源数据的部门信息
	GetSourceDept(sourceDeptID interface{}) (IDeptInfo, error)
	GetTargetDept(targetDeptID interface{}) (IDeptInfo, error)
	// GetMatchedChild 在没有关联ID的情况下，通过父部门id和来源部门信息获取目的数据中匹配的部门
	GetMatchedChild(targetParentID interface{}, child IDeptInfo) (IDeptInfo, bool, error)
	// CreateDept 创建目的部门
	CreateDept(targetParentID interface{}, dept IDeptInfo) (id interface{}, err error)
	// NeedUpdate 部门是否需要更新
	NeedUpdate(targetParentID interface{}, targetDept IDeptInfo, sourceDept IDeptInfo) (bool, error)
	// UpdateDept 更新目的部门
	UpdateDept(targetDept IDeptInfo, sourceDept IDeptInfo) error
	// DeleteDept 删除目的部门
	DeleteDept(target interface{}) error
}

type DeptSyncer struct {
	deptSyncer IDeptSync
	mapping    DeptMapping
}

func NewDeptSyncer(dept IDeptSync, mapping DeptMapping) *DeptSyncer {
	return &DeptSyncer{dept, mapping}
}

func (c *DeptSyncer) SaveDept(deptID interface{}) (id interface{}, err error) {
	after, err := c.deptSyncer.Pre()
	if err != nil {
		return nil, err
	}
	defer after()

	isNeedSync, err := c.deptSyncer.IsNeedSync(deptID)
	if err != nil || !isNeedSync {
		return nil, err
	}

	return c.syncDept(deptID)
}

func (c *DeptSyncer) syncDept(deptID interface{}) (interface{}, error) {
	sourceDept, err := c.deptSyncer.GetSourceDept(deptID)
	if err != nil {
		return nil, err
	}
	// 递归同步父部门
	targetParentID, err := c.syncDept(sourceDept.GetParentID())
	if err != nil {
		return nil, err
	}
	// 在已有的映射中获取目的部门id
	targetDeptID, exist, err := c.mapping.GetTargetDeptID(deptID)
	if err != nil {
		return nil, err
	}
	if exist {
		// 获取目的部门
		targetDept, err := c.deptSyncer.GetTargetDept(targetDeptID)
		if err != nil {
			return nil, err
		}
		if targetDept.IsExist() {
			// 如果已存在目的部门，则直接更新
			needUpdate, err := c.deptSyncer.NeedUpdate(targetParentID, targetDept, sourceDept)
			if err != nil || !needUpdate {
				return targetDept.GetID(), err
			}
			if err = c.deptSyncer.UpdateDept(targetDept, sourceDept); err != nil {
				return nil, err
			}
			return targetDept.GetID(), nil
		}
	}

	// 根据父部门来获取匹配部门
	targetDept, exist, err := c.deptSyncer.GetMatchedChild(targetParentID, sourceDept)
	if exist {
		c.mapping.BindDept(targetDept.GetID(), sourceDept.GetID())
		needUpdate, err := c.deptSyncer.NeedUpdate(targetParentID, targetDept, sourceDept)
		if err != nil || !needUpdate {
			return nil, err
		}
		if err = c.deptSyncer.UpdateDept(targetDept, sourceDept); err != nil {
			return nil, err
		}
		return targetDept.GetID(), nil
	}
	// 已有部门中匹配不到，则直接创建
	targetID, err := c.deptSyncer.CreateDept(targetParentID, sourceDept)
	if err != nil {
		return nil, err
	}
	c.mapping.BindDept(targetID, sourceDept.GetID())
	return targetID, nil
}
