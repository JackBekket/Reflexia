package util

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	ignore "github.com/crackcomm/go-gitignore"
	"github.com/rs/zerolog/log"
)

type WalkDirIgnoredFunction func(path string, d fs.DirEntry) error

func WalkDirIgnored(workdir, gitignorePath string, f WalkDirIgnoredFunction) error {
	var ignoreFile *ignore.GitIgnore
	var err error
	if gitignorePath != "" {
		ignoreFile, err = ignore.CompileIgnoreFile(gitignorePath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Warn().Err(err).Msg("failed to load .gitignore file")
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
