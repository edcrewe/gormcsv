// Models for loading CSV data  - generated 2019-08-15 15:47:17.034352 +0100 BST m=+0.002555724
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
   Latitude float64
   Have int8
   That int8
   Longitude float64
   In int8
   Our int8
   Name string
   Code string
   Other int8
   Data int8
   We int8
   Dont int8
   Alias string
   Some int8

}

type TestTypes struct {
   gorm.Model
   Intcol int16
   Boolcol bool
   Datecol string
   Wordcol string
   Codecol string
   Textcol string
   Bigtextcol string
   Numbercol float32

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
