# scanprotects

`scanprotects` is a command line tool that can scan a Perforce protection table and output debug information about how it was interpreted.

It is intended to be used to debug our parsing logic for protection tables as we often don't have access to them ourselves, so, instead we can send this program to the customer and ask them to run it against their `p4 protects` output.

## Usage

The intention is to pipe the output of `p4 protects` into this tool:

```
p4 protects -u USER | ./scanprotects -d "//some/test/depot/"
```

## Example output

```
...
DEBUG scanProtects perforce/protects.go:228 Scanning protects line {"line": "list group everyone * -//..."}
DEBUG fullRepoPermsScanner perforce/protects.go:382 Relevant depots {"depots": ["//depot/main/"]}
DEBUG fullRepoPermsScanner perforce/protects.go:426 Adding exclude rules {"rules": ["//**"]}
DEBUG scanProtects perforce/protects.go:228 Scanning protects line {"line": "read group readonly * //..."}
DEBUG fullRepoPermsScanner perforce/protects.go:382 Relevant depots {"depots": ["//depot/main/"]}
DEBUG fullRepoPermsScanner perforce/protects.go:391 Adding include rules {"rules": ["//**"]}
DEBUG fullRepoPermsScanner perforce/protects.go:404 Removing conflicting exclude rule {"rule": "//**"}
DEBUG scanProtects perforce/protects.go:228 Scanning protects line {"line": "write group dev * //depot/main/..."}
DEBUG fullRepoPermsScanner perforce/protects.go:382 Relevant depots {"depots": ["//depot/main/"]}
DEBUG fullRepoPermsScanner perforce/protects.go:391 Adding include rules {"rules": ["**"]}
DEBUG scanProtects perforce/protects.go:228 Scanning protects line {"line": "open group dev * //depot/main/migration/..."}
DEBUG fullRepoPermsScanner perforce/protects.go:382 Relevant depots {"depots": ["//depot/main/"]}
DEBUG fullRepoPermsScanner perforce/protects.go:391 Adding include rules {"rules": ["migration/**"]}
DEBUG scanProtects perforce/protects.go:228 Scanning protects line {"line": "write group dev * //depot/training/..."}
DEBUG fullRepoPermsScanner perforce/protects.go:382 Relevant depots {"depots": []}
DEBUG scanProtects perforce/protects.go:228 Scanning protects line {"line": "write group dev * //depot/test/..."}
DEBUG fullRepoPermsScanner perforce/protects.go:382 Relevant depots {"depots": []}
DEBUG scanProtects perforce/protects.go:228 Scanning protects line {"line": "read group baseapp_readonly * //depot/630/..."}
DEBUG fullRepoPermsScanner perforce/protects.go:382 Relevant depots {"depots": []}
DEBUG scanProtects perforce/protects.go:228 Scanning protects line {"line": "read group baseapp_readonly * //depot/636/..."}
DEBUG fullRepoPermsScanner perforce/protects.go:382 Relevant depots {"depots": []}
DEBUG scanProtects perforce/protects.go:228 Scanning protects line {"line": "write group dev * //depot/630/..."}
DEBUG fullRepoPermsScanner perforce/protects.go:382 Relevant depots {"depots": []}
...
DEBUG scanprotects scanprotects/main.go:50 Depot {"depot": "//depot/main/"}
DEBUG scanprotects scanprotects/main.go:53 Sub repo permissions {"depot": "//depot/main/"}
DEBUG scanprotects scanprotects/main.go:55 Include rule {"rule": "base/foo/**"}
DEBUG scanprotects scanprotects/main.go:55 Include rule {"rule": "base/jkl/**"}
```
