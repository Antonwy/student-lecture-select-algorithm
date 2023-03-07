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

type LectureOverlapped struct {
	LectureId  string
	Overlapped bool
}

type LectureToOverlapped map[string]*LectureOverlapped
type DeviceToOverlappedLectures map[string][]*LectureOverlapped
type CompareFunc[T comparable] func(T) bool

func FindPerfectMatch() {
	totalStartTime := time.Now()

	lectures := db.GetLectures()
	devices := db.GetReadyDevices()
	devicesLectures := db.GetDeviceLectures()
	maxAttendedLecturesCount := db.GetMaxAttendedLecturesCount()

	log.Infof("Time to execute SQL queries: %s", time.Now().Sub(totalStartTime))

	startTimeToCreateMaps := time.Now()

	lecturesToOverlappedMap := lectureToOverlappedLectureMap(lectures)
	devicesLecturesMap := devicesLecturesToMap(devicesLectures, lecturesToOverlappedMap)
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
	devicesLectures *DeviceToOverlappedLectures,
	lectures *[]model.IOSLecture,
	maxAttendedLecturesCount int,
) *[]string {
	var overlapped []*LectureOverlapped
	var overlappingStudents []string
	currentMaxAttended := maxAttendedLecturesCount

	for len(overlapped) < len(*lectures) {
		var newMax string
		var overlappingLectures *[]*LectureOverlapped

		newMax, overlappingLectures = findBestNextMatch(
			devicesLectures,
			currentMaxAttended,
		)

		if newMax == "" {
			break
		}

		delete(*devicesLectures, newMax)

		overlappingStudents = append(overlappingStudents, newMax)

		overlapped = append(overlapped, *overlappingLectures...)

		for _, lecture := range *overlappingLectures {
			lecture.Overlapped = true
		}

		overlappingLecturesCount := len(*overlappingLectures)

		if currentMaxAttended > overlappingLecturesCount {
			currentMaxAttended = overlappingLecturesCount
		}
	}

	return &overlappingStudents
}

func findBestNextMatch(
	devicesLectures *DeviceToOverlappedLectures,
	currentMaxAttended int,
) (string, *[]*LectureOverlapped) {
	maxAttends := 0
	studentWithMaxAttends := ""
	var overlappingLectures []*LectureOverlapped

	for device, lectures := range *devicesLectures {
		newAttends := filter(&lectures, func(lecture *LectureOverlapped) bool {
			return !lecture.Overlapped
		})

		newAttendsCount := len(*newAttends)

		if newAttendsCount > maxAttends {
			maxAttends = newAttendsCount
			studentWithMaxAttends = device
			overlappingLectures = *newAttends

			if maxAttends == currentMaxAttended {
				break
			}
		}
	}

	return studentWithMaxAttends, &overlappingLectures
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

func lectureToOverlappedLectureMap(l *[]model.IOSLecture) *LectureToOverlapped {
	lectures := make(LectureToOverlapped)

	for _, lecture := range *l {
		overlapped := LectureOverlapped{
			LectureId:  lecture.Id,
			Overlapped: false,
		}

		lectures[lecture.Id] = &overlapped
	}

	return &lectures
}

func devicesLecturesToMap(dl *[]model.IOSDeviceLecture, lectures *LectureToOverlapped) *DeviceToOverlappedLectures {
	devicesLectures := make(DeviceToOverlappedLectures)

	for _, d := range *dl {
		overlappedLecture := (*lectures)[d.LectureId]

		devicesLectures[d.DeviceId] = append(devicesLectures[d.DeviceId], overlappedLecture)
	}

	return &devicesLectures
}

func devicesToMap(d *[]model.IOSDevice) *map[string]model.IOSDevice {
	devices := make(map[string]model.IOSDevice)

	for _, device := range *d {
		devices[device.DeviceID] = device
	}

	return &devices
}

func checkIfAggregatedStudentsHaveAllLectures(aggregatedStudents *[]string) bool {
	lectures := db.GetLectures()
	lecturesCount := len(*lectures)
	deviceLectures := db.GetDeviceLectures()

	lectureToOverlappedLectures := lectureToOverlappedLectureMap(lectures)
	deviceLecturesMap := devicesLecturesToMap(deviceLectures, lectureToOverlappedLectures)

	log.Infof("Lecture count: %d", lecturesCount)
	log.Infof("Aggregated students count: %d", len(*aggregatedStudents))

	lectureDevicesMap := make(map[string][]string)

	for student, lectures := range *deviceLecturesMap {
		for _, l := range lectures {
			lectureDevicesMap[l.LectureId] = append(lectureDevicesMap[l.LectureId], student)
		}
	}

	log.Infof("Lecture devices map count: %d", len(lectureDevicesMap))

	if len(lectureDevicesMap) != lecturesCount {
		return false
	}

	return true
}
