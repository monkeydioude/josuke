package josuke

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	yaml "github.com/goccy/go-yaml"
)

func parseConfig(configFilePath string) (*Josuke, error) {
	file, err := os.ReadFile(configFilePath)

	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	j := &Josuke{
		LogLevelName: "INFO",
		Host:         "localhost",
		Port:         8082,
		Hooks:        make([]*Hook, 0),
	}

	if err := json.Unmarshal(file, j); err != nil {
		if err := yaml.Unmarshal(file, j); err != nil {
			return nil, errors.New("could not parse yaml/json from config file")
		}
	}
	return j, err
}
