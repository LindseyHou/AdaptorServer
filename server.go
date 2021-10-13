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
	router.Run(":8000")
}

func dataHandler(c *gin.Context) {
	type fireData struct {
		Data     map[string]interface{} `json:"data"`
		DataCode string                 `json:"dataCode"`
		PostTime string                 `json:"postTime"`
	}
	var data fireData
	if err := c.Bind(&data); err != nil {
		log.Info(c.Request.URL, "bind failed")
		c.AbortWithStatus(400)
	}
	log.WithFields(logrus.Fields{"data": data}).Info(c.Request.URL)
	c.AbortWithStatus(204)
}
