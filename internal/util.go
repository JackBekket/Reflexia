package util

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	ignore "github.com/crackcomm/go-gitignore"
)

type WalkDirIgnoredFunction func(path string, d fs.DirEntry) error

func WalkDirIgnored(workdir, gitignorePath string, f WalkDirIgnoredFunction) error {
	ignoreFile, err := ignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		log.Printf("failed to load .gitignore file %v", err)
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

func LoadEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("empty environment key %s", key)
	}
	return value
}
