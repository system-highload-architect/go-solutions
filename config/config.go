// Package config provides a simple way to load configuration from YAML files
// and override fields with environment variables using a configurable prefix.
//
// Example usage:
//
//	type AppConfig struct {
//	    Server   ServerConfig   `yaml:"server"`
//	    Database DatabaseConfig `yaml:"database"`
//	}
//	type ServerConfig struct {
//	    Port int `yaml:"port" env:"SERVER_PORT"`
//	}
//	type DatabaseConfig struct {
//	    DSN string `yaml:"dsn" env:"DATABASE_DSN"`
//	}
//
//	var cfg AppConfig
//	if err := config.Load(&cfg, config.WithPath("config.yaml"), config.WithEnvPrefix("MYAPP_")); err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(cfg.Server.Port) // from YAML or MYAPP_SERVER_PORT
package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Option is a functional option for Load.
type Option func(*loader)

// WithPath sets the path to the YAML configuration file.
// Default is "config.yaml".
func WithPath(path string) Option {
	return func(l *loader) {
		l.path = path
	}
}

// WithEnvPrefix sets the prefix for environment variable overrides.
// Default is "" (no prefix).
func WithEnvPrefix(prefix string) Option {
	return func(l *loader) {
		l.envPrefix = prefix
	}
}

// Load reads a YAML file and then overrides fields with environment variables.
// cfg must be a non‑nil pointer to a struct.
func Load(cfg interface{}, opts ...Option) error {
	l := &loader{
		path:      "config.yaml",
		envPrefix: "",
	}
	for _, opt := range opts {
		opt(l)
	}

	// 1. Read YAML
	data, err := os.ReadFile(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			// File not found is acceptable; we'll rely only on env vars.
			data = nil
		} else {
			return fmt.Errorf("config: cannot read file %s: %w", l.path, err)
		}
	}
	if data != nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("config: cannot parse YAML: %w", err)
		}
	}

	// 2. Override with environment variables
	return l.overrideEnv(cfg)
}

type loader struct {
	path      string
	envPrefix string
}

// overrideEnv walks through the struct and applies environment variables.
func (l *loader) overrideEnv(cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("config: expected a non‑nil pointer to struct, got %T", cfg)
	}
	return l.overrideStruct(v.Elem(), "")
}

func (l *loader) overrideStruct(v reflect.Value, yamlPath string) error {
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Build the full YAML path (e.g., "server.port")
		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" {
			yamlTag = strings.ToLower(fieldType.Name)
		}
		currentPath := yamlTag
		if yamlPath != "" {
			currentPath = yamlPath + "." + yamlTag
		}

		// If the field is a struct, recurse
		if field.Kind() == reflect.Struct {
			if err := l.overrideStruct(field, currentPath); err != nil {
				return err
			}
			continue
		}

		// Check for "env" tag
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}
		envName := l.envPrefix + envTag
		envValue, exists := os.LookupEnv(envName)
		if !exists {
			continue
		}

		// Set the field from the environment variable
		if err := setField(field, envValue); err != nil {
			return fmt.Errorf("config: cannot set %s from env %s: %w", currentPath, envName, err)
		}
	}
	return nil
}

func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
		} else {
			iv, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(iv)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uv, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uv)
	case reflect.Float32, reflect.Float64:
		fv, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(fv)
	case reflect.Bool:
		bv, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(bv)
	case reflect.Slice:
		// Support for string slices: a,b,c
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
			for i, p := range parts {
				slice.Index(i).SetString(strings.TrimSpace(p))
			}
			field.Set(slice)
		}
	default:
		return fmt.Errorf("unsupported type %s", field.Type())
	}
	return nil
}
