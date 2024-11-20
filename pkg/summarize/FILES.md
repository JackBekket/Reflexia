# pkg/summarize/summarize.go  
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
  
The hashStrings function takes a list of strings and returns a hexadecimal representation of their SHA256 hash. It iterates through the strings, writes them to the hash object, and returns the hexadecimal representation of the resulting hash.  
  
### loadCache:  
  
The loadCache function takes the cache path and hash as input and returns the cached content as a string. It reads the file from the cache directory based on the hash and returns the content.  
  
### saveCache:  
  
The saveCache function takes the cache path, hash, and content as input and saves the content to the cache directory. It creates the cache directory if it doesn't exist and writes the content to the file based on the hash.  
  
  
  
