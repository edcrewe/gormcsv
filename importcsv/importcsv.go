package importcsv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
)

type ModelCSV struct {
	structs map[string]reflect.Type
}

/*
Init by finding all the models and loading them to a map
 */
func (mcsv *ModelCSV) init() {
	mcsv.structs = make(map[string]reflect.Type)
	mcsv.structs["country"] = reflect.TypeOf(Country{})
	// &UnitOfMeasure{}, &Organisation{}, &Item{}}

}

/*
Connect to the Database
 */
func (mcsv *ModelCSV) connectDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal("failed to connect database")
	}
	return db
}

/*
Migrate the schema
 */
func (mcsv *ModelCSV) createSchema(db *gorm.DB) {
	for _, model := range mcsv.structs {
		db.AutoMigrate(model)
	}
}

func (mcsv *ModelCSV) modelFromFile(filePath string) (reflect.Type, error) {
	parts := strings.Split(filePath, "/")
	// fmt.Sprintf("%v", filepath.Base)
	parts = strings.Split(parts[0], ".")
	return mcsv.structs[strings.ToLower(parts[0])], nil
}

/*
Main command method for importcsv
 */
func (mcsv *ModelCSV) ImportCSV(file string) {
	db := mcsv.connectDB()
	mcsv.createSchema(db)
	/*files := []string
	if strings.HasSuffix(file, ".csv") {
		files.append(file)
	} */
	csvFile, _ := os.Open(file)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	meta := FieldMeta{}
	model := Country{}
	// model, _ = mcsv.modelFromFile(file)
	meta.setmeta(&model, "name,code,latitude,longtitude,alias")
	errors := []error{}
	var count int = 0
	for {
		record, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			errors = append(errors, error)
			continue
		}
		country, error := meta.FillStruct(&Country{}, record)
		if error != nil {
			errors = append(errors, error)
			continue
		}
		// Create
		db.Create(country)
		count += 1
	}
	fmt.Printf("Imported %d rows to Country\n", count)
	if errors != nil {
		fmt.Printf("Failed import for %d rows due to errors:\n", len(errors))
		for _, error := range errors {
			fmt.Println(error)
		}
	}
	db.Close()
}
