/*
 * Copyright 2024 Hypermode, Inc.
 */

package manifest

const (
	HostTypePostgresql string = "postgresql"
)

type PostgresqlHostInfo struct {
	Name    string `json:"-"`
	Type    string `json:"type"`
	ConnStr string `json:"connString"`
}

func (p PostgresqlHostInfo) HostName() string {
	return p.Name
}

func (PostgresqlHostInfo) HostType() string {
	return HostTypePostgresql
}

func (h PostgresqlHostInfo) GetVariables() []string {
	return extractVariables(h.ConnStr)
}
