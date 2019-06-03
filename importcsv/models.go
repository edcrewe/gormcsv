// Models for loading CSV data  - these can be generated
package importcsv

import (
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

// GORM Model factory
type ModelFactory struct {
	models []string
}

/*
Make a single instance of the model factory for use in importcsv
 */
func MakeModels() ModelFactory {
	return ModelFactory{
		[]string{"country", "unitofmeasure", "organisation", "item"},
	}
}

// ISO country (location) codes.
type Country struct {
	gorm.Model
	Code       string
	Name       string
	Latitude   float64
	Longtitude float64
	Alias      string
}

type UnitOfMeasure struct {
	gorm.Model
	Name string
}

type Organisation struct {
	gorm.Model
	Name string
}

type Item struct {
	gorm.Model
	Type           int32
	CodeShare      string
	CodeOrg        string
	Description    string
	Quantity       int32
	Uom            UnitOfMeasure
	UomID          int
	Organisation   Organisation
	OrganisationID int
	Status         string
	Date           *time.Time
	Country        Country
	CountryID      int
}

/*
Factory New method to create a model given its name
*/
func (f ModelFactory) New(name string) interface{} {
	name = strings.ToLower(name)
	switch name {
	case "country":
		return &Country{}
	case "unitofmeasure":
		return &UnitOfMeasure{}
	case "organisation":
		return &Organisation{}
	case "item":
		return &Item{}
	}
	return nil
}
