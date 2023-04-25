package do

type WeworkEmpStatus int // 激活状态: 1=已激活，2=已禁用，4=未激活，5=退出企业

const (
	WeworkEmpStatusUnknown = iota
	WeworkEmpStatusActive
	WeworkEmpStatusForbidden
	WeworkEmpStatusUnActive = 4
	WeworkEmpStatusQuit     = 5
)

const (
	NotDeptLeader = 0
	IsDeptLeader  = 1
)

type WeworkUser struct {
	UserID      string
	MainDept    int
	Name        string
	Alias       string
	Sex         SexType
	SelfPhone   string
	WorkPhone   string
	WorkMail    string
	Position    string // 职务信息
	Address     string // 地址
	Status      WeworkEmpStatus
	Department  []int
	LeaderDepts []int
	OpenUserID  string
}

func (we WeworkUser) GetID() interface{} {
	return we.UserID
}

func (we WeworkUser) GetDeptID() []interface{} {
	return []interface{}{we.MainDept}
}

func (we WeworkUser) GetLeaderMap() map[int]bool {
	rst := make(map[int]bool)
	for i, v := range we.Department {
		is := false
		if len(we.LeaderDepts) > i {
			is = we.LeaderDepts[i] == IsDeptLeader
		}
		rst[v] = is
	}
	return rst
}

func (we WeworkUser) GetMobile() string {
	return we.SelfPhone
}

func (we WeworkUser) IsNeedSync() bool {
	if we.Status == WeworkEmpStatusForbidden || we.Status == WeworkEmpStatusQuit {
		return false
	}
	return true
}

func (we WeworkUser) IsLeft() bool {
	return we.Status == WeworkEmpStatusQuit
}

func (we WeworkUser) IsExist() bool {
	return we.UserID != ""
}
