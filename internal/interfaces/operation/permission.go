// Package operation
package operation

type Permission int64

// 权限节点上限是64, 超过64需要使用切片
const (
	AdminEntry Permission = 1 << iota
	ControllerEntry
	UserShowList
	UserGetProfile
	UserAdd
	UserEditBaseInfo
	UserEditPermission
	UserEditRating
	ActivityPublish
	ActivityShowList
	ActivityEditContent
	ActivityEditFacility
	ActivityEditState
	ActivityEditPilotState
	ActivityDelete
	ClientSendMessage
	ClientKill
)

var PermissionMap = map[string]Permission{
	"AdminEntry":             AdminEntry,
	"ControllerEntry":        ControllerEntry,
	"UserShowList":           UserShowList,
	"UserGetProfile":         UserGetProfile,
	"UserAdd":                UserAdd,
	"UserEditBaseInfo":       UserEditBaseInfo,
	"UserEditPermission":     UserEditPermission,
	"UserEditRating":         UserEditRating,
	"ActivityPublish":        ActivityPublish,
	"ActivityShowList":       ActivityShowList,
	"ActivityEditContent":    ActivityEditContent,
	"ActivityEditFacility":   ActivityEditFacility,
	"ActivityEditState":      ActivityEditState,
	"ActivityEditPilotState": ActivityEditPilotState,
	"ActivityDelete":         ActivityDelete,
	"ClientSendMessage":      ClientSendMessage,
	"ClientKill":             ClientKill,
}

func (p *Permission) IsValid() bool {
	maxPerm := ClientKill<<1 - 1 // 计算最大有效位
	return *p >= 0 && *p <= maxPerm
}

func (p *Permission) HasPermission(perm Permission) bool {
	return *p&perm != 0
}

func (p *Permission) Grant(perm Permission) {
	*p |= perm
}

func (p *Permission) Revoke(perm Permission) {
	*p &^= perm
}
