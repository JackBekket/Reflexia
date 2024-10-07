package store

import (
	"github.com/tmc/langchaingo/vectorstores"
)

type EmbeddingsService struct {
	Store vectorstores.VectorStore
}
