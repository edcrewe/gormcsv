package meta

import (
	"fmt"
	"reflect"
	"strings"
)

// NewFieldMeta validates a CSV header against the exported fields of model.
// The special "model" column is ignored for compatibility with generated fixtures.
func NewFieldMeta(model any, header []string) (*FieldMeta, error) {
	modelType, err := structType(model)
	if err != nil {
		return nil, err
	}
	if len(header) == 0 {
		return nil, fmt.Errorf("CSV header is empty")
	}

	fields := make([]mappedField, 0, len(header))
	seen := make(map[string]struct{}, len(header))
	for column, heading := range header {
		heading = strings.TrimSpace(heading)
		if heading == "" {
			return nil, fmt.Errorf("CSV column %d has an empty heading", column+1)
		}
		key := strings.ToLower(heading)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("CSV heading %q is duplicated", heading)
		}
		seen[key] = struct{}{}
		if strings.EqualFold(heading, "model") {
			continue
		}
		field, ok := modelType.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, heading)
		})
		if !ok || field.PkgPath != "" {
			return nil, fmt.Errorf("CSV heading %q has no exported field in %s", heading, modelType)
		}
		fields = append(fields, mappedField{column: column, name: field.Name})
	}
	return &FieldMeta{fields: fields}, nil
}

// RecordToModel populates a new model using the mapping compiled from its header.
func (meta *FieldMeta) RecordToModel(model any, record []string) (any, error) {
	structValue, err := structValue(model)
	if err != nil {
		return nil, err
	}
	for _, field := range meta.fields {
		if field.column >= len(record) {
			return nil, fmt.Errorf("record has %d columns, expected at least %d", len(record), field.column+1)
		}
		target := structValue.FieldByName(field.name)
		if !target.IsValid() || !target.CanSet() {
			return nil, fmt.Errorf("field %s cannot be set", field.name)
		}
		value, err := convert(record[field.column], target.Type())
		if err != nil {
			return nil, fmt.Errorf("column %q: %w", field.name, err)
		}
		target.Set(value)
	}
	return model, nil
}

func structType(model any) (reflect.Type, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}
	modelType := reflect.TypeOf(model)
	if modelType.Kind() != reflect.Pointer || modelType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a pointer to a struct, got %T", model)
	}
	return modelType.Elem(), nil
}

func structValue(model any) (reflect.Value, error) {
	if _, err := structType(model); err != nil {
		return reflect.Value{}, err
	}
	value := reflect.ValueOf(model)
	if value.IsNil() {
		return reflect.Value{}, fmt.Errorf("model is a nil pointer")
	}
	return value.Elem(), nil
}
