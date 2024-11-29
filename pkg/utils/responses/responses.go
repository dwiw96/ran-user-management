package responses

import (
	// "encoding/json"
	"log"
	// "net/http"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

var errorMsg = map[int]string{
	422: "Unprocessable Entity",
	409: "Conflict",
	500: "Internal Server Error",
	400: "Bad Request",
	401: "Unauthorized",
}

func ErrorJSON(c *gin.Context, code int, desc []string, remoteAddr string) {
	msg := errorMsg[code]
	response := FailedResponse(msg, desc)

	var temp string
	for _, v := range desc {
		if temp != "" {
			temp += ", "
		}
		temp += v
	}

	err := fmt.Errorf("%v", temp)
	c.Error(err)
	c.IndentedJSON(code, response)
}

func FailedResponse(msg, desc interface{}) map[string]interface{} {
	return map[string]interface{}{
		"error_message": msg,
		"result":        "failure",
		"description":   desc,
		"execute_at":    time.Now().UTC().Add(time.Hour * 9).Format("2006/01/02 15:04:05.000"),
	}
}

func SuccessWithDataResponse(data interface{}, code int, msg string) map[string]interface{} {
	log.Printf(">>> %d, response: %s\n", code, msg)
	return map[string]interface{}{
		"error_message": "",
		"result":        "success",
		"value":         data,
		"description":   msg,
		"execute_at":    time.Now().UTC().Add(time.Hour * 9).Format("2006/01/02 15:04:05.000"),
	}
}

func SuccessWithMultipleDataResponse(data []interface{}, msg string) map[string]interface{} {
	return map[string]interface{}{
		"error_message": "",
		"result":        "success",
		"value":         data,
		"description":   msg,
		"execute_at":    time.Now().UTC().Add(time.Hour * 9).Format("2006/01/02 15:04:05.000"),
	}
}

func SuccessWithDataResponsePagination(data interface{}, currentPage, totalPage int, msg string) map[string]interface{} {
	return map[string]interface{}{
		"error_message": "",
		"result":        "success",
		"value":         data,
		"pagination": map[string]int{
			"current_page": currentPage,
			"total_pages":  totalPage,
		},
		"description": msg,
		"execute_at":  time.Now().UTC().Add(time.Hour * 9).Format("2006/01/02 15:04:05.000"),
	}
}

func SuccessResponse(msg string) map[string]interface{} {
	return map[string]interface{}{
		"error_message": "",
		"result":        "success",
		"value":         "",
		"description":   msg,
		"execute_at":    time.Now().UTC().Add(time.Hour * 9).Format("2006/01/02 15:04:05.000"),
	}
}

func ErrorWithDataResponse(data interface{}, code int, err, msg string) map[string]interface{} {
	log.Printf(">>> %d, response: %s\n", code, msg)
	return map[string]interface{}{
		"error_message": err,
		"result":        "success",
		"value":         data,
		"description":   msg,
		"execute_at":    time.Now().UTC().Add(time.Hour * 9).Format("2006/01/02 15:04:05.000"),
	}
}
