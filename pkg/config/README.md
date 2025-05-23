# config

A package to manage configuration settings, loading environment variables from a `.env` file and providing default values for various parameters.

## Package Structure
- **config.go**: Contains the configuration handling logic.

## Configuration Options
- `CACHE_PATH`: Cache directory path (default: ".reflexia_cache")
- `EMBEDDINGS_AI_URL`: Embeddings AI service URL
- `EMBEDDINGS_AI_TOKEN`: Embeddings AI API token
- `EMBEDDINGS_DB_URL`: Embeddings database URL
- `EMBEDDINGS_SIM_SEARCH_TEST_PROMPT`: Test prompt for similarity search

## Notes
- The `.env` file is loaded but errors are logged as warnings, not fatal.
- No validation is performed to ensure required environment variables are set (e.g., `EMBEDDINGS_AI_URL` could be empty).
- No TODOs, comments, or edge cases are explicitly addressed in the code.