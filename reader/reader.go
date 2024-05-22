/*
 * Copyright 2024 Hypermode, Inc.
 */

package reader

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type HypermodeManifest struct {
	Models []Model `json:"models"`
	Hosts  []Host  `json:"hosts"`
}

type ModelTask string

const (
	ClassificationTask ModelTask = "classification"
	EmbeddingTask      ModelTask = "embedding"
	GenerationTask     ModelTask = "generation"
)

type Model struct {
	Name        string    `json:"name"`
	Task        ModelTask `json:"task"`
	SourceModel string    `json:"sourceModel"`
	Provider    string    `json:"provider"`
	Host        string    `json:"host"`
}

type Host struct {
	Name       string `json:"name"`
	Endpoint   string `json:"endpoint"`
	AuthHeader string `json:"authHeader"`
}

func (m Model) Hash() string {
	// Concatenate the attributes into a single string
	data := m.Name + "|" + string(m.Task) + "|" + m.SourceModel + "|" + m.Provider + "|" + m.Host

	// Compute the SHA-256 hash
	hash := sha256.Sum256([]byte(data))

	// Convert the hash to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr
}

func ReadManifest(content []byte) (manifest HypermodeManifest, err error) {
	err = json.Unmarshal(content, &manifest)
	return
}
