package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	log.Println("<- SetupRouter()")

	router := gin.Default()

	log.Println("-> SetupRouter()")
	return router
}

func StartServer(port string, router *gin.Engine) {
	log.Println("start server at localhost", port)
	log.Fatal(http.ListenAndServe(port, router))
}
