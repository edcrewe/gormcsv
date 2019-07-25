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


//  TODO - get $k value from .Models - $name := (index $.Models $k).Val
type Country struct {
    gorm.Model
   code string 
   have int8 
   our int8 
   that int8 
   dont int8 
   in int8 
   name string 
   longitude int8 
   alias string 
   some int8 
   other int8 
   data int8 
   latitude float32 
   we int8 
   model int8 
   
}

//  TODO - get $k value from .Models - $name := (index $.Models $k).Val
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
