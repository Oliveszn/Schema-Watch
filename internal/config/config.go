package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Oliveszn/Schema-Watch/internal/schema"
)

type Config struct {
	Target          string   `yaml:"target"`
	Port            string   `yaml:"port"`
	IgnoreFields    []string `yaml:"ignore_fields"`
	IgnoreEndpoints []string `yaml:"ignore_endpoints"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) IsEndpointIgnored(endpoint string) bool {
	if c == nil {
		return false
	}
	for _, rule := range c.IgnoreEndpoints {
		if strings.HasSuffix(rule, "*") {
			prefix := strings.TrimSuffix(rule, "*")
			if strings.HasPrefix(endpoint, prefix) {
				return true
			}
			continue
		}
		if endpoint == rule {
			return true
		}
	}
	return false
}

func (c *Config) FilterSchema(s schema.Schema) schema.Schema {
	if c == nil || len(c.IgnoreFields) == 0 {
		return s
	}

	filtered := make(schema.Schema, len(s))
	for path, fieldType := range s {
		if c.isFieldIgnored(path) {
			continue
		}
		filtered[path] = fieldType
	}
	return filtered
}

func (c *Config) isFieldIgnored(path string) bool {
	leaf := path
	if idx := strings.LastIndex(path, "."); idx != -1 {
		leaf = path[idx+1:]
	}
	leaf = strings.TrimSuffix(leaf, "[]")

	for _, ignored := range c.IgnoreFields {
		if path == ignored || leaf == ignored {
			return true
		}
	}
	return false
}
