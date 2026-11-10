package config

import (
	"fmt"
	"regexp"
)

type Database struct {
	Type     string `env:"TYPE"`
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"5432"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	Name     string `env:"NAME"`
	SSLMode  string `env:"SSLMODE" envDefault:"disable"`
	Schema   string `env:"SCHEMA" envDefault:"public"`
}

var schemaPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

func (db *Database) Validate() []string {
	var problems []string
	if db.Schema == "" {
		db.Schema = "public"
	}
	if !schemaPattern.MatchString(db.Schema) {
		problems = append(problems, "DATABASE_SCHEMA must match ^[a-z_][a-z0-9_]*$")
	}
	if db.Type != "" && db.Type != "postgres" {
		problems = append(problems, fmt.Sprintf("DATABASE_TYPE must be postgres or empty (got %q)", db.Type))
	}
	if db.Type == "postgres" {
		for _, required := range []struct{ name, value string }{{"DATABASE_HOST", db.Host}, {"DATABASE_USER", db.User}, {"DATABASE_NAME", db.Name}} {
			if required.value == "" {
				problems = append(problems, fmt.Sprintf("%s is required when DATABASE_TYPE=postgres", required.name))
			}
		}
		if db.Port < 1 || db.Port > 65535 {
			problems = append(problems, "DATABASE_PORT must be between 1 and 65535")
		}
	}
	return problems
}
