// Base class equivalent = embedded Files for ModelCSV and CSVMeta
package importcsv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Embedded stuct to hang the FilesFetch method on
type Files struct {
}

/*
Return a map of lowercase filename without suffix vs file object from file or dir path
*/
func (mcsv *Files) FilesFetch (path string) (map[string]*os.File, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("The fixture file %s does not exist", path)
	}
	filesMap := map[string]*os.File{}
	files := []*os.File{}
	f, err := os.Open(path)
	if strings.HasSuffix(path, ".csv") {
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	} else {
		if err != nil {
			return nil, err
		}
		fileInfos, err := f.Readdir(-1)
		f.Close()
		for _, fileInfo := range fileInfos {
			csvFile, err := os.Open(filepath.Join(path, fileInfo.Name()))
			if err != nil {
				return nil, err
			}
			files = append(files, csvFile)
		}
		if err != nil {
			return nil, err
		}
	}
	for _, file := range files {
		parts := strings.Split(file.Name(), "/")
		parts = strings.Split(parts[len(parts) - 1], ".")
		// name := strings.ToLower(parts[0])
		filesMap[parts[0]] = file
	}
	return filesMap, nil
}
