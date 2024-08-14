/*
 * Copyright 2024 Hypermode, Inc.
 */

package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	HostTypeDgraphCloud string = "dgraph-cloud"
)

type DgraphCloudHostInfo struct {
	Name     string `json:"-"`
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
	Key      string `json:"key"`
}

func (p DgraphCloudHostInfo) HostName() string {
	return p.Name
}

func (DgraphCloudHostInfo) HostType() string {
	return HostTypeDgraphCloud
}

func (h DgraphCloudHostInfo) GetVariables() []string {
	return extractVariables(h.Key)
}

func (h DgraphCloudHostInfo) Hash() string {
	// Concatenate the attributes into a single string
	data := fmt.Sprintf("%v|%v|%v|%v", h.Name, h.Type, h.Endpoint, h.Key)

	// Compute the SHA-256 hash
	hash := sha256.Sum256([]byte(data))

	// Convert the hash to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr
}
