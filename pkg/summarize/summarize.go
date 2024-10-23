package summarize

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	helper "github.com/JackBekket/hellper/lib/langchain"
	"github.com/tmc/langchaingo/llms"
)

type SummarizeService struct {
	HelperURL      string
	Model          string
	ApiToken       string
	Network        string
	LlmOptions     []llms.CallOption
	OverwriteCache bool
	IgnoreCache    bool
	CachePath      string
}

func (s *SummarizeService) LLMRequest(format string, a ...string) (string, error) {
	finalPrompt := fmt.Sprintf(format, a)
	cacheHash, err := hashStrings(finalPrompt)
	if err != nil {
		return "", err
	}
	response := ""

	if !(s.OverwriteCache || s.IgnoreCache) {
		response, err = loadCache(s.CachePath, cacheHash)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
	}

	if response == "" {
		if response, err = helper.GenerateContentInstruction(s.HelperURL,
			finalPrompt, s.Model, s.ApiToken, s.Network, s.LlmOptions...,
		); err != nil {
			return "", err
		}
		if !s.IgnoreCache {
			if err = saveCache(s.CachePath, cacheHash, response); err != nil {
				return "", err
			}
		}
	} else {
		fmt.Printf("Using cached result:\n%s\n", response)
	}

	return response, err
}

func hashStrings(a ...string) (string, error) {
	hash := sha256.New()
	for _, s := range a {
		if _, err := hash.Write([]byte(s)); err != nil {
			return "", err
		}
	}

	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

func loadCache(cachePath, hash string) (string, error) {
	cache, err := os.ReadFile(filepath.Join(cachePath, hash))
	return string(cache), err
}

func saveCache(cachePath, hash, content string) error {
	if err := os.MkdirAll(cachePath, os.ModePerm); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(cachePath, hash))
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(content); err != nil {
		return err
	}
	return nil
}
