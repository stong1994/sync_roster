package do

type FeishuEmpStatus struct {
	IsFrozen    bool // 是否暂停
	IsResigned  bool // 是否离职
	IsActivated bool // 是否激活  和是否暂停是否冲突
	IsExited    bool // 是否主动退出，主动退出一段时间后用户会自动转为已离职
	IsUnjoin    bool // 是否未加入
}

type SexType int

const (
	SexTypeUnknown SexType = iota
	SexTypeMale
	SexTypeFemale
)

type FeishuUser struct {
	UserID       string
	Name         string
	Alias        string
	Sex          SexType
	SelfPhone    string
	WorkMail     string
	Position     string // 职务信息
	Address      string // 地址
	Status       FeishuEmpStatus
	Department   []string
	LeaderUserID string
	City         string
	EmpNo        string
	EnName       string
	JoinTime     int
}

func (f FeishuUser) GetID() interface{} {
	return f.UserID
}

func (f FeishuUser) GetDeptID() []interface{} {
	rst := make([]interface{}, len(f.Department))
	for i, v := range f.Department {
		rst[i] = v
	}
	return rst
}

func (f FeishuUser) GetMobile() string {
	return f.SelfPhone
}

func (f FeishuUser) IsLeft() bool {
	return f.Status.IsExited || f.Status.IsResigned
}

func (f FeishuUser) IsExist() bool {
	return f.UserID != ""
}
