// Package operation
package operation

import "time"

type ActivityStatus int

const (
	Open     ActivityStatus = iota // 报名中
	InActive                       // 活动中
	Closed                         // 已结束
)

type ActivityPilotStatus int

const (
	Signed    ActivityPilotStatus = iota // 已报名
	Clearance                            // 已放行
	Takeoff                              // 已起飞
	Landing                              // 已落地
)

// ActivityOperationInterface 联飞活动操作接口定义
type ActivityOperationInterface interface {
	// NewActivity 创建新活动
	NewActivity(user *User, title string, imageUrl string, activeTime time.Time, dep string, arr string, route string, distance int, notams string) (activity *Activity)
	// NewActivityFacility 创建新活动管制席位
	NewActivityFacility(activity *Activity, rating int, callsign string, frequency float64) (activityFacility *ActivityFacility)
	// NewActivityAtc 创建新参加活动的管制员
	NewActivityAtc(activity *Activity, facility *ActivityFacility) (activityAtc *ActivityATC)
	// NewActivityPilot 创建新参加活动的飞行员
	NewActivityPilot(activity *Activity, user *User, callsign string, aircraftType string) (activityPilot *ActivityPilot)
	// GetActivities 获取指定日期内的所有活动, 当err为nil时返回值activities有效
	GetActivities(startDay, endDay time.Time) (activities []*Activity, err error)
	// GetActivityById 通过活动Id获取活动详细内容, 当err为nil时返回值activity有效
	GetActivityById(id uint) (activity *Activity, err error)
	// SaveActivity 保存活动到数据库, 当err为nil时保存成功
	SaveActivity(activity *Activity) (err error)
	// DeleteActivity 删除活动, 当err为nil时删除成功
	DeleteActivity(activity *Activity) (err error)
	// GetActivityPilots 获取所有参加活动的飞行员, 当err为nil时返回值pilots有效
	GetActivityPilots(activity *Activity) (pilots []*ActivityPilot, err error)
	// GetActivityATCs 获取所有参加活动的管制员, 当err为nil时返回值acts有效
	GetActivityATCs(activity *Activity) (atcs []*ActivityATC, err error)
	// GetActivityFacilities 获取活动所有管制席位, 当err为nil时返回值facilities有效
	GetActivityFacilities(activity *Activity) (facilities []*ActivityFacility, err error)
	// SetActivityStatus 设置活动状态, 当err为nil时设置成功
	SetActivityStatus(activity *Activity, status ActivityStatus) (err error)
	// SetActivityStatusOpen 设置活动状态为开放报名, 当err为nil时设置成功
	SetActivityStatusOpen(activity *Activity) (err error)
	// SetActivityStatusActive 设置活动状态为活动中, 当err为nil时设置成功
	SetActivityStatusActive(activity *Activity) (err error)
	// SetActivityStatusClosed 设置活动状态为已结束, 当err为nil时设置成功
	SetActivityStatusClosed(activity *Activity) (err error)
	// SaveActivityPilot 保存参与活动的飞行员到数据库, 当err为nil时保存成功
	SaveActivityPilot(activityPilot *ActivityPilot) (err error)
	// SetActivityPilotStatus 设置参与活动的飞行员的状态, 当err为nil时设置成功
	SetActivityPilotStatus(activityPilot *ActivityPilot, status ActivityPilotStatus) (err error)
	// SetActivityPilotStatusSigned 设置参与活动的飞行员状态为已报名, 当err为nil时设置成功
	SetActivityPilotStatusSigned(activityPilot *ActivityPilot) (err error)
	// SetActivityPilotStatusClearance 设置参与活动的飞行员状态为已放行, 当err为nil时设置成功
	SetActivityPilotStatusClearance(activityPilot *ActivityPilot) (err error)
	// SetActivityPilotStatusTakeoff 设置参与活动的飞行员状态为已起飞, 当err为nil时设置成功
	SetActivityPilotStatusTakeoff(activityPilot *ActivityPilot) (err error)
	// SetActivityPilotStatusLanding 设置参与活动的飞行员状态为已落地, 当err为nil时设置成功
	SetActivityPilotStatusLanding(activityPilot *ActivityPilot) (err error)
	// SaveActivityAtc 保存参与活动的管制员到数据库, 当err为nil时保存成功
	SaveActivityAtc(activityAtc *ActivityATC) (err error)
	// SetActivityFacility 设置报名席位的用户, 当err为nil时保存成功
	SetActivityFacility(activity *ActivityFacility, user *User) (err error)
}
