package do

import (
	"sync_roster/service/sync"
)

const WeworkRootDeptID = 1

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

func (w WeworkDeptInfo) ToTree() *WeworkDeptTree {
	return &WeworkDeptTree{
		WeworkDeptInfo: WeworkDeptInfo{
			ID:       w.ID,
			Name:     w.Name,
			Code:     w.Code,
			Leaders:  w.Leaders,
			ParentID: w.ParentID,
		},
		List: nil,
	}
}

type WeworkDeptTree struct {
	WeworkDeptInfo
	List []*WeworkDeptTree
}

func (w WeworkDeptTree) GetChildren() []sync.IDeptTree {
	rst := make([]sync.IDeptTree, len(w.List))
	for i, v := range w.List {
		rst[i] = v
	}
	return rst
}

type WeworkDeptMap struct {
	data map[int]*WeworkDeptTree
}

func NewWeworkDeptMap(root *WeworkDeptTree) *WeworkDeptMap {
	data := make(map[int]*WeworkDeptTree)
	var dfs func(dept *WeworkDeptTree)
	dfs = func(dept *WeworkDeptTree) {
		data[dept.ID] = dept
		for _, child := range dept.List {
			dfs(child)
		}
	}
	return &WeworkDeptMap{data: data}
}

func (w *WeworkDeptMap) Get(id int) *WeworkDeptTree {
	return w.data[id]
}

func (w *WeworkDeptMap) Add(dept *WeworkDeptTree) {
	w.data[dept.ID] = dept
	if p, ok := w.data[dept.ParentID]; ok {
		for _, v := range p.List {
			if v.ID == dept.ID {
				return
			}
		}
		p.List = append(p.List, dept)
	}
}

func (w *WeworkDeptMap) Update(info WeworkDeptUpdateInfo) {
	dept, ok := w.data[info.ID]
	if !ok {
		panic("dept map must have the updating dept")
	}
	dept.Name = info.Name

	if dept.ParentID == info.ParentID {
		dept.Name = info.Name
		return
	}
	if p, ok := w.data[dept.ParentID]; ok {
		w.removeSub(p, info.ID)
	}
	dept.ParentID = info.ParentID
	if p, ok := w.data[info.ParentID]; ok {
		p.List = append(p.List, dept)
	}
}

func (w *WeworkDeptMap) Delete(id int) {
	dept, ok := w.data[id]
	if !ok {
		panic("dept map must have the deleting dept")
	}
	delete(w.data, id)
	if p, ok := w.data[dept.ParentID]; ok {
		w.removeSub(p, id)
	}
}

func (w *WeworkDeptMap) removeSub(p *WeworkDeptTree, subID int) {
	for i, v := range p.List {
		if v.ID == subID {
			p.List = append(p.List[:i], p.List[i+1:]...)
		}
	}
}

type WeworkDeptCreateInfo struct {
	Name     string
	ParentID int
}

type WeworkDeptUpdateInfo struct {
	ID       int
	Name     string
	ParentID int
}
