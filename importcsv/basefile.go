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

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil{
		return false, err
	}
	return fileInfo.IsDir(), err
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
	dir, err := IsDirectory(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if !dir {
		files = append(files, f)
	} else {
		fileInfos, err := f.Readdir(-1)
		f.Close()
		for _, fileInfo := range fileInfos {
			filePath := filepath.Join(path, fileInfo.Name())
			dir, err := IsDirectory(filePath)
			if err != nil {
				return nil, err
			}
			if dir {
				continue
			}
			csvFile, err := os.Open(filePath)
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
		_, name := filepath.Split(file.Name())
		parts := strings.Split(name, ".")
		// name := strings.ToLower(parts[0])
		filesMap[parts[0]] = file
	}
	return filesMap, nil
}
