package validator

import (
	"time"

	"github.com/asaskevich/govalidator"
)

const (
	CustomDateFormat string = "2006-01-02"
)

func Init() {
	// `valid:"customDate,"`
	govalidator.TagMap["customDate"] = govalidator.Validator(func(str string) bool {
		_, err := time.Parse(CustomDateFormat, str)
		return err == nil
	})
}
