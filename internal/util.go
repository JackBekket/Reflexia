package util

import (
	"io/fs"
	"log"
	"path/filepath"

	ignore "github.com/crackcomm/go-gitignore"
)

type WalkDirIgnoredFunction func(path string, d fs.DirEntry) error

func WalkDirIgnored(workdir, gitignorePath string, f WalkDirIgnoredFunction) error {
	var ignoreFile *ignore.GitIgnore
	var err error
	if gitignorePath != "" {
		ignoreFile, err = ignore.CompileIgnoreFile(gitignorePath)
		if err != nil {
			log.Printf("failed to load .gitignore file %v", err)
		}
	}
	err = filepath.WalkDir(workdir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(workdir, path)
		if err != nil {
			return err
		}
		if ignoreFile != nil && ignoreFile.MatchesPath(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		return f(path, d)
	})
	return err
}
