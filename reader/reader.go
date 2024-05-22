package reader

import (
	"crypto/sha256"
	"encoding/hex"
)

type Manifest any

type HypermodeManifest struct {
	Models               []Model               `json:"models"`
	Hosts                []Host                `json:"hosts"`
	EmbeddingSpecs       []EmbeddingSpec       `json:"embeddingSpecs"`
	TrainingInstructions []TrainingInstruction `json:"trainingInstructions"`
	Manifest
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
	data := m.Name + "|" + string(m.Task) + "|" + m.SourceModel + "|" + m.Provider

	// Compute the SHA-256 hash
	hash := sha256.Sum256([]byte(data))

	// Convert the hash to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr
}
