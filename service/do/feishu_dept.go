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
	FeishuDeptInfo
	List []*FeishuDeptTree
}

type FeishuDeptMap struct {
	data map[string]*FeishuDeptTree
}

func (f *FeishuDeptMap) Get(id string) *FeishuDeptTree {
	return f.data[id]
}

func (f *FeishuDeptMap) Add(dept *FeishuDeptTree) {
	f.data[dept.ID] = dept
	if p, ok := f.data[dept.ParentID]; ok {
		for _, v := range p.List {
			if v.ID == dept.ID {
				return
			}
		}
		p.List = append(p.List, dept)
	}
}
