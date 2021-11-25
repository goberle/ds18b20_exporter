package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"periph.io/x/devices/v3/ds18b20"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/netlink"
)

func recordMetrics() {
	log.Println("Initializing...")

	_, err := host.Init()
	if err != nil {
		log.Println(err)
	}

	oneBus, err := netlink.New(001)
	if err != nil {
		log.Println(err)
	}

	addrs, err := oneBus.Search(false)
	if err != nil {
		log.Println(err)
	}

	var sensors []*ds18b20.Dev

	for _, addr := range addrs {
		sensor, err := ds18b20.New(oneBus, addr, 10)
		if err != nil {
			log.Println(err)
		}
		sensors = append(sensors, sensor)
	}

	log.Printf("Found sensors : %v\n", sensors)
	log.Println("Initialized!")

	ticker := time.NewTicker(5 * time.Second)

	for _ = range ticker.C {
		err := ds18b20.ConvertAll(oneBus, 10)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, sensor := range sensors {
			sensorTemp, err := sensor.LastTemp()
			if err != nil {
				log.Printf("%s: %s", sensor.String(), err)
				continue
			}
			sensorsMetrics.WithLabelValues(sensor.String()).Set(sensorTemp.Celsius())
			log.Printf("%s: %s", sensor.String(), sensorTemp.String())
		}
	}
}

var (
	sensorsMetrics = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ds18b20_temperature",
		Help: "DS18B20 sensors temperature",
	},
		[]string{"sensor"})
)

func main() {
	go recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":9101", nil)
	if err != nil {
		log.Println(err)
	}
}
