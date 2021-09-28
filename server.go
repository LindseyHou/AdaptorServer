package main

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
	logFile, err := os.OpenFile("adaptor_server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func main() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"hello": "world",
		})
	})
	router.POST("/data", dataHandler)
	addr := "127.0.0.1:8000"
	if gin.EnvGinMode == gin.ReleaseMode {
		addr = "0.0.0.0:8000"
	}
	router.Run(addr)
}

func dataHandler(c *gin.Context) {
	var m map[string]interface{}
	err := c.Bind(&m)
	if err != nil {
		return
	}
	log.WithFields(logrus.Fields{"data": m}).Info(c.Request.URL)
	c.AbortWithStatus(204)
}
