package date

import (
	"os"
	"strconv"
	"time"
)

func GetToday() (int, error) {
	jst, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		return -1, err
	}
	now := time.Now().In(jst)
	today := now.Format("20060102")
	todayInt, err := strconv.Atoi(today)
	if err != nil {
		return -1, err
	}
	return todayInt, nil
}
