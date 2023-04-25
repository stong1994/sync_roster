package do

type WeworkDeptInfo struct {
	ID       int
	Name     string
	Code     string
	Leaders  []string
	ParentID int
}

func (w WeworkDeptInfo) GetID() interface{} {
	return w.ID
}

func (w WeworkDeptInfo) GetParentID() interface{} {
	return w.ParentID
}

func (w WeworkDeptInfo) GetName() string {
	return w.Name
}

func (w WeworkDeptInfo) IsExist() bool {
	return w.ID != 0
}

type WeworkDeptTree struct {
	ID       int
	Name     string
	Code     string
	ParentID int
	List     []WeworkDeptTree
}
