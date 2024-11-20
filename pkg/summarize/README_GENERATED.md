## Package: summarize

### Imports:

```
crypto/sha256
encoding/hex
errors
fmt
io/fs
os
path/filepath
github.com/JackBekket/hellper/lib/langchain
github.com/tmc/langchaingo/llms
```

### External Data, Input Sources:

- HelperURL: URL for the helper service.
- Model: Name of the language model to use.
- ApiToken: API token for the helper service.
- Network: Network for the language model (e.g., "http" or "https").
- LlmOptions: Options for the language model call.
- OverwriteCache: Flag to overwrite the cache.
- IgnoreCache: Flag to ignore the cache.
- CachePath: Path to the cache directory.

### SummarizeService:

The SummarizeService struct is responsible for handling the summarization process. It has the following fields:

- HelperURL: URL for the helper service.
- Model: Name of the language model to use.
- ApiToken: API token for the helper service.
- Network: Network for the language model (e.g., "http" or "https").
- LlmOptions: Options for the language model call.
- OverwriteCache: Flag to overwrite the cache.
- IgnoreCache: Flag to ignore the cache.
- CachePath: Path to the cache directory.

The LLMRequest method takes a format string and a list of arguments to construct the final prompt. It then hashes the prompt and checks if the result is already cached. If the result is not cached or the OverwriteCache flag is set, it calls the helper service to generate the content and saves the result to the cache.

### hashStrings:

This function takes a list of strings and returns a hash of the concatenated strings. It uses the SHA256 algorithm to generate the hash and encodes the result as a hexadecimal string.

### loadCache:

This function loads the cached result for a given hash from the cache directory. It returns the cached content and an error if the cache file does not exist or if there is an error reading the file.

### saveCache:

This function saves the given content to the cache directory with the specified hash. It creates the cache directory if it does not exist and writes the content to the file.

```
summarize/
  summarize.go
```

