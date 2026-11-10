// Package config parses shared environment settings and app-owned configuration.
package config

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

// Load parses cfg in place. The caller owns defaults, warnings and caching.
func Load[T any](cfg *T, shared func(*T) *Shared, validate func(*T) []string) (warnings []string, err error) {
	if cfg == nil || shared == nil {
		return nil, errors.New("configuration and shared accessor are required")
	}
	common := shared(cfg)
	if common == nil {
		return nil, errors.New("shared configuration is required")
	}
	var problems []string
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		warnings = append(warnings, fmt.Sprintf("could not load .env file: %v", err))
	}
	if err := env.Parse(cfg); err != nil {
		problems = append(problems, fmt.Sprintf("environment parse error: %v", err))
	}
	problems = append(problems, common.Validate()...)
	if validate != nil {
		problems = append(problems, validate(cfg)...)
	}
	if len(problems) > 0 {
		return warnings, fmt.Errorf("invalid configuration: %s", strings.Join(problems, "; "))
	}
	return warnings, nil
}
