package main

import (
	log "github.com/sirupsen/logrus"
	"test-student-lecture-selection-algorithm/db"
	"test-student-lecture-selection-algorithm/model"
	"time"
)

func main() {
	db.Init()

	FindPerfectMatch()
}

type DeviceLectureAttend map[string][]string
type CompareFunc[T comparable] func(T) bool

func FindPerfectMatch() {
	totalStartTime := time.Now()

	lectures := db.GetLectures()
	devices := db.GetReadyDevices()
	devicesLectures := db.GetDeviceLectures()
	maxAttendedLecturesCount := db.GetMaxAttendedLecturesCount()

	log.Infof("Time to execute SQL queries: %s", time.Now().Sub(totalStartTime))

	startTimeToCreateMaps := time.Now()

	devicesLecturesMap := devicesLecturesToMap(devicesLectures)
	devicesDetailsMap := devicesToMap(devices)

	log.Infof("Time to create maps: %s", time.Now().Sub(startTimeToCreateMaps))

	log.Infof("Lecture count: %d", len(*lectures))
	log.Infof("Device count: %d", len(*devicesDetailsMap))
	log.Infof("Device x lecture count: %d", len(*devicesLectures))

	log.Infof("------------------")

	log.Infof("Searching for perfect student set...")

	startTime := time.Now()

	overlappingStudents := getOverlapping(
		devicesLecturesMap,
		devicesDetailsMap,
		lectures,
		maxAttendedLecturesCount,
	)

	log.Infof("Found perfect set: %d (students) in %s", len(*overlappingStudents), time.Now().Sub(startTime))

	log.Infof("------------------")

	log.Infof("Total execution time: %s", time.Now().Sub(totalStartTime))

	log.Infof("------------------")

	log.Infof("Checking if all lectures are covered...")

	log.Infof("All lectures are covered: %t", checkIfAggregatedStudentsHaveAllLectures(overlappingStudents))
}

func getOverlapping(
	devicesLectures *DeviceLectureAttend,
	devicesDetails *map[string]model.IOSDeviceWithAvgResponseTime,
	lectures *[]model.IOSLecture,
	maxAttendedLecturesCount int,
) *[]string {
	var overlapped []string
	var overlappingStudents []string
	currentMaxAttended := maxAttendedLecturesCount

	for len(overlapped) < len(*lectures) {
		var newMax string
		var overlappingLectures *[]string

		newMax, overlappingLectures = findBestNextMatch(
			devicesLectures,
			devicesDetails,
			&overlappingStudents,
			currentMaxAttended,
		)

		if newMax == "" {
			break
		}

		delete(*devicesLectures, newMax)

		overlappingStudents = append(overlappingStudents, newMax)

		overlapped = append(overlapped, *overlappingLectures...)

		overlappingLecturesCount := len(*overlappingLectures)

		if currentMaxAttended > overlappingLecturesCount {
			currentMaxAttended = overlappingLecturesCount
		}
	}

	return &overlappingStudents
}

func findBestNextMatch(
	devicesLectures *DeviceLectureAttend,
	devicesDetails *map[string]model.IOSDeviceWithAvgResponseTime,
	overlapped *[]string,
	currentMaxAttended int,
) (string, *[]string) {
	maxAttends := 0
	studentWithMaxAttends := ""
	var overlappingLectures []string
	var responseTime float64

	for student, lectures := range *devicesLectures {
		newAttends := filter(&lectures, func(lecture string) bool {
			return !contains(overlapped, lecture)
		})

		newAttendsCount := len(*newAttends)

		if newAttendsCount > maxAttends || (newAttendsCount == maxAttends && (*devicesDetails)[student].AvgResponseTime < responseTime) {
			maxAttends = len(*newAttends)
			studentWithMaxAttends = student
			overlappingLectures = *newAttends
			responseTime = (*devicesDetails)[student].AvgResponseTime
		}
	}

	return studentWithMaxAttends, &overlappingLectures
}

func areEqual[T comparable](s *[]T, t *[]T) bool {
	if len(*s) != len(*t) {
		return false
	}

	for i, v := range *s {
		if v != (*t)[i] {
			return false
		}
	}

	return true
}

func filter[T comparable](s *[]T, fn CompareFunc[T]) *[]T {
	var p []T
	for _, v := range *s {
		if fn(v) {
			p = append(p, v)
		}
	}
	return &p
}

func contains[T comparable](s *[]T, e T) bool {
	for _, a := range *s {
		if a == e {
			return true
		}
	}
	return false
}

func devicesLecturesToMap(dl *[]model.IOSDeviceLecture) *DeviceLectureAttend {
	devicesLectures := make(DeviceLectureAttend)

	for _, d := range *dl {
		devicesLectures[d.DeviceId] = append(devicesLectures[d.DeviceId], d.LectureId)
	}

	return &devicesLectures
}

func devicesToMap(d *[]model.IOSDeviceWithAvgResponseTime) *map[string]model.IOSDeviceWithAvgResponseTime {
	devices := make(map[string]model.IOSDeviceWithAvgResponseTime)

	for _, device := range *d {
		devices[device.DeviceID] = device
	}

	return &devices
}

func checkIfAggregatedStudentsHaveAllLectures(aggregatedStudents *[]string) bool {
	lectures := db.GetLecturesThatHaveAtLeastOneDevice()
	lecturesCount := len(*lectures)
	deviceLectures := db.GetDeviceLectures()

	deviceLecturesMap := devicesLecturesToMap(deviceLectures)

	log.Infof("Lecture count: %d", lecturesCount)
	log.Infof("Aggregated students count: %d", len(*aggregatedStudents))

	lectureDevicesMap := make(map[string][]string)

	for student, lectures := range *deviceLecturesMap {
		for _, l := range lectures {
			lectureDevicesMap[l] = append(lectureDevicesMap[l], student)
		}
	}

	log.Infof("Lecture devices map count: %d", len(lectureDevicesMap))

	if len(lectureDevicesMap) != lecturesCount {
		return false
	}

	return true
}
