/*
 * Copyright 2024 Hypermode, Inc.
 */

package manifest

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/tailscale/hujson"
)

//go:embed hypermode.json
var schemaContent string

type HypermodeManifest struct {
	Models map[string]ModelInfo `json:"models"`
	Hosts  map[string]HostInfo  `json:"hosts"`
}

type ModelTask string

const (
	ClassificationTask ModelTask = "classification"
	EmbeddingTask      ModelTask = "embedding"
	GenerationTask     ModelTask = "generation"
)

type ModelInfo struct {
	Name        string    `json:"-"`
	Task        ModelTask `json:"task"`
	SourceModel string    `json:"sourceModel"`
	Provider    string    `json:"provider"`
	Host        string    `json:"host"`
}

type HostInfo struct {
	Name            string            `json:"-"`
	Endpoint        string            `json:"endpoint"`
	BaseURL         string            `json:"baseURL"`
	Headers         map[string]string `json:"headers"`
	QueryParameters map[string]string `json:"queryParameters"`
}

func (m ModelInfo) Hash() string {
	// Concatenate the attributes into a single string
	data := m.Name + "|" + string(m.Task) + "|" + m.SourceModel + "|" + m.Provider + "|" + m.Host

	// Compute the SHA-256 hash
	hash := sha256.Sum256([]byte(data))

	// Convert the hash to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr
}

func ReadManifest(content []byte) (HypermodeManifest, error) {

	data, err := standardizeJSON(content)
	if err != nil {
		return HypermodeManifest{}, err
	}

	// Parse the JSON
	manifest := HypermodeManifest{}
	err = json.Unmarshal(data, &manifest)

	// Copy map keys to Name fields
	for key, model := range manifest.Models {
		model.Name = key
		manifest.Models[key] = model
	}
	for key, host := range manifest.Hosts {
		host.Name = key
		manifest.Hosts[key] = host
	}

	return manifest, err
}

// standardizeJSON removes comments and trailing commas to make the JSON valid
func standardizeJSON(b []byte) ([]byte, error) {
	ast, err := hujson.Parse(b)
	if err != nil {
		return b, err
	}
	ast.Standardize()
	return ast.Pack(), nil
}

func ValidateManifest(content []byte) error {
	sch, err := jsonschema.CompileString("hypermode.json", schemaContent)
	if err != nil {
		return err
	}

	content, err = standardizeJSON(content)
	if err != nil {
		return fmt.Errorf("failed to standardize manifest: %w", err)
	}

	var v interface{}
	err = json.Unmarshal(content, &v)
	if err != nil {
		return fmt.Errorf("failed to deserialize manifest: %w", err)
	}

	err = sch.Validate(v)
	if err != nil {
		return fmt.Errorf("failed to validate manifest: %w", err)
	}

	return nil
}

func (m *HypermodeManifest) GetHostVariables() map[string][]string {
	results := make(map[string][]string, len(m.Hosts))

	for _, host := range m.Hosts {
		vars := host.GetVariables()
		if len(vars) > 0 {
			results[host.Name] = vars
		}
	}

	return results
}

func (h *HostInfo) GetVariables() []string {
	cap := 2 * (len(h.Headers) + len(h.QueryParameters))
	set := make(map[string]bool, cap)
	results := make([]string, 0, cap)

	for _, header := range h.Headers {
		vars := extractVariables(header)
		for _, v := range vars {
			if _, ok := set[v]; !ok {
				set[v] = true
				results = append(results, v)
			}
		}
	}

	for _, v := range h.QueryParameters {
		vars := extractVariables(v)
		for _, v := range vars {
			if _, ok := set[v]; !ok {
				set[v] = true
				results = append(results, v)
			}
		}
	}

	return results
}

var templateRegex = regexp.MustCompile(`{{\s*(?:base64:\((.+?):(.+?)\)|(.+?))\s*}}`)

func extractVariables(s string) []string {
	matches := templateRegex.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return []string{}
	}

	results := make([]string, 0, len(matches)*2)
	for _, match := range matches {
		for j := 1; j < len(match); j++ {
			if match[j] != "" {
				results = append(results, match[j])
			}
		}
	}

	return results
}
