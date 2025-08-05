package utils

import "strconv"

func StrToInt(str string, defaultValue int) int {
	result, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}
	return result
}

func StrToFloat(str string, defaultValue float64) float64 {
	result, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return defaultValue
	}
	return result
}
