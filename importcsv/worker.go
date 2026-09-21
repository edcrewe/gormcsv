package importcsv

import (
	"reflect"
	"strings"

	"github.com/edcrewe/gormcsv/meta"
	"gorm.io/gorm"
)

type batchResult struct {
	count      int
	duplicates int
	errors     []error
}

func worker(jobs <-chan [][]string, results chan<- batchResult, fieldMeta *meta.FieldMeta, db *gorm.DB, factory ModelFactory, name string) {
	for batch := range jobs {
		res := batchResult{}
		modelType := reflect.TypeOf(factory.New(name))
		sliceType := reflect.SliceOf(modelType)
		sliceVal := reflect.MakeSlice(sliceType, 0, len(batch))

		var models []interface{}
		for _, record := range batch {
			m, err := fieldMeta.RecordToModel(factory.New(name), record)
			if err != nil {
				res.errors = append(res.errors, err)
				continue
			}
			sliceVal = reflect.Append(sliceVal, reflect.ValueOf(m))
			models = append(models, m)
		}

		if sliceVal.Len() > 0 {
			dbRes := db.Create(sliceVal.Interface())
			if dbRes.Error != nil {
				// Fallback to sequential insertion if batch fails (e.g., duplicate constraint)
				for _, m := range models {
					singleRes := db.Create(m)
					if singleRes.Error != nil {
						if strings.Contains(singleRes.Error.Error(), "UNIQUE constraint failed") || strings.Contains(singleRes.Error.Error(), "duplicate key value") {
							res.duplicates++
						} else {
							res.errors = append(res.errors, singleRes.Error)
						}
					} else {
						res.count++
					}
				}
			} else {
				res.count += int(dbRes.RowsAffected)
			}
		}
		results <- res
	}
}
