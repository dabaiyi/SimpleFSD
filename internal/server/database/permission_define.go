// Package permission
package database

type Permission int64

// 权限节点上限是64, 超过64需要使用切片
const (
	AdminEntry Permission = 1 << iota
	UserShow
	UserAdd
	UserEditEmail
	UserEditUsername
	UserEditPassword
	UserEditPermission
	UserEditRating
	ActivityPublish
	ActivityEditContent
	ActivityEditFacility
	ActivityEditState
	ActivityDelete
)

func (p *Permission) IsValid() bool {
	maxPerm := ActivityDelete<<1 - 1 // 计算最大有效位
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
