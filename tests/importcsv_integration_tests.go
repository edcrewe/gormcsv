// +build integration

package tests

import (
	"fmt"
	"github.com/edcrewe/gormcsv/importcsv"
	"testing"
)

var meta importcsv.FieldMeta
/*
Setup first then run the tests
*/
func TestMain(m *testing.M) {
	meta = importcsv.FieldMeta{}
	factory := importcsv.MakeModels()
	model := factory.New("country")
	meta.Setmeta(model, "name,code,latitude,longtitude,alias")
	m.Run()
}