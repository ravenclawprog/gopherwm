package xwindowmgr_test

import (
	"fmt"
	wm "gopherwm/xwindowmgr"
	"os"
	"reflect"
	"testing"
)

func TestConfigSuccessfull(t *testing.T) {
	successfulJson := `
{
  "bindings": [
    {
      "key": "Return",
      "modifiers": ["Super"],
      "action": "spawn",
      "command": ["xterm"]
    }
  ]
}
`
	configPath := t.TempDir() + "/config.json"
	err := os.WriteFile(configPath, []byte(successfulJson), 0644)
	if err != nil {
		t.Fatal("Error at saving json file:", err)
	}
	config, err := wm.LoadConfig(configPath)
	if err != nil {
		t.Fatal("want: no error, but got:", err)
	}
	wantConfig := wm.Config{
		Bindings: []wm.Binding{
			{
				Key:       "Return",
				Modifiers: []string{"Super"},
				Action:    "spawn",
				Command:   []string{"xterm"},
			},
		},
	}
	if !reflect.DeepEqual(config, wantConfig) {
		t.Fatalf("config mismatch:\ngot:  %+v\nwant: %+v", config, wantConfig)
	}
}

func TestConfigSuccessfullWithDefaults(t *testing.T) {
	configPath := ""
	config, err := wm.LoadConfig(configPath)
	if err != nil {
		t.Fatal("want: no error, but got:", err)
	}
	wantConfig := wm.DefaultConfig()
	if !reflect.DeepEqual(config, wantConfig) {
		t.Fatalf("config mismatch:\ngot:  %+v\nwant: %+v", config, wantConfig)
	}
}

func TestConfigReadError(t *testing.T) {
	_, err := wm.LoadConfig(t.TempDir())
	if err == nil {
		t.Fatal("Expected read error, but got nothing")
	}
	fmt.Println("Want:read error, got:", err)
}

func TestConfigParseErrorUnknownField(t *testing.T) {
	unsuccessfulJson := `
{
  "binding": [
    {
      "key": "Return",
      "modifiers": ["Super"],
      "action": "spawn",
      "command": ["xterm"]
    }
  ]
}
`
	configPath := t.TempDir() + "/config.json"
	err := os.WriteFile(configPath, []byte(unsuccessfulJson), 0644)
	if err != nil {
		t.Fatal("Error at saving json file:", err)
	}
	_, err = wm.LoadConfig(configPath)
	if err == nil {
		t.Fatal("want: error, but got: nothing")
	}
	fmt.Println("want:parse error, got:", err)
}

func TestConfigParseErrorSingleValue(t *testing.T) {
	unsuccessfulJson := `
{
  "bindings": []
}
  {}
`
	configPath := t.TempDir() + "/config.json"
	err := os.WriteFile(configPath, []byte(unsuccessfulJson), 0644)
	if err != nil {
		t.Fatal("Error at saving json file:", err)
	}
	_, err = wm.LoadConfig(configPath)
	if err == nil {
		t.Fatal("want: error, but got: nothing")
	}
	fmt.Println("want:parse error, got:", err)
}

func TestConfigParseIncomplete(t *testing.T) {
	unsuccessfulJson := `
{
  "bindings":
`
	configPath := t.TempDir() + "/config.json"
	err := os.WriteFile(configPath, []byte(unsuccessfulJson), 0644)
	if err != nil {
		t.Fatal("Error at saving json file:", err)
	}
	_, err = wm.LoadConfig(configPath)
	if err == nil {
		t.Fatal("want: error, but got: nothing")
	}
	fmt.Println("want:parse error, got:", err)
}

func TestBindingValidateSpawn(t *testing.T) {
	var wantBinding wm.Binding = wm.Binding{
		Key:       "Return",
		Modifiers: []string{"Super"},
		Action:    "spawn",
		Command:   []string{"xterm"},
	}
	err := wantBinding.Validate()
	if err != nil {
		t.Fatal("Validation fail, want: no error, but got:", err)
	}
}

func TestBindingValidateToggleMotion(t *testing.T) {
	var wantBinding wm.Binding = wm.Binding{
		Key:       "m",
		Modifiers: []string{"Super"},
		Action:    "toggle_motion",
	}
	err := wantBinding.Validate()
	if err != nil {
		t.Fatal("Validation fail, want: no error, but got:", err)
	}
}

func TestBindingValidateKeyEmptyFail(t *testing.T) {
	var wantBinding wm.Binding = wm.Binding{
		Key:       "",
		Modifiers: []string{"Super"},
		Action:    "spawn",
		Command:   []string{"xterm"},
	}
	if err := wantBinding.Validate(); err == nil {
		t.Fatal("Validation fail, want: err, but got: noting")
	}
}

func TestBindingValidateKeySpacesFail(t *testing.T) {
	var wantBinding wm.Binding = wm.Binding{
		Key:       "  \t",
		Modifiers: []string{"Super"},
		Action:    "spawn",
		Command:   []string{"xterm"},
	}
	if err := wantBinding.Validate(); err == nil {
		t.Fatal("Validation fail, want: err, but got: noting")
	}
}

func TestBindingValidateActionSpawnFail(t *testing.T) {
	var wantBinding wm.Binding = wm.Binding{
		Key:       "Return",
		Modifiers: []string{"Super"},
		Action:    "spawn",
		Command:   []string{},
	}
	if err := wantBinding.Validate(); err == nil {
		t.Fatal("Validation fail, want: err, but got: noting")
	}
}
