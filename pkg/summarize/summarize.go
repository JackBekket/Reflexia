package summarize

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Swarmind/libagent/pkg/agent"
	agentUtil "github.com/Swarmind/libagent/pkg/util"
	"github.com/tmc/langchaingo/llms"
)

type SummarizeService struct {
	Agent          agent.Agent
	LlmOptions     []llms.CallOption
	OverwriteCache bool
	IgnoreCache    bool
	CachePath      string
	StopWords      []string
	Model          string
}

func (s SummarizeService) LLMRequest(ctx context.Context, format string, a ...any) (string, error) {
	finalPrompt := fmt.Sprintf(format, a...)
	cacheHash, err := hashStrings(finalPrompt + s.Model)
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

	if response != "" {
		return response, nil
	}

	if response, err = s.Agent.SimpleRun(
		ctx, finalPrompt, s.LlmOptions...,
	); err != nil {
		return "", err
	}

	response = agentUtil.RemoveThinkTag(response)
	for _, stopWord := range s.StopWords {
		response = strings.TrimSuffix(response, stopWord)
	}
	response = strings.TrimSpace(response)

	if !s.IgnoreCache {
		if err = saveCache(s.CachePath, cacheHash, response); err != nil {
			return "", err
		}
	}

	return response, nil
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
