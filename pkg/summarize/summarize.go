package summarize

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	util "github.com/JackBekket/reflexia/internal"
	"github.com/JackBekket/reflexia/pkg/project"

	helper "github.com/JackBekket/hellper/lib/langchain"
	"github.com/tmc/langchaingo/llms"
)

type SummarizerService struct {
	HelperURL      string
	Model          string
	ApiToken       string
	Network        string
	LlmOptions     []llms.CallOption
	OverwriteCache bool
	CachePath      string
}

func (s *SummarizerService) CodeSummaryRequest(prompt, content string) (string, error) {
	response, err := helper.GenerateContentInstruction(s.HelperURL,
		prompt+"```"+content+"```",
		s.Model, s.ApiToken, s.Network,
		s.LlmOptions...,
	)

	return response, err
}

func (s *SummarizerService) SummarizeRequest(prompt, content string) (string, error) {
	response, err := helper.GenerateContentInstruction(s.HelperURL,
		prompt+"\n\n"+content,
		s.Model, s.ApiToken, s.Network,
		s.LlmOptions...,
	)

	return response, err
}

// This function is summarizing code of files and create map fileMap[relPath] = generation_content
func (s *SummarizerService) SummarizeCode(
	projectConfig *project.ProjectConfig,
) (map[string]string, error) {

	fileMap := map[string]string{}
	cacheFileMap := map[string]string{}
	if !s.OverwriteCache {
		hash, err := projectConfig.ProjectHash()
		if err != nil {
			return map[string]string{}, err
		}
		err = project.LoadCacheFileMap(filepath.Join(s.CachePath, hash), cacheFileMap)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return map[string]string{}, err
			}
		}
	}

	if err := util.WalkDirIgnored(
		projectConfig.RootPath,
		filepath.Join(projectConfig.RootPath, ".gitignore"),
		func(path string, d fs.DirEntry) error {
			for _, filter := range projectConfig.FileFilter {
				if strings.HasSuffix(d.Name(), filter) {
					content, err := os.ReadFile(path)
					if err != nil {
						return err
					}

					relPath, err := filepath.Rel(projectConfig.RootPath, path)
					if err != nil {
						return err
					}
					fmt.Println(relPath)

					contentStr := string(content)
					response := "Empty file"

					if len(strings.TrimSpace(contentStr)) > 0 {
						response = cacheFileMap[relPath]
						if response == "" {
							response, err = s.CodeSummaryRequest(
								projectConfig.CodeSummaryPrompt, string(content))
							if err != nil {
								return err
							}
						} else {
							fmt.Println("Using cached result:\n", response)
						}
					} else {
						fmt.Println("Empty file")
					}

					fmt.Printf("\n")
					fileMap[relPath] = response

					hash, err := projectConfig.ProjectHash()
					if err != nil {
						return err
					}
					err = project.SaveCacheFileMap(filepath.Join(s.CachePath, hash), fileMap)
					if err != nil {
						return err
					}

					break
				}
			}
			return nil
		}); err != nil {
		return nil, err
	}

	return fileMap, nil
}
