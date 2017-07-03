package main

import (
	"log"
	"strconv"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

var (
	labels = []string{"model", "channel", "id", "name"}
)

type DeviceMessage struct {
	Time       string  `json:"time"`
	Model      string  `json:"model"`
	ID         int     `json:"device"`
	Channel    int     `json:"channel"`
	TempF      float64 `json:"temperature_F"`
	Humidity   float64 `json:"humidity"`
	LowBattery string  `json:"battery"`
	Name       string
}

func (msg *DeviceMessage) Tags() map[string]string {
	tag := make(map[string]string)
	tag["model"] = msg.Model
	tag["channel"] = strconv.Itoa(msg.Channel)
	tag["id"] = strconv.Itoa(msg.ID)
	tag["name"] = msg.Name
	return tag
}

func (msg *DeviceMessage) Fields() map[string]interface{} {
	field := map[string]interface{}{
		"temperature": msg.TempF,
		"humidity":    msg.Humidity,
		"low_battery": false,
	}
	if msg.LowBattery != "OK" {
		field["low_battery"] = true
	}

	return field
}

func (msg *DeviceMessage) ToInfluxPoint() *client.Point {
	// Create Point
	pt, err := client.NewPoint("temp_sensor", msg.Tags(), msg.Fields(), time.Now())
	if err != nil {
		log.Println(err)
		return nil
	}

	return pt
}
