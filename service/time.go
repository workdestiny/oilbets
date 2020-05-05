package service

import (
	"strconv"
	"strings"
	"time"

	"github.com/workdestiny/amlporn/config"
)

// FormatTime is parse date to th e.g. วันที่ 01 ม.ค. 2552 เวลา 12:00 น.
func FormatTime(t time.Time) string {
	var sTime string
	formatTime := t.In(loc).Format("2006-01-02 15:04")
	values := strings.Split(formatTime, "-")
	if values[1] != "" {
		month := GetMonthString(values[1])
		dayTime := strings.Split(values[2], " ")
		convert, _ := strconv.Atoi(dayTime[0])
		day := strconv.Itoa(convert)
		sTime = day + " " + month + " " + values[0] + " / " + dayTime[1] + " น."
	}
	return sTime
}

// FormatDate is parse date to th e.g. 01 ม.ค. 2552
func FormatDate(t time.Time) string {
	var sTime string
	formatDate := t.In(loc).Format("2006-01-02 15:04")
	values := strings.Split(formatDate, "-")
	if values[1] != "" {
		month := GetMonthString(values[1])
		dayTime := strings.Split(values[2], " ")
		convert, _ := strconv.Atoi(dayTime[0])
		day := strconv.Itoa(convert)
		sTime = day + " " + month + " " + values[0]
	}
	return sTime
}

// ReFormatTime input timestamp output string (3 วินาทีแล้ว)
func ReFormatTime(ti time.Time) string {
	var sTime string

	now := time.Now().In(loc)
	ti = ti.In(loc)

	duration := now.Sub(ti)

	if duration < time.Minute {
		sDuration := strconv.FormatFloat(duration.Seconds(), 'f', 0, 64)
		sTime = sDuration + " วินาทีที่แล้ว"
		return sTime
	} else if duration < time.Hour {
		mDuration := strconv.FormatFloat(duration.Minutes(), 'f', 0, 64)
		sTime = mDuration + " นาที"
		return sTime
	} else if duration < 24*time.Hour {
		hDuration := strconv.FormatFloat(duration.Hours(), 'f', 0, 64)
		sTime = hDuration + " ชั่วโมง"
		return sTime
	}
	if duration >= 86400 {
		formatTime := ti.Format("2006-01-02 15:04:05")
		values := strings.Split(formatTime, "-")

		if values[1] != "" {
			month := GetMonthString(values[1])
			dayTime := strings.Split(values[2], " ")
			sTime = dayTime[0] + " " + month + " " + values[0] + " " + dayTime[1]
		} else {
			sTime = formatTime
		}
	}
	return sTime
}

// GetMonthString change month number to string
func GetMonthString(m string) string {

	var month string
	switch m {
	case "01":
		month = "มกราคม"
	case "02":
		month = "กุมภาพันธ์"
	case "03":
		month = "มีนาคม"
	case "04":
		month = "เมษายน"
	case "05":
		month = "พฤษภาคม"
	case "06":
		month = "มิถุนายน"
	case "07":
		month = "กรกฎาคม"
	case "08":
		month = "สิงหาคม"
	case "09":
		month = "กันยายน"
	case "10":
		month = "ตุลาคม"
	case "11":
		month = "พฤศจิกายน"
	case "12":
		month = "ธันวาคม"
	default:
		month = ""
	}
	return month
}

// GetShortMonthString is change format to มกราคม -> ม.ค.
func GetShortMonthString(m string) string {
	var month string
	switch m {
	case "01":
		month = "ม.ค."
	case "02":
		month = "ก.พ."
	case "03":
		month = "มี.ค."
	case "04":
		month = "เม.ย."
	case "05":
		month = "พ.ค."
	case "06":
		month = "มิ.ย."
	case "07":
		month = "ก.ค."
	case "08":
		month = "ส.ค."
	case "09":
		month = "ก.ย."
	case "10":
		month = "ต.ค."
	case "11":
		month = "พ.ย."
	case "12":
		month = "ธ.ค."
	default:
		month = ""
	}
	return month
}

// FormatPostViewCount is parse 21/09/18
func FormatPostViewCount(t time.Time) string {
	year, month, day := t.In(loc).Date()
	sDay := strconv.Itoa(day)
	sMonth := strconv.Itoa(int(month))
	if len(sDay) == 1 {
		sDay = "0" + sDay
	}
	if len(sMonth) == 1 {
		sMonth = "0" + sMonth
	}
	return sDay + "/" + sMonth + "/" + strconv.Itoa(year)[2:]
}

// FormatCustomType Time insights
func FormatCustomType(t time.Time) string {
	year, month, day := t.In(loc).Date()
	sDay := strconv.Itoa(day)
	sMonth := strconv.Itoa(int(month))
	if len(sMonth) == 1 {
		sMonth = "0" + sMonth
	}
	return sDay + " " + GetShortMonthString(sMonth) + " " + strconv.Itoa(year)
}

// FormatRevenueDateType Time insights
func FormatRevenueDateType(t time.Time) string {
	year, month, day := t.In(loc).Date()
	sDay := strconv.Itoa(day)
	sMonth := strconv.Itoa(int(month))
	if len(sMonth) == 1 {
		sMonth = "0" + sMonth
	}
	return sDay + " " + GetMonthString(sMonth) + " " + strconv.Itoa(year)
}

// FormatRevenueTimeType Time insights
func FormatRevenueTimeType(t time.Time) string {
	ti := t.In(loc)
	h := strconv.Itoa(ti.Hour())
	m := strconv.Itoa(ti.Minute())

	if ti.Hour() < 10 {
		h = "0" + h
	}

	if ti.Minute() < 10 {
		m = "0" + m
	}

	return h + ":" + m
}

// GetParserTime input type
func GetParserTime(typ string, start string, end string) (time.Time, time.Time) {

	var zeroTime time.Time
	startTime := time.Now()
	endTime := time.Now()
	day := startTime.Day()
	switch typ {
	case "all":
		return zeroTime, endTime
	case "today":
		return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC), endTime

	case "7dayago":
		return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -7), endTime

	case "lastmonth":
		return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, -1, -(day - 1)), time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(day - 1))

	case "30dayago":
		return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -30), endTime

	case "thismonth":
		return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(day - 1)), endTime

	case "custom":
		format := "2006-01-02"
		tStart, err := time.Parse(format, start)
		tEnd, err := time.Parse(format, end)
		if err != nil {
			return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(day - 1)), endTime
		}
		return tStart.UTC(), tEnd.UTC()

	default:
		return zeroTime, endTime
	}
}

//IsNewEditor Check New Editor
func IsNewEditor(datePost time.Time) bool {
	if datePost.UTC().Unix() > config.PivotDay.Unix() {
		return true
	}
	return false
}
