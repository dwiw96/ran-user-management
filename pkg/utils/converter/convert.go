package converter

import (
	"fmt"
	"log"
	"strconv"
)

func ConvertStrToInt(s string) (int, error) {
	res, err := strconv.Atoi(s)
	return res, err
}

func ConvertStrToInt32(s string) (int32, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Print("convert str to int32")
			log.Println("recovered from panic:", r)
		}
	}()

	res, err := strconv.Atoi(s)
	if err != nil {
		log.Println("failed to convert string to int32, err:", err)
		return -1, fmt.Errorf("")
	}

	return int32(res), err
}

func ConvertInt32ToString(n int32) string {
	return strconv.Itoa(int(n))
}
