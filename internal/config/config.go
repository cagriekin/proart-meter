package config

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Sensor              string `yaml:"sensor"`
	PollIntervalSeconds int    `yaml:"poll_interval_seconds"`
	Thresholds          []int  `yaml:"thresholds"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if cfg.Sensor == "" {
		return fmt.Errorf("sensor is required")
	}
	if cfg.PollIntervalSeconds <= 0 {
		return fmt.Errorf("poll_interval_seconds must be > 0")
	}
	if len(cfg.Thresholds) != 5 {
		return fmt.Errorf("thresholds must have exactly 5 entries (got %d)", len(cfg.Thresholds))
	}
	if !sort.IntsAreSorted(cfg.Thresholds) {
		return fmt.Errorf("thresholds must be in ascending order")
	}
	for i, t := range cfg.Thresholds {
		if t <= 0 {
			return fmt.Errorf("thresholds[%d] must be > 0 (got %d)", i, t)
		}
	}
	return nil
}
