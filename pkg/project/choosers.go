package project

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func FirstChooser(projectConfigVariants map[string]*ProjectConfig) (*ProjectConfig, error) {
	if len(projectConfigVariants) == 0 {
		return nil, errors.New(
			"failed to detect project language, available languages: go, python, typescript",
		)
	}
	keys := slices.Collect(maps.Keys(projectConfigVariants))
	sort.Strings(keys)
	return projectConfigVariants[keys[0]], nil
}

func CLIChooser(projectConfigVariants map[string]*ProjectConfig) (*ProjectConfig, error) {
	switch len(projectConfigVariants) {
	case 0:
		return nil, errors.New(
			"failed to detect project language, available languages: go, python, typescript",
		)
	case 1:
		for _, pc := range projectConfigVariants {
			return pc, nil
		}
	default:
		var filenames []string
		for filename := range projectConfigVariants {
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
				log.Fatal().Err(err).Msg("scanln")
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
