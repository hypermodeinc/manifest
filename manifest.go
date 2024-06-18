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

	v1_manifest "github.com/hypermodeAI/manifest/compat/v1"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/tailscale/hujson"
)

// This version should only be incremented if there are major breaking changes
// to the manifest schema.  In general, we don't consider the schema to be versioned,
// from the user's perspective, so this should be rare.
// NOTE: We intentionally do not expose the *current* version number outside this package.
const currentVersion = 2

// for backward compatibility
const V1AuthHeaderVariableName = "__V1_AUTH_HEADER_VALUE__"

//go:embed hypermode.json
var schemaContent string

type HypermodeManifest struct {
	Version     int                       `json:"-"`
	Models      map[string]ModelInfo      `json:"models"`
	Hosts       map[string]HostInfo       `json:"hosts"`
	Collections map[string]CollectionInfo `json:"collections"`
}

func (m *HypermodeManifest) IsCurrentVersion() bool {
	return m.Version == currentVersion
}

func IsCurrentVersion(version int) bool {
	return version == currentVersion
}

type ModelInfo struct {
	Name        string `json:"-"`
	SourceModel string `json:"sourceModel"`
	Provider    string `json:"provider"`
	Host        string `json:"host"`
}

type HostInfo struct {
	Name            string            `json:"-"`
	Endpoint        string            `json:"endpoint"`
	BaseURL         string            `json:"baseURL"`
	Headers         map[string]string `json:"headers"`
	QueryParameters map[string]string `json:"queryParameters"`
}

type CollectionInfo struct {
	SearchMethods map[string]SearchMethodInfo `json:"searchMethods"`
}

type SearchMethodInfo struct {
	Embedder string    `json:"embedder"`
	Index    IndexInfo `json:"index"`
}

type IndexInfo struct {
	Type    string      `json:"type"`
	Options OptionsInfo `json:"options"`
}

type OptionsInfo struct {
	EfConstruction int `json:"efConstruction"`
	MaxLevels      int `json:"maxLevels"`
}

func (m ModelInfo) Hash() string {
	// Concatenate the attributes into a single string
	data := m.Name + "|" + m.SourceModel + "|" + m.Provider + "|" + m.Host

	// Compute the SHA-256 hash
	hash := sha256.Sum256([]byte(data))

	// Convert the hash to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr
}

func (h HostInfo) Hash() string {
	// Concatenate the attributes into a single string
	data := h.Name + "|" + h.Endpoint + "|" + h.BaseURL + "|" + fmt.Sprintf("%v", h.Headers) + "|" + fmt.Sprintf("%v", h.QueryParameters)

	// Compute the SHA-256 hash
	hash := sha256.Sum256([]byte(data))

	// Convert the hash to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr
}

func ReadManifest(content []byte) (HypermodeManifest, error) {
	// Create standard JSON before attempting to parse
	var manifest HypermodeManifest
	data, err := standardizeJSON(content)
	if err != nil {
		return manifest, err
	}

	// Try to parse using the current format first
	err = parseManifestJson(data, &manifest)
	if err == nil {
		return manifest, nil
	}

	// Try the older format if that failed
	err = parseManifestJsonV1(data, &manifest)
	if err == nil {
		return manifest, nil
	}

	return manifest, fmt.Errorf("failed to parse manifest: %w", err)

}

func parseManifestJson(data []byte, manifest *HypermodeManifest) error {
	err := json.Unmarshal(data, &manifest)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	manifest.Version = currentVersion

	// Copy map keys to Name fields
	for key, model := range manifest.Models {
		model.Name = key
		manifest.Models[key] = model
	}
	for key, host := range manifest.Hosts {
		host.Name = key
		manifest.Hosts[key] = host
	}

	return nil
}

func parseManifestJsonV1(data []byte, manifest *HypermodeManifest) error {
	// Parse the v1 manifest
	var v1_man v1_manifest.HypermodeManifest
	err := json.Unmarshal(data, &v1_man)
	if err != nil {
		return err
	}

	manifest.Version = 1

	// Copy the v1 models to the current structure.
	manifest.Models = make(map[string]ModelInfo, len(v1_man.Models))
	for _, model := range v1_man.Models {
		manifest.Models[model.Name] = ModelInfo{
			Name:        model.Name,
			SourceModel: model.SourceModel,
			Provider:    model.Provider,
			Host:        model.Host,
		}
	}

	// Copy the v1 hosts to the current structure.
	manifest.Hosts = make(map[string]HostInfo, len(v1_man.Hosts))
	for _, host := range v1_man.Hosts {
		h := HostInfo{
			Name: host.Name,
			// In v1 the endpoint was used for both endpoint and baseURL purposes.
			// We'll retain that behavior here so the usage doesn't need to change in the Runtime.
			Endpoint: host.Endpoint,
			BaseURL:  host.Endpoint,
		}
		if host.AuthHeader != "" {
			h.Headers = map[string]string{
				// Use a special variable name for the old auth header value.
				// The runtime will replace this with the old auth header secret value if it exists.
				host.AuthHeader: "{{" + V1AuthHeaderVariableName + "}}",
			}
		}
		manifest.Hosts[host.Name] = h
	}

	return nil
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

var templateRegex = regexp.MustCompile(`{{\s*(?:base64\((.+?):(.+?)\)|(.+?))\s*}}`)

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
