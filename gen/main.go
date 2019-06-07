// A models.go code generator, see templates docs at https://golang.org/pkg/text/template/
package main

import (
	"os"
	"text/template"
)

type field struct {
	Name string
	Type string
	Tag string
}

type data struct {
	Models map[string]string
	Fields map[string][]field
}

func main() {
	var d data
	d.Models = map[string]string{}
	d.Models["country"] = "Country"
	d.Models["unitofmeasure"] = "UnitOfMeasure"
	d.Models["organisation"] = "Organisation"
	d.Models["item"] = "Item"
	/* d.Fields["Country"] = {field}
	d.Fields["UnitOfMeasure"] = []field{}
	d.Fields["Organisation"] = []field{}
	d.Fields["Item"] = []field{}
	{{range $k, $v := .Fields}}
	struct {{$k}} {
	}
	{{end}}

	*/
	t := template.Must(template.New("models").Parse(modelsTemplate))
	t.Execute(os.Stdout, d)
}

var modelsTemplate = `
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
	factory := ModelFactory{}
	{{range $k, $v := .Models}}
    factory.models = append(factory.models, "{{$k}}")
    {{end}}
	return factory
	}
}

func (f ModelFactory) New(name string) interface{} {
	name = strings.ToLower(name)
	switch name {
	{{range $k, $v := .Models}}
	case "{{$k}}":
		return &{{$v}}()
	{{end}}
	}
	return nil
}
`