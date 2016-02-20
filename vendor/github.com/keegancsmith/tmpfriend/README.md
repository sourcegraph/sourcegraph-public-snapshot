# tmpfriend

`tmpfriend` is a Go library to help prevent misbehaving subprocesses / code
from forgetting to cleanup after themselves. It works by modifying the
location of the temporary directory to one unique for the current process, and
on start will clean up older temporary directories for non-existant processes.

```
func main() {
	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()
	// ...
}
```
