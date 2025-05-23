package project

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/JackBekket/reflexia/internal/util"

	"github.com/pelletier/go-toml/v2"
)

type ProjectConfig struct {
	FileFilter        []string                        `toml:"file_filter"`
	ProjectRootFilter []string                        `toml:"project_root_filter"`
	ModuleMatch       string                          `toml:"module_match"`
	StopWords         []string                        `toml:"stop_words"`
	Prompts           map[string]ProjectConfigPrompts `toml:"prompts"`

	RootPath string
}

type ProjectConfigPrompts struct {
	CodePrompt            string  `toml:"code"`
	CodePromptFallback    *string `toml:"code_fallback"`
	PackagePrompt         string  `toml:"package"`
	PackagePromptFallback *string `toml:"package_fallback"`
}

func GetProjectConfig(
	currentDirectory, withConfigFile string, lightCheck bool,
) (map[string]*ProjectConfig, error) {
	if _, err := os.Stat("project_config"); os.IsNotExist(err) {
		if _, err := os.Stat(withConfigFile); err == nil {
			content, err := os.ReadFile(withConfigFile)
			if err != nil {
				return nil, err
			}
			var config ProjectConfig
			if err = toml.Unmarshal(content, &config); err != nil {
				return nil, err
			}
			config.RootPath = currentDirectory
			return map[string]*ProjectConfig{withConfigFile: &config}, nil
		}
	}

	projectConfigs, err := ListProjectConfigs(currentDirectory)
	if err != nil {
		return nil, fmt.Errorf("list project configs: %w", err)
	}

	if config, exists := projectConfigs[filepath.Base(withConfigFile)]; exists {
		return map[string]*ProjectConfig{withConfigFile: &config}, nil
	}

	var projectConfigVariants = map[string]*ProjectConfig{}
	for filename, config := range projectConfigs {
		foundFilterFiles, err := hasFilterFiles(currentDirectory, config.FileFilter)
		if err != nil {
			return nil, fmt.Errorf("has filter files: %w", err)
		}

		if lightCheck && foundFilterFiles {
			projectConfigVariants[filename] = &config
			continue
		}

		foundRootFilterFile, err := hasRootFilterFile(currentDirectory, config.ProjectRootFilter)
		if err != nil {
			return nil, fmt.Errorf("has root filter file: %w", err)
		}
		if foundFilterFiles && foundRootFilterFile {
			projectConfigVariants[filename] = &config
		}
	}
	return projectConfigVariants, nil
}

func ListProjectConfigs(currentDirectory string) (map[string]ProjectConfig, error) {
	var projectConfigs = map[string]ProjectConfig{}

	if err := util.WalkDirIgnored(
		"project_config", "",
		func(path string, d fs.DirEntry) error {
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".toml") {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				var config ProjectConfig
				if err = toml.Unmarshal(content, &config); err != nil {
					return err
				}
				config.RootPath = currentDirectory
				projectConfigs[d.Name()] = config
			}
			return nil
		}); err != nil {
		return nil, err
	}

	return projectConfigs, nil
}

func (pc *ProjectConfig) BuildPackageFiles() (map[string][]string, error) {
	packageFileMap := map[string][]string{}
	switch pc.ModuleMatch {
	case "directory":
		if err := util.WalkDirIgnored(
			pc.RootPath,
			filepath.Join(pc.RootPath, ".gitignore"),
			func(path string, d fs.DirEntry) error {
				if filepath.Dir(path) == pc.RootPath || path == pc.RootPath {
					return nil
				}
				for _, filter := range pc.FileFilter {
					if strings.HasSuffix(d.Name(), filter) {
						relPath, err := filepath.Rel(pc.RootPath, path)
						if err != nil {
							return err
						}
						key := filepath.Dir(relPath)

						if _, exists := packageFileMap[key]; !exists {
							packageFileMap[key] = []string{}
						}
						packageFileMap[key] = append(packageFileMap[key], relPath)
					}
				}

				return nil
			}); err != nil {
			return nil, err
		}

	case "go_package":
		if err := util.WalkDirIgnored(
			pc.RootPath,
			filepath.Join(pc.RootPath, ".gitignore"),
			func(path string, d fs.DirEntry) error {
				for _, filter := range pc.FileFilter {
					if strings.HasSuffix(d.Name(), filter) {
						fset := token.NewFileSet()
						ast, err := parser.ParseFile(fset, path, nil, 0)
						if err != nil {
							return err
						}
						relPath, err := filepath.Rel(pc.RootPath, path)
						if err != nil {
							return err
						}
						key := fmt.Sprintf("%s:%s", filepath.Dir(relPath), ast.Name.Name)

						if _, exists := packageFileMap[key]; !exists {
							packageFileMap[key] = []string{}
						}
						packageFileMap[key] = append(packageFileMap[key], relPath)
					}
				}

				return nil
			}); err != nil {
			return nil, err
		}

	default:
		return nil, errors.New(pc.ModuleMatch + " module match mode unimplemented")
	}

	return packageFileMap, nil
}

func hasFilterFiles(workdir string, filters []string) (bool, error) {
	found := false

	err := util.WalkDirIgnored(
		workdir, filepath.Join(workdir, ".gitignore"),
		func(path string, d fs.DirEntry) error {
			for _, filter := range filters {
				if strings.HasSuffix(d.Name(), filter) {
					found = true
					return io.EOF
				}
			}
			return nil
		})
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return false, err
	}
	return found, nil
}

func hasRootFilterFile(workdir string, filters []string) (bool, error) {
	for _, filter := range filters {
		if _, err := os.Stat(filepath.Join(workdir, filter)); err != nil {
			if !os.IsNotExist(err) {
				return false, err
			}
		} else {
			return true, nil
		}
	}

	return false, nil
}
