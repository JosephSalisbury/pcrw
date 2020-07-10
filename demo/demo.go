package main

import (
	"flag"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/JosephSalisbury/pcrw"
)

var (
	gauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pcrw_example_gauge",
		Help: "Gauge as an example of pcrw.",
	})

	url = *flag.String("url", "http://localhost:1234/receive", "URL for remote write config")
)

func init() {
	prometheus.MustRegister(gauge)
}

func main() {
	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stdout)

	if url == "" {
		logger.Log("err", "remote write URL cannot be empty")
		os.Exit(1)
	}

	go func() {
		ticker := time.NewTicker(15 * time.Second)

		for {
			select {
			case t := <-ticker.C:
				logger.Log("msg", "setting gauge to current number of seconds")
				gauge.Set(float64(t.Second()))
			}
		}
	}()

	if err := pcrw.Push(logger, prometheus.DefaultRegisterer, prometheus.DefaultGatherer, 30*time.Second, url); err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}
}
