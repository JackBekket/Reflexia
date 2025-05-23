package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

type EmbeddingsService struct {
	Store vectorstores.VectorStore
}

func NewVectorStoreWithPreDelete(
	ctx context.Context,
	aiURL, aiToken, dbURL, name string,
) (vectorstores.VectorStore, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	llm, err := openai.New(
		openai.WithBaseURL(aiURL),
		openai.WithToken(aiToken),
		openai.WithEmbeddingModel("text-embedding-ada-002"),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		return nil, fmt.Errorf("new openai: %w", err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("new embedder: %w", err)
	}

	store, err := pgvector.New(ctx,
		pgvector.WithCollectionName(name),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithConn(pool),
		pgvector.WithEmbedder(e),
	)
	if err != nil {
		return nil, fmt.Errorf("create store: %w", err)
	}

	return store, nil
}
