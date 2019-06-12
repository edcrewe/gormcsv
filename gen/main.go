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
.Models = map[string]string{}
func main() {
	var d data
	d
	d.Models["country"] = "Country"
	d.Models["unitofmeasure"] = "UnitOfMeasure"
	d.Models["organisation"] = "Organisation"
	d.Models["item"] = "Item"
	d.Fields = map[string][]field{}
	d.Fields["Country"] = []field{
		field{"Code","string", `gorm:"primary_key"`},
	}
	d.Fields["UnitOfMeasure"] = []field{
		field{"Name","string", ""},
	}
	d.Fields["Organisation"] = []field{
		field{"Name","string", ""},
	}
	d.Fields["Item"] = []field{}
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

{{range $k, $v := .Fields}}
struct {{$k}} {
   {{ range $f := $v }}
   {{$f.Name}} {{$f.Type}} {{$f.Tag}}
   {{end}}
}
{{end}}

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