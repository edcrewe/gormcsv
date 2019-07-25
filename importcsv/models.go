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
   dont int8 
   have int8 
   in int8 
   model int8 
   other int8 
   code string 
   longitude int8 
   alias string 
   some int8 
   that int8 
   we int8 
   name string 
   our int8 
   latitude float32 
   data int8 
   
}

//  TODO - get $k value from .Models - $name := (index $.Models $k).Val
type TestTypes struct {
    gorm.Model
   boolcol bool 
   datecol string 
   wordcol string 
   codecol string 
   textcol string 
   bigtextcol string 
   numbercol float32 
   intcol int16 
   
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
