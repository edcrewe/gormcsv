package meta

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// PopulateMeta inspects each CSV file while preserving header order.
func (csvMeta *CSVMeta) PopulateMeta(path string) error {
	files, err := CSVFiles(path)
	if err != nil {
		return err
	}
	csvMeta.Models = make(map[string]string, len(files))
	csvMeta.Fields = make(map[string][]Field, len(files))
	csvMeta.UsesTime = false

	for _, input := range files {
		if err := csvMeta.populateFile(input); err != nil {
			return err
		}
	}
	return nil
}

func (csvMeta *CSVMeta) populateFile(input CSVFile) error {
	file, err := os.Open(input.Path) // #nosec G304 -- path is explicitly supplied by the user
	if err != nil {
		return fmt.Errorf("open %q: %w", input.Path, err)
	}
	defer func() { _ = file.Close() }()

	reader := NewCSVReader(file)
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header from %q: %w", input.Path, err)
	}
	if len(header) == 0 {
		return fmt.Errorf("CSV header in %q is empty", input.Path)
	}
	samples := make([]inference, len(header))
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("inspect %q: %w", input.Path, err)
		}
		for column, value := range record {
			samples[column].add(value)
		}
	}

	modelName := exportedName(input.Model)
	if modelName == "" {
		return fmt.Errorf("filename %q does not produce a valid Go model name", input.Path)
	}
	csvMeta.Models[strings.ToLower(input.Model)] = modelName
	fields := make([]Field, 0, len(header))
	seen := make(map[string]struct{}, len(header))
	for column, heading := range header {
		name := exportedName(heading)
		if name == "" || name == "Model" {
			continue
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("CSV header %q produces duplicate Go field %q", heading, name)
		}
		seen[name] = struct{}{}
		field := samples[column].field(name)
		if field.Type == "time.Time" {
			csvMeta.UsesTime = true
		}
		fields = append(fields, field)
	}
	csvMeta.Fields[modelName] = fields
	return nil
}

func exportedName(value string) string {
	parts := strings.FieldsFunc(strings.TrimSpace(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	var result strings.Builder
	for _, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		result.WriteRune(unicode.ToUpper(runes[0]))
		result.WriteString(string(runes[1:]))
	}
	name := result.String()
	if name != "" && unicode.IsDigit([]rune(name)[0]) {
		name = "Field" + name
	}
	return name
}

// GetField infers a Go type from sampled strings.
func (csvMeta *CSVMeta) GetField(name string, values []string) Field {
	var sample inference
	for _, value := range values {
		sample.add(value)
	}
	return sample.field(name)
}

type inference struct {
	count       int
	allBool     bool
	allTime     bool
	allSigned   bool
	allUnsigned bool
	allFloat    bool
	minSigned   int64
	maxSigned   int64
	maxUnsigned uint64
}

func (sample *inference) add(value string) {
	if value == "" {
		return
	}
	if sample.count == 0 {
		sample.allBool = true
		sample.allTime = true
		sample.allSigned = true
		sample.allUnsigned = true
		sample.allFloat = true
	}
	sample.count++
	if _, err := strconv.ParseBool(value); err != nil {
		sample.allBool = false
	}
	if _, err := parseTime(value); err != nil {
		sample.allTime = false
	}
	if parsed, err := strconv.ParseInt(value, 10, 64); err != nil {
		sample.allSigned = false
	} else if sample.count == 1 {
		sample.minSigned, sample.maxSigned = parsed, parsed
	} else {
		sample.minSigned = min(sample.minSigned, parsed)
		sample.maxSigned = max(sample.maxSigned, parsed)
	}
	if parsed, err := strconv.ParseUint(value, 10, 64); err != nil {
		sample.allUnsigned = false
	} else {
		sample.maxUnsigned = max(sample.maxUnsigned, parsed)
	}
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		sample.allFloat = false
	}
}

func (sample inference) field(name string) Field {
	if sample.count == 0 {
		return Field{Name: name, Type: "string"}
	}
	if sample.allBool {
		return Field{Name: name, Type: "bool"}
	}
	if sample.allTime {
		return Field{Name: name, Type: "time.Time"}
	}
	if sample.allSigned {
		return Field{Name: name, Type: signedType(sample.minSigned, sample.maxSigned)}
	}
	if sample.allUnsigned {
		return Field{Name: name, Type: unsignedType(sample.maxUnsigned)}
	}
	if sample.allFloat {
		return Field{Name: name, Type: "float64"}
	}
	return Field{Name: name, Type: "string"}
}

func signedType(minimum, maximum int64) string {
	switch {
	case minimum >= math.MinInt8 && maximum <= math.MaxInt8:
		return "int8"
	case minimum >= math.MinInt16 && maximum <= math.MaxInt16:
		return "int16"
	case minimum >= math.MinInt32 && maximum <= math.MaxInt32:
		return "int32"
	default:
		return "int64"
	}
}

func unsignedType(maximum uint64) string {
	switch {
	case maximum <= math.MaxUint8:
		return "uint8"
	case maximum <= math.MaxUint16:
		return "uint16"
	case maximum <= math.MaxUint32:
		return "uint32"
	default:
		return "uint64"
	}
}
