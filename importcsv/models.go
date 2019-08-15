// Models for loading CSV data  - generated 2019-08-15 11:15:06.552148 +0100 BST m=+0.002498526
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
   code string 
   longitude int8 
   we int8 
   name string 
   alias string 
   some int8 
   other int8 
   have int8 
   our int8 
   latitude float32 
   data int8 
   that int8 
   dont int8 
   in int8 
   model int8 
   
}

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
