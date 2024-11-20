# pkg/project/project.go  
## Package: project  
  
### Imports:  
  
```  
errors  
go/parser  
go/token  
io  
io/fs  
log  
os  
path/filepath  
strings  
github.com/JackBekket/reflexia/internal  
github.com/pelletier/go-toml/v2  
```  
  
### External Data, Input Sources:  
  
1. Project configuration files with `.toml` extension. These files can be located in the current directory or in a specified path.  
2. `.gitignore` files for ignoring specific files and directories during the file walk.  
  
### Summary of Major Code Parts:  
  
#### ProjectConfig Struct:  
  
This struct defines the configuration for a project, including file filters, project root filters, module match mode, stop words, code prompts, and package prompts.  
  
#### GetProjectConfig Function:  
  
This function retrieves the project configuration from the specified path or the current directory. It first checks if a project configuration file exists in the current directory. If not, it searches for a configuration file in the specified path. If a configuration file is found, it parses the TOML data and returns a map of project configurations.  
  
#### BuildPackageFiles Function:  
  
This function builds a map of package files based on the specified module match mode. It iterates through the files in the project root directory and filters them based on the file filters and module match mode. The function returns a map of package names to a list of corresponding file paths.  
  
#### hasFilterFiles Function:  
  
This function checks if any files in the specified directory match the provided filters. It iterates through the files in the directory and returns true if any file matches a filter, otherwise false.  
  
#### hasRootFilterFile Function:  
  
This function checks if any files in the specified directory match the provided filters. It iterates through the filters and checks if each filter file exists in the directory. If any filter file exists, it returns true, otherwise false.  
  
