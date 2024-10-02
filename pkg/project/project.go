package project

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	util "github.com/JackBekket/reflexia/internal"

	"github.com/pelletier/go-toml/v2"
)

type ProjectConfig struct {
	FileFilter        []string `toml:"file_filter"`
	ProjectRootFilter []string `toml:"project_root_filter"`
	ModuleMatch       string   `toml:"module_match"`
	StopWords         []string `toml:"stop_words"`
	CodePrompt        string   `toml:"code_prompt"`
	ExplainPrompt     string   `toml:"explain_code_prompt"`
	CodeSummaryPrompt string   `toml:"code_summary_prompt"`
	SummaryPrompt     string   `toml:"summary_prompt"`
	PackagePrompt     string   `toml:"package_prompt"`
	ReadmePrompt      string   `toml:"readme_prompt"`
	RootPath          string
}

func GetProjectConfig(
	currentDirectory, withConfigFile string, lightCheck bool,
) (*ProjectConfig, error) {
	var projectConfigs = map[string]ProjectConfig{}

	if err := util.WalkDirIgnored(
		"project_config", filepath.Join(currentDirectory, ".gitignore"),
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

	if config, exists := projectConfigs[filepath.Base(withConfigFile)]; exists {
		return &config, nil
	}
	var projectConfigVariants = map[string]*ProjectConfig{}
	for filename, config := range projectConfigs {
		if lightCheck && hasFilterFiles(currentDirectory, config.FileFilter) {
			projectConfigVariants[filename] = &config
		}
		if hasFilterFiles(currentDirectory, config.FileFilter) &&
			hasRootFilterFile(currentDirectory, config.ProjectRootFilter) {
			projectConfigVariants[filename] = &config
		}
	}

	switch len(projectConfigVariants) {
	case 0:
		return nil, errors.New(
			"failed to detect project language, available languages: go, python, typescript",
		)
	case 1:
		for _, config := range projectConfigVariants {
			return config, nil
		}
	default:
		var filenames []string
		for filename, _ := range projectConfigVariants {
			filenames = append(filenames, filename)
		}
		fmt.Println("Multiple project config matches found!")
		for i, filename := range filenames {
			fmt.Printf("%d. %v\n", i+1, filename)
		}
		fmt.Print("Enter the number or filename: ")
		for {
			var input string
			if _, err := fmt.Scanln(&input); err != nil {
				return nil, err
			}
			if index, err := strconv.Atoi(input); err == nil && index > 0 && index <= len(filenames) {
				return projectConfigVariants[filenames[index-1]], nil
			} else {
				for filename, config := range projectConfigVariants {
					if filename == input || strings.TrimSuffix(filename, ".toml") == input {
						return config, nil
					}
				}
			}
		}
	}
	panic("unreachable")
}

func (pc *ProjectConfig) ProjectHash() (string, error) {
	hash := sha256.New()

	v := reflect.ValueOf(*pc)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := field.Type()

		if fieldType.Kind() == reflect.String {
			if _, err := hash.Write([]byte(field.String())); err != nil {
				return "", err
			}
		}

		if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.String {
			joinedStrings := strings.Join(field.Interface().([]string), "")
			if _, err := hash.Write([]byte(joinedStrings)); err != nil {
				return "", err
			}
		}
	}

	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
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

	case "package_name":
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
						key := ast.Name.Name
						relPath, err := filepath.Rel(pc.RootPath, path)
						if err != nil {
							return err
						}

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

func hasFilterFiles(workdir string, filters []string) bool {
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
		log.Fatal(err)
	}
	return found
}

func hasRootFilterFile(workdir string, filters []string) bool {
	for _, filter := range filters {
		if _, err := os.Stat(filepath.Join(workdir, filter)); err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
		} else {
			return true
		}
	}

	return false
}
