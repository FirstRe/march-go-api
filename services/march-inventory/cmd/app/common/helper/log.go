package helper

import (
	"fmt"
	"log"
)

type logs string

func LogContext(_class string, name string) logs {
	return logs(_class + ":" + name)
}

func (l logs) Logger(values []interface{}, name string) {

	var logMessage string
	for _, value := range values {
		logMessage += fmt.Sprintf("%v: %+v ", name, value)
	}
	log.Printf("%v %v", logMessage, l)

}
