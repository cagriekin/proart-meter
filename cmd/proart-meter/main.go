package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cagri/proart-meter/internal/config"
	"github.com/cagri/proart-meter/internal/device"
	"github.com/cagri/proart-meter/internal/hwmon"
)

func main() {
	configPath := flag.String("config", "/etc/proart-meter/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	sensorPath, err := hwmon.FindSensor(cfg.Sensor)
	if err != nil {
		log.Fatalf("sensor: %v", err)
	}

	dev, err := device.Open()
	if err != nil {
		log.Fatalf("device: %v", err)
	}
	defer dev.Close()

	if err := dev.Init(); err != nil {
		log.Fatalf("device init: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	log.Printf("sensor=%s interval=%ds thresholds=%v", sensorPath, cfg.PollIntervalSeconds, cfg.Thresholds)

	currentLevel := -1
	ticker := time.NewTicker(time.Duration(cfg.PollIntervalSeconds) * time.Second)
	defer ticker.Stop()

	// Read and set initial level before entering ticker loop
	temp, err := hwmon.ReadTemp(sensorPath)
	if err != nil {
		log.Fatalf("initial temp read: %v", err)
	}
	currentLevel = tempToLevel(temp, cfg.Thresholds)
	if err := dev.SetMeterLevel(currentLevel); err != nil {
		log.Fatalf("initial meter set: %v", err)
	}
	log.Printf("%.1fC -> %d/5", temp, currentLevel)

	for {
		select {
		case <-ctx.Done():
			if err := dev.SetMeterLevel(0); err != nil {
				log.Printf("shutdown meter reset: %v", err)
			}
			log.Print("stopped")
			return
		case <-ticker.C:
			temp, err := hwmon.ReadTemp(sensorPath)
			if err != nil {
				log.Printf("temp read: %v", err)
				continue
			}
			level := tempToLevel(temp, cfg.Thresholds)
			if level != currentLevel {
				if err := dev.SetMeterLevel(level); err != nil {
					log.Fatalf("meter set: %v", err)
				}
				log.Printf("%.1fC -> %d/5", temp, level)
				currentLevel = level
			}
		}
	}
}

func tempToLevel(tempC float64, thresholds []int) int {
	level := 0
	for _, t := range thresholds {
		if tempC >= float64(t) {
			level++
		}
	}
	return level
}
