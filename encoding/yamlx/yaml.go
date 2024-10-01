// Package yamlx contains convenience utilities for yaml.
package yamlx

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Unmarshal deserializes data from YAML to a variable of some type
func Unmarshal[T any](data []byte) (T, error) {
	var rt T
	if err := yaml.Unmarshal(data, &rt); err != nil {
		var zero T
		return zero, err
	}
	return rt, nil
}

// UnmarshalStrict deserializes data from YAML to a variable of some type.
// Fails if there are unknown fields in the data.
func UnmarshalStrict[T any](data []byte) (T, error) {
	var rt T
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&rt); err != nil {
		var zero T
		return zero, err
	}
	return rt, nil
}

// UnmarshalFromFile reads and unmarshals a yaml file.
func UnmarshalFromFile[T any](filename string) (T, error) {
	var rt T

	data, err := os.ReadFile(filename)
	if err != nil {
		return rt, fmt.Errorf("failed to load yaml file %v: %v", filename, err)
	}
	if rt, err = Unmarshal[T](data); err != nil {
		return rt, fmt.Errorf("failed to unmarshal yaml file %v: %v", filename, err)
	}
	return rt, nil
}

// UnmarshalStrictFromFile reads and unmarshals a yaml file.
// Fails if there are unknown fields in the data.
func UnmarshalStrictFromFile[T any](filename string) (T, error) {
	var rt T

	data, err := os.ReadFile(filename)
	if err != nil {
		return rt, fmt.Errorf("failed to load yaml file %v: %v", filename, err)
	}
	if rt, err = UnmarshalStrict[T](data); err != nil {
		return rt, fmt.Errorf("failed to unmarshal yaml file %v: %v", filename, err)
	}
	return rt, nil
}
