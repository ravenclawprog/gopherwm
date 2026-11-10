package xwindowmgr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Config struct {
	Bindings []Binding `json:"bindings"`
}

type Binding struct {
	Key       string   `json:"key"`
	Modifiers []string `json:"modifiers"`
	Action    string   `json:"action"`
	Command   []string `json:"command,omitempty"`
}

func DefaultConfig() Config {
	var ret Config = Config{
		Bindings: []Binding{
			{Key: "Return",
				Modifiers: []string{"Super"},
				Action:    "spawn",
				Command:   []string{"xterm"},
			},
			{
				Key:       "m",
				Modifiers: []string{"Super"},
				Action:    "toggle_motion",
			},
		},
	}
	return ret
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultConfig(), nil
	}

	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var config Config

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	errjson := decoder.Decode(&config)
	if errjson != nil {
		return Config{}, fmt.Errorf("parse config: %w", errjson)
	}
	var extra any
	err = decoder.Decode(&extra)
	if err != io.EOF {
		if err != nil {
			return Config{}, fmt.Errorf("parse config: %w", err)
		}
		return Config{}, fmt.Errorf("parse config: expected a single JSON value")
	}

	return config, nil
}
