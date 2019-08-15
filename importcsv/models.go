// Models for loading CSV data  - generated 2019-08-15 10:30:50.292744 +0100 BST m=+0.002527033
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
   longitude int8 
   data int8 
   we int8 
   dont int8 
   have int8 
   in int8 
   name string 
   latitude float32 
   our int8 
   model int8 
   some int8 
   alias string 
   that int8 
   code string 
   other int8 
   
}

type TestTypes struct {
   gorm.Model
   codecol string 
   textcol string 
   bigtextcol string 
   numbercol float32 
   intcol int16 
   boolcol bool 
   datecol string 
   wordcol string 
   
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
