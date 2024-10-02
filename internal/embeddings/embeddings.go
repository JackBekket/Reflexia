/package embeddings

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JackBekket/hellper/lib/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

func LoadSummary(base_link string, api_token string, db_link string, repo_name string, summary string, package_name string) {

	store, err := embeddings.GetVectorStoreWithOptions(api_token, db_link, base_link, repo_name)

	if err != nil {
		fmt.Println(err)
	} else {
		

		doc := []schema.Document{
			{
				PageContent: summary,
				Metadata: map[string]interface{}{
					"package":   package_name,
				},
			},
		}
		LoadDocsToStoreOptions(doc, store)

		fmt.Println("Data for " + id + " successfully loaded into vector store💾")
	}
}

func LoadDocsToStoreOptions(docs []schema.Document, store vectorstores.VectorStore) {
	//a := vectorstores.WithNameSpace("GeoTest")

	_, err := store.AddDocuments(context.Background(), docs)

	if err != nil {
		log.Panic(err)
	}

	fmt.Println("data successfully loaded into vector store")

	log.Println(err)
}
