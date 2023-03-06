package db

import (
	"database/sql"
	"github.com/bxcodec/faker/v4"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"math/rand"
	"os"
	"test-student-lecture-selection-algorithm/model"
	"time"
)

var DB *gorm.DB
var dbHost = os.Getenv("DB_DSN")

func Init() {
	log.Infof("Connecting to dsn: %s\n", dbHost)

	conn := mysql.Open(dbHost)

	db, err := gorm.Open(conn, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		panic("failed to connect database")
	}

	DB = db

	err = Migrate()

	if err != nil {
		log.WithError(err).Error("Migrate failed")
		return
	}

	PopulateWithDummyData()
}

func Migrate() error {
	err := DB.AutoMigrate(
		&model.IOSDevice{},
		&model.IOSDeviceRequestLog{},
		&model.IOSEncryptedGrade{},
		&model.IOSScheduledUpdateLog{},
		&model.IOSSchedulingPriority{},
		&model.IOSLecture{},
		&model.IOSDeviceLecture{},
	)
	return err
}

func GetDevices() *[]model.IOSDevice {
	var devices []model.IOSDevice

	DB.Model(&model.IOSDevice{}).Find(&devices)

	return &devices
}

func GetReadyDevices() *[]model.IOSDeviceWithAvgResponseTime {
	var devices []model.IOSDeviceWithAvgResponseTime

	DB.Raw(buildDevicesWithAvgResponseTimeQuery()).Scan(&devices)

	return &devices
}

func GetLectures() *[]model.IOSLecture {
	var lectures []model.IOSLecture

	DB.Model(&model.IOSLecture{}).Find(&lectures)

	return &lectures
}

func GetDeviceLectures() *[]model.IOSDeviceLecture {
	var deviceLectures []model.IOSDeviceLecture

	DB.Model(&model.IOSDeviceLecture{}).Find(&deviceLectures)

	return &deviceLectures
}

func GetLecturesThatHaveAtLeastOneDevice() *[]string {
	var lectureIds []string

	DB.Raw("select dl.lecture_id from ios_device_lectures dl group by dl.lecture_id;").Scan(&lectureIds)

	return &lectureIds
}

func GetMaxAttendedLecturesCount() int {
	var maxCount int

	DB.Raw("select max(lecture_count) from (select count(*) as lecture_count from ios_device_lectures group by device_id) as t;").Scan(&maxCount)

	return maxCount
}

func PopulateWithDummyData() {
	devices := fakeDevices(20000)
	lectures := fakeLectures(500)

	fakeDeviceLectureRelation(devices, lectures)

	fakeDevicesRequestLogs(devices)

	log.Infof("Populated %d devices and %d lectures", len(*devices), len(*lectures))
}

func fakeDevices(count int) *[]model.IOSDevice {
	var devicesCount int64
	DB.Model(&model.IOSDevice{}).Count(&devicesCount)

	if int(devicesCount) >= count {
		log.Warn("Devices already populated")

		return GetDevices()
	}

	var devices []model.IOSDevice

	for i := 0; i < count; i++ {
		device := model.IOSDevice{}

		faker.FakeData(&device)

		devices = append(devices, device)

		if i%1000 == 0 {
			DB.Create(&devices)
			devices = []model.IOSDevice{}
		}
	}

	DB.Create(&devices)

	return GetDevices()
}

func fakeLectures(count int) *[]model.IOSLecture {
	var lecturesCount int64
	DB.Model(&model.IOSLecture{}).Count(&lecturesCount)

	if int(lecturesCount) >= count {
		log.Warn("Lectures already populated")

		return GetLectures()
	}

	var lectures []model.IOSLecture

	for i := 0; i < count; i++ {
		lecture := model.IOSLecture{}

		faker.FakeData(&lecture)

		lectures = append(lectures, lecture)

		if i%1000 == 0 {
			DB.Create(&lectures)
			lectures = []model.IOSLecture{}
		}
	}

	DB.Create(&lectures)

	return GetLectures()
}

func fakeDeviceLectureRelation(devices *[]model.IOSDevice, lectures *[]model.IOSLecture) {
	var deviceLecturesCount int64
	DB.Model(&model.IOSDeviceLecture{}).Count(&deviceLecturesCount)

	if int(deviceLecturesCount) > 0 {
		log.Warn("Lectures already populated")

		return
	}

	rand.Seed(time.Now().Unix())

	tmpLectures := *lectures

	var deviceLectures []model.IOSDeviceLecture

	for _, device := range *devices {
		numLectures := rand.Intn(8) + 1

		for i := 0; i < numLectures; i++ {
			lecture := tmpLectures[0]
			deviceLecture := model.IOSDeviceLecture{
				LectureId: lecture.Id,
				DeviceId:  device.DeviceID,
			}

			deviceLectures = append(deviceLectures, deviceLecture)

			// Remove the assigned lecture from the list
			tmpLectures = tmpLectures[1:]
			if len(tmpLectures) == 0 {
				tmpLectures = *lectures
				break
			}
		}

		if len(deviceLectures)%1000 == 0 {
			DB.Create(&deviceLectures)
			deviceLectures = []model.IOSDeviceLecture{}
		}
	}

	DB.Create(&deviceLectures)
}

func fakeDevicesRequestLogs(devices *[]model.IOSDevice) {
	var devicesRequestLogs int64
	DB.Model(&model.IOSDeviceRequestLog{}).Count(&devicesRequestLogs)

	if int(devicesRequestLogs) > 0 {
		log.Warn("Request logs already populated")

		return
	}

	var requestLogs []model.IOSDeviceRequestLog

	for _, device := range *devices {
		requestLogsCount := rand.Intn(10) + 1

		for i := 0; i < requestLogsCount; i++ {
			timeToAdd := time.Duration(rand.Intn(50)+10+i*30) * time.Minute * -1

			createdRequest := time.Now().Add(timeToAdd)

			handledRequest := sql.NullTime{
				Time:  createdRequest.Add(time.Duration(rand.Intn(59)+1) * time.Second),
				Valid: true,
			}

			if rand.Intn(2) == 0 {
				handledRequest = sql.NullTime{
					Time:  time.Time{},
					Valid: false,
				}
			}

			requestLog := model.IOSDeviceRequestLog{
				DeviceID:    device.DeviceID,
				RequestType: model.IOSLectureUpdateRequestType,
				CreatedAt:   createdRequest,
				HandledAt:   handledRequest,
			}

			requestLogs = append(requestLogs, requestLog)

			if len(requestLogs)%1000 == 0 {
				DB.Create(&requestLogs)
				requestLogs = []model.IOSDeviceRequestLog{}
			}
		}
	}

	DB.Create(&requestLogs)
}

func pickNRandomElements[T interface{}](n int, elements *[]T) []T {
	if n > len(*elements) {
		log.Panicf("n (%d) is greater than elements length (%d)", n, len(*elements))
	}

	rand.Seed(time.Now().Unix())

	elementsMap := map[int]T{}

	for i := 0; i < n; i++ {
		randomIndex := rand.Intn(len(*elements))

		if _, ok := elementsMap[randomIndex]; ok {
			i--
			continue
		}

		elementsMap[randomIndex] = (*elements)[randomIndex]
	}

	var result []T

	for _, element := range elementsMap {
		result = append(result, element)
	}

	return result
}

func buildDevicesWithAvgResponseTimeQuery() string {
	return `
with ios_devices_response_time
         as (select avg(timestampdiff(SECOND, drl.created_at, drl.handled_at)) as avg_response_time, d.device_id
             from ios_devices d
                      left join ios_device_request_logs drl on d.device_id = drl.device_id
             where drl.handled_at is not null
             group by d.device_id)
select d.*, drt.avg_response_time as avg_response_time
from ios_devices d
         left join ios_devices_response_time drt on d.device_id = drt.device_id
group by d.device_id
order by avg_response_time asc;
`
}
