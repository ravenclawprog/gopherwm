package xwindowmgr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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

func (b Binding) Validate() error {
	if strings.TrimSpace(b.Key) == "" {
		return fmt.Errorf("Empty key")
	}
	switch b.Action {
	case "spawn":
		if len(b.Command) == 0 {
			return fmt.Errorf("Action spawn doesn't have a command")
		}
		if strings.TrimSpace(b.Command[0]) == "" {
			return fmt.Errorf("spawn command must not be blank")
		}
	case "toggle_motion":
		if len(b.Command) != 0 {
			return fmt.Errorf("Action toggle_motion doesn't need a command field")
		}
	default:
		return fmt.Errorf("Unknown action %q", b.Action)
	}
	return nil
}

func (c Config) Validate() error {
	for i, b := range c.Bindings {
		if err := b.Validate(); err != nil {
			return fmt.Errorf("binding %d: %w", i+1, err)
		}
	}
	return nil
}
