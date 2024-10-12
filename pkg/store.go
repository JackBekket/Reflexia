package store

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

type EmbeddingsService struct {
	Store vectorstores.VectorStore
}

func NewVectorStoreWithPreDelete(ai_url string, api_token string, db_link string, name string) (vectorstores.VectorStore, error) {
	base_url := ai_url
	pgConnURL := db_link

	config, err := pgxpool.ParseConfig(pgConnURL)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	llm, err := openai.New(
		openai.WithBaseURL(base_url),
		openai.WithAPIVersion("v1"),
		openai.WithEmbeddingModel("text-embedding-ada-002"),
		openai.WithToken(api_token),
	)
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		return nil, err
	}

	store, err := pgvector.New(
		context.Background(),
		pgvector.WithCollectionName(name),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithConn(pool),
		pgvector.WithEmbedder(e),
	)
	if err != nil {
		return nil, err
	}

	return store, nil
}
