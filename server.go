package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type fireData struct {
	Data     map[string]interface{} `json:"data"`
	DataCode string                 `json:"dataCode"`
	PostTime string                 `json:"postTime"`
}

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
	resp, err := http.Post("http://127.0.0.1:8008/api/v1/data/status", "application/json",
		bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	log.WithFields(logrus.Fields{"data": data, "res": res["json"]}).Info(c.Request.URL)
	c.AbortWithStatus(204)
}

func deviceType_to_partType(DeviceType int) int {
	dict := map[int]int{
		341: 1,
		27:  10,
		360: 11,
		361: 12,
		362: 13,
		342: 16,
		363: 17,
		364: 18,
		343: 21, // 21, 141
		327: 22,
		46:  23, //23, 138
		331: 24,
		329: 25,
		354: 30,
		355: 31,
		501: 32,
		502: 33,
		356: 34,
		503: 35,
		504: 36,
		357: 37,
		350: 40,
		505: 41,
		506: 42,
		507: 43,
		351: 44,
		365: 50,
		508: 51,
		344: 52,
		345: 53,
		366: 61,
		367: 62,
		368: 69,
		369: 78,
		370: 79,
		346: 81,
		371: 82,
		328: 83,
		372: 84,
		373: 85,
		374: 86,
		375: 87,
		114: 88,
		434: 81,
		45:  92, // 92, 401
		585: 93,
		376: 95, // 95, 256
		319: 96,
		310: 97,
		321: 98,
		318: 99,
		377: 101,
		386: 102,
		322: 103,
		387: 104,
		383: 105,
		378: 106,
		379: 113,
		380: 114,
		381: 115,
		336: 116,
		335: 117,
		334: 118,
		382: 121,
		388: 128,
		389: 129,
		390: 130,
		391: 131,
		392: 132,
		393: 133,
		325: 134,
		394: 135,
		395: 136,
		330: 137,
		326: 139,
		323: 140,
		396: 142,
		418: 143,
		397: 144,
		398: 145,
		399: 146,
		400: 147,
		348: 148,
		401: 149,
		332: 150,
		402: 151,
		403: 152,
		404: 153,
		405: 154,
		349: 155,
		409: 156,
		410: 157,
		411: 158,
		412: 159,
		413: 160,
		414: 161,
		415: 162,
		416: 163,
		417: 164,
		406: 165,
		408: 167,
		57:  168,
		320: 169,
		333: 170,
		385: 172,
		514: 174,
		72:  199,
		269: 202,
		419: 257,
		420: 258,
		425: 301,
		426: 302,
		427: 303,
		428: 304,
		429: 305,
		578: 351,
		577: 352,
		433: 402,
		359: 452,
		358: 453,
		584: 584,
		583: 583,
		575: 575,
		549: 549,
		548: 548,
		547: 547,
		515: 517,
		509: 507,
		499: 499,
		498: 498,
		435: 435,
		431: 431,
		430: 430,
		424: 424,
		423: 423,
		422: 422,
		421: 421,
		407: 407,
		384: 384,
		353: 353,
		352: 354,
		347: 347,
		339: 339,
		338: 338,
		337: 337,
		324: 324,
	}
	return dict[DeviceType]
}

func convert_json(raw fireData) (map[string]interface{}, bool) {
	data := raw.Data
	var partCode = data["device_id"]
	var time = raw.PostTime
	ret := map[string]interface{}{
		"partType":    0, // todo: change device_id → DeviceType → partType
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
