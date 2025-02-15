package utils

import (
	"strconv"
	"time"
)

func BoolAddr(b bool) *bool {
	boolVar := b
	return &boolVar
}

func StringAddr(b string) *string {
	stringVar := b
	return &stringVar
}

func IsInteger(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

func ValidateExpiryDate(expiryDate string) bool {
	dateFormat := "02-01-2006"
	currentDate := time.Now()
	expiryDateObj, err := time.Parse(dateFormat, expiryDate)
	if err != nil {
		return false
	}
	return expiryDateObj.After(currentDate) && expiryDateObj.Year() < 2099
}
