package model

import (
	"fmt"
	"time"
)

// IOSDevice stores relevant device information.
// E.g. the PublicKey which is used to encrypt push notifications
// The DeviceID can be used to send push notifications via APNs
type IOSDevice struct {
	DeviceID          string    `gorm:"primary_key" json:"deviceId" faker:"uuid_hyphenated"`
	CreatedAt         time.Time `gorm:"default:now()" json:"createdAt"`
	PublicKey         string    `gorm:"not null" json:"publicKey" faker:"jwt"`
	ActivityToday     int32     `gorm:"default:0" json:"activityToday" faker:"boundary_start=0, boundary_end=10"`
	ActivityThisWeek  int32     `gorm:"default:0" json:"activityThisWeek" faker:"boundary_start=10, boundary_end=50"`
	ActivityThisMonth int32     `gorm:"default:0" json:"activityThisMonth" faker:"boundary_start=50, boundary_end=100"`
	ActivityThisYear  int32     `gorm:"default:0" json:"activityThisYear" faker:"boundary_start=100, boundary_end=1000"`
}

type IOSDeviceWithAvgResponseTime struct {
	IOSDevice
	AvgResponseTime float64 `json:"avgResponseTime"`
}

func (device *IOSDevice) String() string {
	return fmt.Sprintf("IOSDevice{DeviceID: %s}", device.DeviceID)
}
