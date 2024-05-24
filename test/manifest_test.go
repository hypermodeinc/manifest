package manifest_test

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/hypermodeAI/manifest"
)

//go:embed valid_hypermode.json
var validManifest []byte

func TestReadManifest(t *testing.T) {
	// This should match the content of valid_hypermode.json
	expectedManifest := manifest.HypermodeManifest{
		Models: map[string]manifest.ModelInfo{
			"model-1": {
				Name:        "model-1",
				Task:        manifest.ClassificationTask,
				SourceModel: "source-model-1",
				Provider:    "provider-1",
				Host:        "my-model-host",
			},
			"model-2": {
				Name:        "model-2",
				Task:        manifest.EmbeddingTask,
				SourceModel: "source-model-2",
				Provider:    "provider-2",
				Host:        "hypermode",
			},
			"model-3": {
				Name:        "model-3",
				Task:        manifest.GenerationTask,
				SourceModel: "source-model-3",
				Provider:    "provider-3",
				Host:        "hypermode",
			},
		},
		Hosts: map[string]manifest.HostInfo{
			"my-model-host": {
				Name:     "my-model-host",
				Endpoint: "https://models.example.com/full/path/to/model-1",
				Headers: map[string]string{
					"X-API-Key": "{{API_KEY}}",
				},
			},
			"my-graphql-api": {
				Name:     "my-graphql-api",
				Endpoint: "https://api.example.com/graphql",
				Headers: map[string]string{
					"Authorization": "Bearer {{AUTH_TOKEN}}",
				},
			},
			"my-rest-api": {
				Name:    "my-rest-api",
				BaseURL: "https://api.example.com/v1/",
				QueryParameters: map[string]string{
					"api_token": "{{API_TOKEN}}",
				},
			},
			"another-rest-api": {
				Name:    "another-rest-api",
				BaseURL: "https://api.example.com/v2/",
				Headers: map[string]string{
					"Authorization": "Basic {{base64:(USERNAME:PASSWORD)}}",
				},
			},
		},
	}

	actualManifest, err := manifest.ReadManifest(validManifest)
	if err != nil {
		t.Errorf("Error reading manifest: %v", err)
	}

	if !reflect.DeepEqual(actualManifest, expectedManifest) {
		t.Errorf("Expected manifest: %+v, but got: %+v", expectedManifest, actualManifest)
	}
}

func TestValidateManifest(t *testing.T) {
	err := manifest.ValidateManifest(validManifest)
	if err != nil {
		t.Error(err)
	}
}

func TestModelInfo_Hash(t *testing.T) {
	model := manifest.ModelInfo{
		Name:        "my-model",
		Task:        manifest.ClassificationTask,
		SourceModel: "my-source-model",
		Provider:    "my-provider",
		Host:        "my-host",
	}

	expectedHash := "c53ed7d572cb8619c4817eb4a66de580754cc52058004e4b0ae0484e19f3e043"

	actualHash := model.Hash()
	if actualHash != expectedHash {
		t.Errorf("Expected hash: %s, but got: %s", expectedHash, actualHash)
	}
}

func TestGetHostVariablesFromManifest(t *testing.T) {

	// This should match the host variables that are present in valid_hypermode.json
	expectedVars := map[string][]string{
		"my-model-host":    {"API_KEY"},
		"my-graphql-api":   {"AUTH_TOKEN"},
		"my-rest-api":      {"API_TOKEN"},
		"another-rest-api": {"USERNAME", "PASSWORD"},
	}

	m, err := manifest.ReadManifest(validManifest)
	if err != nil {
		t.Errorf("Error reading manifest: %v", err)
	}

	vars := m.GetHostVariables()

	if !reflect.DeepEqual(vars, expectedVars) {
		t.Errorf("Expected vars: %+v, but got: %+v", expectedVars, vars)
	}
}
