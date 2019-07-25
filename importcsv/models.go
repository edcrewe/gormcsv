// Models for loading CSV data  - these can be generated
package importcsv

import (
	// "github.com/jinzhu/gorm"
	// "strings"
	// "time"
)



// GORM Model factory
type ModelFactory struct {
	models []string
}

/*
Make a single instance of the model factory for use in importcsv
 */
func MakeModels() ModelFactory {
	factory := ModelFactory{}
	return factory
}




func (f ModelFactory) New(name string) interface{} {
	return nil
}
