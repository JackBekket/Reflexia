# internal/util.go  
## Package: util  
  
### Imports:  
  
- io/fs  
- log  
- path/filepath  
- ignore "github.com/crackcomm/go-gitignore"  
  
### External Data, Input Sources:  
  
- workdir (string): The directory to walk.  
- gitignorePath (string): The path to a .gitignore file.  
- f (WalkDirIgnoredFunction): A function to be called for each file and directory encountered during the walk.  
  
### Summary:  
  
#### WalkDirIgnored Function:  
  
The `WalkDirIgnored` function takes a working directory, a path to a .gitignore file, and a function to be called for each file and directory encountered during the walk. It walks the directory tree starting at the given working directory, and for each file and directory encountered, it checks if it should be ignored based on the .gitignore file. If the file or directory should be ignored, it skips it. Otherwise, it calls the provided function with the path and directory entry.  
  
The function first checks if a .gitignore file is provided. If so, it compiles the .gitignore file into an ignore.GitIgnore object. Then, it walks the directory tree using `filepath.WalkDir`, and for each file and directory encountered, it checks if it should be ignored based on the .gitignore file. If the file or directory should be ignored, it skips it. Otherwise, it calls the provided function with the path and directory entry.  
  
The function returns an error if any errors occur during the walk or if the .gitignore file cannot be loaded.  
  
