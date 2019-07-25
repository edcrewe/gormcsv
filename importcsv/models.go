// Models for loading CSV data  - these can be generated
package importcsv

import (
	"github.com/jinzhu/gorm"
	"strings"
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
    factory.models = append(factory.models, "country")
    factory.models = append(factory.models, "testtypes")
	return factory
}


type Country struct {
    gorm.Model
	Code       string `gorm:"primary_key"`
	Name       string
	Latitude   float64
	Longtitude float64
	Alias      string
   
}

type TestTypes struct {
   gorm.Model
   datecol string 
   wordcol string 
   codecol string 
   textcol string 
   bigtextcol string 
   numbercol float32 
   intcol int16 
   boolcol bool 
   
}


func (f ModelFactory) New(name string) interface{} {
	name = strings.ToLower(name)
	switch name {
	
	case "country":
		return &Country{}
	
	case "testtypes":
		return &TestTypes{}
	
	}
	return nil
}
