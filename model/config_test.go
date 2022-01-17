package model

import (
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	c := NewConfig()
	if c.BackupCount != 10 {
		t.Errorf("unexpected default values for config")
	}
}

func TestConfigKeys(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		expected []string
	}{
		{"keys", fakeManifest(1, 1), []string{"verbose", "aliasesOnly", "mode", "backupCount"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.manifest.Config.Keys(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestConfigFields(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		expected map[string]interface{}
	}{
		{
			"keys",
			fakeManifest(1, 1),
			map[string]interface{}{
				"verbose":     true,
				"aliasesOnly": false,
				"mode":        "concatenate",
				"backupCount": 10,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.manifest.Config.Fields(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}
