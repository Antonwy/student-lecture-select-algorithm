package model

import (
	"time"
)

const (
	IOSLectureSemesterWinter = "winter"
	IOSLectureSemesterSummer = "summer"
)

type IOSLecture struct {
	Id            string               `gorm:"primaryKey"`
	Year          int16                `faker:"boundary_start=2022, boundary_end=2023"`
	Semester      string               `gorm:"type:enum ('winter', 'summer');" faker:"oneof: winter, summer"`
	LastUpdate    time.Time            `gorm:"default:now()" faker:"-"`
	LastRequestId *string              `gorm:"default:NULL;" faker:"-"`
	LastRequest   *IOSDeviceRequestLog `gorm:"constraint:OnDelete:SET NULL;" faker:"-"`
}
