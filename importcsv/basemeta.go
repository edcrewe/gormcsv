// Base class equivalent = embedded Meta for FieldMeta and CSVMeta
package importcsv

import (
"fmt"
"strconv"
"strings"
"time"
)

// Embedded stuct to hang the convert method on
type Meta struct {
}

/*
Type converter for CSV string fields to correct basic type or time
*/
func (base *Meta) Convert(value string, to string) (interface{}, error) {
	if to == "string" {
		return value, nil
	}
	bits := 0
	var err error
	prefixes := []string{"float", "uint", "int"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(to, prefix){
			bitstr := to[len(prefix):]
			if bitstr != "" {
				bits, err = strconv.Atoi(bitstr)
			}
			if err != nil {
				return nil, err
			}
			to = prefix
		}
	}
	switch to {
	case "float":
		return strconv.ParseFloat(value, bits)
	case "uint":
		return strconv.ParseUint(value, 10, bits)
	case "int":
		return strconv.ParseInt(value, 10, bits)
	case "bool":
		return strconv.ParseBool(value)
	case "date":
		var err error
		layouts := []string{"2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05", "28/02/2003"}
		var date time.Time
		for _, layout := range layouts {
			date, err = time.Parse(layout, value)
			if err != nil {
				continue
			}
			return date, nil
		}
		if err != nil {
			return nil, fmt.Errorf("Could not convert time %s to %s, unknown date fmt", value, to)
		}

	}
	return nil, fmt.Errorf("Could not convert %s to %s", value, to)
}
