package do

import (
	"sync_roster/service/sync"
)

const FeishuRootDeptID = "0"

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

func (f FeishuDeptTree) GetChildren() []sync.IDeptTree {
	rst := make([]sync.IDeptTree, len(f.List))
	for i, v := range f.List {
		rst[i] = v
	}
	return rst
}

type FeishuDeptMap struct {
	data map[string]*FeishuDeptTree
}

func NewFeishuDeptMap(root *FeishuDeptTree) *FeishuDeptMap {
	data := make(map[string]*FeishuDeptTree)
	var dfs func(dept *FeishuDeptTree)
	dfs = func(dept *FeishuDeptTree) {
		data[dept.ID] = dept
		for _, child := range dept.List {
			dfs(child)
		}
	}
	return &FeishuDeptMap{data: data}
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
