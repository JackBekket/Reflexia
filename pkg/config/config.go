package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	EmbeddingsAIURL               string
	EmbeddingsAIToken             string
	EmbeddingsDBURL               string
	CachePath                     string
	EmbeddingsSimSearchTestPrompt string
}

func NewConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("load .env file")
	}

	config := Config{}

	config.CachePath = os.Getenv("CACHE_PATH")
	if config.CachePath == "" {
		config.CachePath = ".reflexia_cache"
	}

	config.EmbeddingsAIURL = os.Getenv("EMBEDDINGS_AI_URL")
	config.EmbeddingsAIToken = os.Getenv("EMBEDDINGS_AI_TOKEN")
	config.EmbeddingsDBURL = os.Getenv("EMBEDDINGS_DB_URL")
	config.EmbeddingsSimSearchTestPrompt = os.Getenv("EMBEDDINGS_SIM_SEARCH_TEST_PROMPT")

	return config
}
