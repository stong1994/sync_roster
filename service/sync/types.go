package sync

type IDeptInfo interface {
	GetID() interface{}
	GetParentID() interface{}
	GetName() string
	IsExist() bool
}

var NotExistDept notExistDept

type notExistDept struct{}

func (n notExistDept) GetID() interface{} {
	return ""
}

func (n notExistDept) GetParentID() interface{} {
	return ""
}

func (n notExistDept) GetName() string {
	return ""
}

func (n notExistDept) IsExist() bool {
	return false
}

type IDeptTree interface {
	IDeptInfo
	GetChildren() []IDeptTree
}

func ForeachDeptTree(tree IDeptTree, f func(deptTree IDeptTree) error) error {
	f(tree)
	childs := tree.GetChildren()
	for i := range childs {
		if err := ForeachDeptTree(childs[i], f); err != nil {
			return err
		}
	}
	return nil
}

func ReverseForeachDeptTree(tree IDeptTree, f func(deptTree IDeptTree) error) error {
	childs := tree.GetChildren()
	for i := range childs {
		if err := ReverseForeachDeptTree(childs[i], f); err != nil {
			return err
		}
	}
	return f(tree)
}

type IUserInfo interface {
	GetID() interface{}
	GetDeptID() []interface{}
	GetMobile() string
	IsLeft() bool // 是否离职
	IsExist() bool
}

var NotExistUser notExistUser

type notExistUser struct{}

func (n notExistUser) GetID() interface{} {
	return ""
}

func (n notExistUser) GetUserNo() string {
	return ""
}

func (n notExistUser) IsLeft() bool {
	return false
}

func (n notExistUser) IsExist() bool {
	return false
}
func (n notExistUser) GetMobile() string {
	return ""
}
func (n notExistUser) GetDeptID() interface{} {
	return ""
}

func (n notExistUser) GetName() string {
	return ""
}

type DeptSyncErr struct {
	ID  interface{}
	Err error
}

func NewDeptSyncErr(id interface{}, err error) DeptSyncErr {
	return DeptSyncErr{
		ID:  id,
		Err: err,
	}
}

type DeptMapping interface {
	// GetTargetDeptID 获取已有的关联映射
	GetTargetDeptID(sourceID interface{}) (targetID interface{}, exist bool, err error)
	// BindDept 绑定部门映射
	BindDept(targetID, sourceID interface{}) error
}

type UserMapping interface {
	GetTargetUserID(sourceID interface{}) (targetID interface{}, exist bool, err error)
	GetSourceUserID(targetID interface{}) (sourceID interface{}, exist bool, err error)
	BindUser(targetID, sourceID interface{}) error
}
