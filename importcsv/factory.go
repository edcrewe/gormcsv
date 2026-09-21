package importcsv

import "strings"

// ModelFactory creates fresh model pointers by case-insensitive name.
type ModelFactory struct {
	models map[string]func() any
}

// NewModelFactory creates a factory from model constructors.
func NewModelFactory(models map[string]func() any) ModelFactory {
	normalized := make(map[string]func() any, len(models))
	for name, constructor := range models {
		normalized[strings.ToLower(name)] = constructor
	}
	return ModelFactory{models: normalized}
}

// New returns a fresh model or nil when the name is not registered.
func (factory ModelFactory) New(name string) any {
	constructor := factory.models[strings.ToLower(name)]
	if constructor == nil {
		return nil
	}
	return constructor()
}
