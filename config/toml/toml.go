package toml

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"github.com/pelletier/go-toml/v2"
)

// Load reads a TOML file from path into T, applying struct tag defaults first.
func Load[T any](path string) (*T, error) {
	var t T

	if err := defaults.Set(&t); err != nil {
		return nil, fmt.Errorf("failed to apply defaults: %w", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %q: %w", path, err)
	}

	if err = toml.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("failed to parse config %q: %w", path, err)
	}

	return &t, nil
}
