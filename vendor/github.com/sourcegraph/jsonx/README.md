# jsonx

Package jsonx is an extended JSON library for Go. It is highly tolerant of
errors, and it supports trailing commas and comments (`//` and `/* ... */`).

It is ported from [Visual Studio Code's](https://github.com/Microsoft/vscode)
comment-aware JSON parsing and editing APIs in TypeScript, specifically in these
files:

* [src/vs/base/common/json.ts](https://github.com/Microsoft/vscode/tree/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts)
* [src/vs/base/common/jsonEdit.ts](https://github.com/Microsoft/vscode/tree/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonEdit.ts)
* [src/vs/base/common/jsonFormatter.ts](https://github.com/Microsoft/vscode/tree/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonFormatter.ts)

## Status: Experimental

* Where the original TypeScript code's API is not idiomatic in Go, this library
  does not (yet) attempt to provide an idiomatic Go API. This is mainly evident
  in the error return API for parsing and scanning errors.
