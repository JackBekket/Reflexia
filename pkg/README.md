## Package: store

### Imports:

- context
- log
- github.com/jackc/pgx/v5/pgxpool
- github.com/tmc/langchaingo/embeddings
- github.com/tmc/langchaingo/llms/openai
- github.com/tmc/langchaingo/vectorstores
- github.com/tmc/langchaingo/vectorstores/pgvector

### External Data, Input Sources:

- ai_url (string): Base URL for the OpenAI API.
- api_token (string): API token for the OpenAI API.
- db_link (string): Connection URL for the PostgreSQL database.
- name (string): Name of the collection to be created in the database.

### EmbeddingsService:

This struct represents an embeddings service and holds a reference to a vector store.

### NewVectorStoreWithPreDelete:

This function creates a new vector store with pre-delete collection enabled. It takes the following parameters:

- ai_url (string): Base URL for the OpenAI API.
- api_token (string): API token for the OpenAI API.
- db_link (string): Connection URL for the PostgreSQL database.
- name (string): Name of the collection to be created in the database.

The function first parses the database connection URL and creates a new PostgreSQL connection pool. Then, it initializes an OpenAI client using the provided base URL, API token, embedding model, and API version. Next, it creates an embeddings object using the OpenAI client.

Finally, it creates a new vector store using the PostgreSQL connection pool, embedding object, and collection name. The `WithPreDeleteCollection(true)` option ensures that the collection will be deleted before creating a new one. The function returns the newly created vector store and any errors encountered during the process.



