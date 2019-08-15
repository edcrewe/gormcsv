// Models for loading CSV data  - generated 2019-08-15 15:14:59.205253 +0100 BST m=+0.002567647
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
   Code string 
   Latitude float32 
   We int8 
   Name string 
   Other int8 
   Data int8 
   Have int8 
   That int8 
   Dont int8 
   In int8 
   Our int8 
   Model int8 
   Longitude int8 
   Alias string 
   Some int8 
   
}

type TestTypes struct {
   gorm.Model
   Bigtextcol string 
   Numbercol float32 
   Intcol int16 
   Boolcol bool 
   Datecol string 
   Wordcol string 
   Codecol string 
   Textcol string 
   
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
