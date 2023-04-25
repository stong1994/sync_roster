package do

type FeishuDeptInfo struct {
	ID          string
	Name        string
	FullName    string
	ParentID    string
	LeaderID    string
	IsUnderRoot bool
}

func (f FeishuDeptInfo) GetID() interface{} {
	return f.ID
}

func (f FeishuDeptInfo) GetParentID() interface{} {
	return f.ParentID
}

func (f FeishuDeptInfo) GetName() string {
	return f.Name
}

func (f FeishuDeptInfo) IsExist() bool {
	return f.ID != ""
}

type FeishuDeptTree struct {
	ID          string
	Name        string
	FullName    string
	ParentID    string
	LeaderID    string
	IsUnderRoot bool
	List        []*FeishuDeptTree
}
