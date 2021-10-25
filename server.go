package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type fireData struct {
	Data     map[string]interface{} `json:"data"`
	DataCode string                 `json:"dataCode"`
	PostTime string                 `json:"postTime"`
}

var log = logrus.New()
var partCode2partType = map[string]int32{}
var endpoint_url string

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
	logFile, err := os.OpenFile(
		"adaptor_server.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func main() {
	if godotenv.Load() != nil {
		log.Fatal("Error loading .env file")
	}
	mongo_uri := os.Getenv("MONGO_URI")
	mongo_db_name := os.Getenv("MONGO_DB_NAME")
	endpoint_url = os.Getenv("ENDPOINT_URL")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(mongo_uri))
	collection := client.Database(mongo_db_name).Collection("info")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var itemBson bson.M
		var itemMap map[string]interface{}
		if err = cursor.Decode(&itemBson); err != nil {
			log.Fatal(err)
		}
		b, _ := bson.Marshal(itemBson)
		bson.Unmarshal(b, &itemMap)
		s := reflect.ValueOf(itemMap["datas"])
		for i := 0; i < s.Len(); i++ {
			data_map := s.Index(i).Interface().(map[string]interface{})
			partCode := data_map["partCode"].(string)
			partType := data_map["partType"].(int32)
			partCode2partType[partCode] = partType
		}
	}
	log.Info(
		fmt.Sprintf("Load partCode2partType succeed, result count: %d",
			len(partCode2partType)),
	)
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
	var data fireData
	if err := c.Bind(&data); err != nil {
		log.Info(c.Request.URL, "bind failed")
		c.AbortWithStatus(400)
	}
	values, _ := convert_json(data)
	json_data, err := json.Marshal(values)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post(
		endpoint_url,
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		log.Fatal(err)
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	log.WithFields(
		logrus.Fields{"data": data, "res": res["json"]},
	).Info(c.Request.URL)
	c.AbortWithStatus(204)
}

func convert_json(raw fireData) (map[string]interface{}, bool) {
	data := raw.Data
	var partCode = data["device_id"].(string)
	var time = raw.PostTime
	ret := map[string]interface{}{
		"partType":    partCode2partType[partCode],
		"partCode":    partCode,
		"time":        time,
		"fireAlarm":   0,
		"errorStatus": 0,
	}
	if _, ok := data["event_type"]; !ok {
		return nil, false
	}
	if data["event_type"] == "1" {
		ret["fireAlarm"] = 1
		ret["errorStatus"] = 0
	} else if data["event_type"] == "2" {
		ret["fireAlarm"] = 0
		ret["errorStatus"] = 1
	}
	return ret, true
}
