package clients

import (
	"strings"
	"testing"
)

func TestSetupVaultPaths(t *testing.T) {
	dataPath, metadataPath := setupVaultPaths(1, "mock", "")
	if dataPath != "mock" || metadataPath != "mock" {
		t.Error("Unexpected data paths")
	}

	dataPath, metadataPath = setupVaultPaths(1, "mock", "mockpath")
	if !strings.Contains(dataPath, "mockpath") || !strings.Contains(metadataPath, "mockpath") {
		t.Error("Expected path not found")
	}

	dataPath, metadataPath = setupVaultPaths(2, "mock", "")
	if dataPath != "mock/data" || metadataPath != "mock/metadata" {
		t.Error("Unexpected data paths")
	}

	dataPath, metadataPath = setupVaultPaths(2, "mock", "mockpath")
	if !strings.Contains(dataPath, "mockpath") || !strings.Contains(metadataPath, "mockpath") {
		t.Error("Expected path not found")
	}
}

func TestLoadVaultSecretsAtPath(t *testing.T) {
	c := make(chan *Secret)
	data := map[string]interface{}{
		"keys": []interface{}{"mock1", "mockpath/", "mock2"},
	}

	results := make([]*Secret, 0)
	go func() {
		for secret := range c {
			results = append(results, secret)
		}
	}()

	loadVaultSecretsAtPath(c, "mock", data)
	close(c)

	if len(results) != 2 {
		t.Error("Expected secret cound did not match")
	}
}
