package hwmon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const hwmonBase = "/sys/class/hwmon"

func FindSensor(hwmonName string) (string, error) {
	entries, err := os.ReadDir(hwmonBase)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", hwmonBase, err)
	}

	for _, entry := range entries {
		namePath := filepath.Join(hwmonBase, entry.Name(), "name")
		nameBytes, err := os.ReadFile(namePath)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(nameBytes)) == hwmonName {
			sensorPath := filepath.Join(hwmonBase, entry.Name(), "temp1_input")
			if _, err := os.Stat(sensorPath); err != nil {
				return "", fmt.Errorf("sensor file %s not found: %w", sensorPath, err)
			}
			return sensorPath, nil
		}
	}

	return "", fmt.Errorf("CPU temperature sensor %q not found", hwmonName)
}

func ReadTemp(sensorPath string) (float64, error) {
	data, err := os.ReadFile(sensorPath)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", sensorPath, err)
	}

	millideg, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse temperature from %s: %w", sensorPath, err)
	}

	return float64(millideg) / 1000.0, nil
}
