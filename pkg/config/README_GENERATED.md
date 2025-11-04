## Package: `config`

This package manages application configuration by loading environment variables and providing a structured `Config` type for access. It uses `godotenv` to optionally load `.env` files, logging warnings if the process fails but continuing execution. The primary function is `NewConfig()`, which initializes a `Config` struct with values from environment variables or defaults (e.g., `.reflexia_cache` for `CachePath`).

**Environment Variables:**

*   `CACHE_PATH`: Cache directory path (defaults to `.reflexia_cache`).
*   `EMBEDDINGS_AI_URL`: Embeddings AI service URL.
*   `EMBEDDINGS_AI_TOKEN`: Embeddings AI authentication token.
*   `EMBEDDINGS_DB_URL`: Embeddings database URL.
*   `EMBEDDINGS_SIM_SEARCH_TEST_PROMPT`: Test prompt for similarity search testing.

**Config Structure:**

The `Config` struct contains:

*   `EmbeddingsAIURL`: String (from `EMBEDDINGS_AI_URL`).
*   `EmbeddingsAIToken`: String (from `EMBEDDINGS_AI_TOKEN`).
*   `EmbeddingsDBURL`: String (from `EMBEDDINGS_DB_URL`).
*   `CachePath`: String (from `CACHE_PATH`, defaults to `.reflexia_cache`).
*   `EmbeddingsSimSearchTestPrompt`: String (from `EMBEDDINGS_SIM_SEARCH_TEST_PROMPT`).

**File Structure:**

```
pkg/config/
├── config.go
```