package manifest_test

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/hypermodeAI/manifest"
)

//go:embed valid_hypermode.json
var validManifest []byte

//go:embed old_v1_hypermode.json
var oldV1Manifest []byte

func TestReadManifest(t *testing.T) {
	// This should match the content of valid_hypermode.json
	expectedManifest := manifest.HypermodeManifest{
		Version: 2,
		Models: map[string]manifest.ModelInfo{
			"model-1": {
				Name:        "model-1",
				SourceModel: "example/source-model-1",
				Provider:    "hugging-face",
				Host:        "hypermode",
			},
			"model-2": {
				Name:        "model-2",
				SourceModel: "source-model-2",
				Host:        "my-model-host",
				Path:        "path/to/model-2",
			},
			"model-3": {
				Name:        "model-3",
				SourceModel: "source-model-3",
				Host:        "my-model-host",
			},
		},
		Hosts: map[string]manifest.HostInfo{
			"my-model-host": manifest.HTTPHostInfo{
				Name:    "my-model-host",
				Type:    manifest.HostTypeHTTP,
				BaseURL: "https://models.example.com/",
				Headers: map[string]string{
					"X-API-Key": "{{API_KEY}}",
				},
			},
			"another-model-host": manifest.HTTPHostInfo{
				Name:     "another-model-host",
				Type:     manifest.HostTypeHTTP,
				Endpoint: "https://models.example.com/full/path/to/model-3",
				Headers: map[string]string{
					"X-API-Key": "{{API_KEY}}",
				},
			},
			"my-graphql-api": manifest.HTTPHostInfo{
				Name:     "my-graphql-api",
				Type:     manifest.HostTypeHTTP,
				Endpoint: "https://api.example.com/graphql",
				Headers: map[string]string{
					"Authorization": "Bearer {{AUTH_TOKEN}}",
				},
			},
			"my-rest-api": manifest.HTTPHostInfo{
				Name:    "my-rest-api",
				Type:    manifest.HostTypeHTTP,
				BaseURL: "https://api.example.com/v1/",
				QueryParameters: map[string]string{
					"api_token": "{{API_TOKEN}}",
				},
			},
			"another-rest-api": manifest.HTTPHostInfo{
				Name:    "another-rest-api",
				Type:    manifest.HostTypeHTTP,
				BaseURL: "https://api.example.com/v2/",
				Headers: map[string]string{
					"Authorization": "Basic {{base64(USERNAME:PASSWORD)}}",
				},
			},
			"api-with-type": manifest.HTTPHostInfo{
				Name:    "api-with-type",
				Type:    manifest.HostTypeHTTP,
				BaseURL: "https://api.example.com/v2/",
			},
			"neon": manifest.PostgresqlHostInfo{
				Name:    "neon",
				Type:    "postgresql",
				ConnStr: "postgresql://{{POSTGRESQL_USERNAME}}:{{POSTGRESQL_PASSWORD}}@1.2.3.4:5432/data?sslmode=disable",
			},
		},
		Collections: map[string]manifest.CollectionInfo{
			"collection1": {
				SearchMethods: map[string]manifest.SearchMethodInfo{
					"searchMethod1": {
						Embedder: "embedder1",
					},
					"searchMethod2": {
						Embedder: "embedder1",
						Index: manifest.IndexInfo{
							Type: "hnsw",
							Options: manifest.OptionsInfo{
								EfConstruction: 100,
								MaxLevels:      3,
							},
						},
					},
				},
			},
		},
	}

	actualManifest, err := manifest.ReadManifest(validManifest)
	if err != nil {
		t.Errorf("Error reading manifest: %v", err)
		return
	}

	if !reflect.DeepEqual(actualManifest, expectedManifest) {
		t.Errorf("Expected manifest: %+v, but got: %+v", expectedManifest, actualManifest)
	}
}

func TestReadV1Manifest(t *testing.T) {
	// This should match the content of old_v1_hypermode.json, after translating it to the new structure
	expectedManifest := manifest.HypermodeManifest{
		Version: 1,
		Models: map[string]manifest.ModelInfo{
			"model-1": {
				Name:        "model-1",
				SourceModel: "source-model-1",
				Provider:    "provider-1",
				Host:        "my-model-host",
			},
			"model-2": {
				Name:        "model-2",
				SourceModel: "source-model-2",
				Provider:    "provider-2",
				Host:        "hypermode",
			},
			"model-3": {
				Name:        "model-3",
				SourceModel: "source-model-3",
				Provider:    "provider-3",
				Host:        "hypermode",
			},
		},
		Hosts: map[string]manifest.HostInfo{
			"my-model-host": manifest.HTTPHostInfo{
				Name:     "my-model-host",
				Endpoint: "https://models.example.com/full/path/to/model-1",
				BaseURL:  "https://models.example.com/full/path/to/model-1",
				Headers: map[string]string{
					"X-API-Key": "{{" + manifest.V1AuthHeaderVariableName + "}}",
				},
			},
			"my-graphql-api": manifest.HTTPHostInfo{
				Name:     "my-graphql-api",
				Endpoint: "https://api.example.com/graphql",
				BaseURL:  "https://api.example.com/graphql",
				Headers: map[string]string{
					"Authorization": "{{" + manifest.V1AuthHeaderVariableName + "}}",
				},
			},
		},
	}

	actualManifest, err := manifest.ReadManifest(oldV1Manifest)
	if err != nil {
		t.Errorf("Error reading manifest: %v", err)
		return
	}

	if !reflect.DeepEqual(actualManifest, expectedManifest) {
		t.Errorf("Expected manifest: %+v, but got: %+v", expectedManifest, actualManifest)
	}
}

func TestValidateManifest(t *testing.T) {
	if err := manifest.ValidateManifest(validManifest); err != nil {
		t.Error(err)
	}
}

func TestModelInfo_Hash(t *testing.T) {
	model := manifest.ModelInfo{
		Name:        "my-model",
		SourceModel: "my-source-model",
		Provider:    "my-provider",
		Host:        "my-host",
	}

	expectedHash := "f0e05986e8fc7c7986337990cfd175adc62a323e287a7802f43e60eea77c93ac"

	actualHash := model.Hash()
	if actualHash != expectedHash {
		t.Errorf("Expected hash: %s, but got: %s", expectedHash, actualHash)
	}
}

func TestHttpHostInfo_Hash(t *testing.T) {
	host := manifest.HTTPHostInfo{
		Name:     "my-host",
		Endpoint: "https://example.com/api",
		BaseURL:  "https://example.com/api",
		Headers: map[string]string{
			"Authorization": "Bearer {{API_TOKEN}}",
		},
		QueryParameters: map[string]string{
			"api_token": "{{API_TOKEN}}",
		},
	}

	expectedHash := "897ba7738c819211a9291f402bbdda529aadd4f83107ee08157e72bc12e915ec"

	actualHash := host.Hash()
	if actualHash != expectedHash {
		t.Errorf("Expected hash: %s, but got: %s", expectedHash, actualHash)
	}
}

func TestHPostgresHostInfo_Hash(t *testing.T) {
	host := manifest.PostgresqlHostInfo{
		Name:    "my-database",
		ConnStr: "postgresql://{{USERNAME}}:{{PASSWORD}}@database.example.com:5432/dbname?sslmode=require",
	}

	expectedHash := "bca6ac337c06274506199ceab0f5ce6145b2abd88504284ad0125b8ea1e1d051"

	actualHash := host.Hash()
	if actualHash != expectedHash {
		t.Errorf("Expected hash: %s, but got: %s", expectedHash, actualHash)
	}
}

func TestGetHostVariablesFromManifest(t *testing.T) {
	// This should match the host variables that are present in valid_hypermode.json
	expectedVars := map[string][]string{
		"my-model-host":      {"API_KEY"},
		"another-model-host": {"API_KEY"},
		"my-graphql-api":     {"AUTH_TOKEN"},
		"my-rest-api":        {"API_TOKEN"},
		"another-rest-api":   {"USERNAME", "PASSWORD"},
		"neon":               {"POSTGRESQL_USERNAME", "POSTGRESQL_PASSWORD"},
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
