package generations

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

func LoadDatabases(paths []string, baseDir string) (*MemoryDatabase, error) {
	database := NewMemoryDatabase()
	basePath, err := homedir.Expand(baseDir)
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		path, err := homedir.Expand(path)
		if err != nil {
			return nil, err
		}
		dbFiles, err := filepath.Glob(path)
		if err != nil {
			return nil, err
		}
		if len(dbFiles) == 0 {
			dbFiles = []string{path}
		}
		for _, db := range dbFiles {
			if !filepath.IsAbs(db) {
				db = filepath.Join(basePath, db)
			}
			if err := database.ParseYamlFile(db); err != nil {
				return nil, err
			}
		}
	}
	return database, nil
}
