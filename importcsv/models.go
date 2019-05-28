// Models for loading CSV data  - these can be generated
package importcsv

import (
	"github.com/jinzhu/gorm"
	"time"
)

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
