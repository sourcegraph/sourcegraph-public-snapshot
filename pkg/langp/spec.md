# Language Processor REST API

This document contains the specification for the Sourcegraph Language Processor REST API.

- [Overview](#overview)
- [Methods](#methods)
  - [POST /prepare](#post-prepare)
  - [POST /defspec-to-position](#post-defspec-to-position)
  - [POST /definition](#post-definition)
  - [POST /hover](#post-hover)
  - [POST /local-refs](#post-local-refs)
  - [POST /external-refs](#post-external-refs)
  - [POST /defspec-refs](#post-defspec-refs)
  - [POST /exported-symbols](#post-exported-symbols)
- Data Types
  - [Error](#type-error)
  - [Position](#type-position)
  - [Range](#type-range)
  - [RefLocations](#type-reflocations)
  - [Hover](#type-hover)
  - [HoverContent](#type-hovercontent)
  - [RepoRev](#type-reporev)
  - [ExternalRefs](#type-externalrefs)
  - [ExportedSymbols](#type-exportedsymbols)
  - [DefSpec](#type-defspec)
  - [Symbol](#type-symbol)

# Overview

The application (i.e. the main codebase serving [sourcegraph.com](http://sourcegraph.com/) today) talks to a Language Processor server using a custom REST protocol. **VS Code LSP servers do not implement this.**

The REST protocol solely serves the needs of the App → Language Processor exchange and as such is inherently a (modified) version of what LSP provides, see section at the end of this document describing differences).

# Methods

Below are all of the methods supported by the API.

Every method is only HTTP POST, even for things that traditionally would be an HTTP GET. This enables us to speak purely in JSON objects and not need query parameter parsing/encoding (parameters are just JSON objects in the body of the POST request).

## POST /prepare

Informs the Language Processor that it should prepare a workspace for the specified repo / commit. It is sent prior to an actual user request (e.g. as soon as we have access to their repos) in hopes of having preparation completed already when a user makes their first request. The request should not block / workspace preparation should be done in the background by the language processor, and the request should return immediately.

- LSP equivalent: None (Language Processor only)
- Request:
  - Body: [{RepoRev Object}](#type-reporev)
  - Response: `{}` OR [{Error Object}](#type-error)

## POST /defspec-to-position

Converts a DefSpec into a Position object.

- LSP equivalent: None (Language Processor only)
- Request:
  - Body: [{DefSpec Object}](#type-defspec)
  - Response: [{Position Object}](#type-position) OR [{Error Object}](#type-error)

## POST /definition

Resolves the specified position, effectively returning where the given definition is defined. For example, this is used for go to definition.

- LSP equivalent: [textDocument/definition](https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#goto-definition)
- Request:
  - Body: [{Position Object}](#type-position)
  - Response: [{Range Object}](#type-range) OR [{Error Object}](#type-error)

## POST /hover

Returns hover-over information about the def/ref/etc at the given position.

- LSP equivalent: [textDocument/hover](https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#hover)
- Request:
  - Body: [{Position Object}](#type-position)
  - Response: [{Hover Object}](#type-hover) OR [{Error Object}](#type-error)

## POST /local-refs

Used for resolving references to repository-local definitions.

- LSP equivalent: [textDocument/references](https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#find-references)
- Request:
  - Body: [{Position Object}](#type-position)
  - Response: [{RefLocations Object}](#type-reflocations) OR [{Error Object}](#type-error)

## POST /external-refs

Used for listing defs used in a repository but defined outside of it.

- LSP equivalent: [workspace/symbol (not really)](https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#workspace-symbols)
- Request:
  - Body: [{RepoRev Object}](#type-reporev)
  - Response: [{ExternalRefs Object}](#type-externalrefs) OR [{Error Object}](#type-error)

## POST /defspec-refs

Used for listing refs of a definition defined in a repository. If repository in def spec is same as target repository, indicates find local references of a definition.

- LSP equivalent: [workspace/symbol](https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#workspace-symbols)
- Request:
  - Body: [{DefSpec Object}](#type-defspec)
  - Response: [{RefLocations Object}](#type-reflocations) OR [{Error Object}](#type-error)

## POST /exported-symbols

Used for listing defs in a repository that can be used externally.

- LSP equivalent: [workspace/symbol](https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#workspace-symbols)
- Request:
  - Body: [{RepoRev Object}](#type-reporev)
  - Response: [{ExportedSymbols Object}](#type-exportedsymbols) OR [{Error Object}](#type-error)

# Data Types

Below are all of the JSON data types which the API uses. The syntax used below to describe the objects is pseudo-JSON for readability purposes, but the actual objects are always JSON-encoded.

## Type: Error

`{Error Object}` is simply a JSON object with one field (an error message). It is returned in the event of any request error. If the request was invalid, HTTP status 400 Bad Request should be returned. Otherwise, if the error is unexpected, HTTP 500 Internal Server Error should be returned.

```
{
    // Error, if any, specifies that there was an error serving the request.
    Error: "",
}
```

## Type: Position

`{Position Object}` is a JSON object representing a single specific position within a file located in a repository at a given revision.

```
{
    // Repo is the repository URI in which the file is located.
    Repo: "github.com/gorilla/mux",

    // Commit is the Git commit ID (not branch) of the repository.
    Commit: "6ba0a61e811b8b846aacc704d5cb306b71c5d858",

    // File is the file which the user is viewing, relative to the repository root.
    File: "mux.go",

    // Line is the line number in the file (zero based), e.g. where a user's cursor
    // is located within the file.
    Line: 1,

    // Character is the character offset on a line in the file (zero based), e.g.
    // where a user's cursor is located within the file.
    Character: 30,
}
```

## Type: Range

`{Range Object}` represents a specific range within a file.

```
{
    // Repo is the repository URI in which the file is located.
    Repo: "github.com/gorilla/mux",

    // Commit is the Git commit ID (not branch) of the repository.
    Commit: "6ba0a61e811b8b846aacc704d5cb306b71c5d858",

    // File is the file which the user is viewing, relative to the repository root.
    File: "mux.go",

    // StartLine is the starting line number in the file (zero based), i.e.
    // where the range starts.
    StartLine: 1,

    // EndLine is the ending line number in the file (zero based), i.e. where
    // the range ends.
    EndLine: 1,

    // StartCharacter is the starting character offset on the starting line in
    // the file (zero based).
    StartCharacter: 30,

    // EndCharacter is the ending character offset on the ending line in the
    // file (zero based).
    EndCharacter: 30,
}
```

## Type: RefLocations

`{RefLocations Object}` represents references to a specific definition.

```
{
    // Refs is a list of references to a definition defined within the requested
    // repository.
    Refs: [Array of {Range Object}]
}
```

See also: [{Range Object}](#type-range)

## Type: Hover

`{Hover Object}` is a JSON object representing a message for when a user “hovers” over a definition. It is a human-readable description of a definition.

```
{
    Contents: [Array of {HoverContent Object}]
}
```

See also: [{HoverContent Object}](#type-hovercontent)

## Type: HoverContent

`{HoverContent Object}` represents a subset of the content for when a user “hovers” over a definition. For example, one HoverContent object may represent the comments of a function, while the another HoverContent object may represent the function signature. In the future we may abuse this field to carry more data, and thus we use “type” instead of “language” like in LSP. In practice at this point, it always maps to a language (Go, Java, etc).

```
{
    Type: "Go",
    Value: "func NewRequest() *Request",
}
```

## Type: RepoRev

`{RepoRev Object}` represents a repository at a specific commit.

```
{
    // Repo is the repository URI.
    Repo: "github.com/gorilla/mux",

    // Commit is the Git commit ID (not branch) of the repository.
    Commit: "6ba0a61e811b8b846aacc704d5cb306b71c5d858",
}
```

## Type: ExternalRefs

`{ExternalRefs Object}` is a JSON object containing an array of all definitions used by a repository but defined in other repositories.

```
{
    Defs: [Array of {DefSpec Object}]
}
```

See also: [{DefSpec Object}](#type-defspec)

## Type: ExportedSymbols

`{ExportedSymbols Object}` is a JSON object containing an array of all definitions defined by a repository.

```
{
    Defs: [Array of {DefSpec Object}]
}
```

See also: [{DefSpec Object}](#type-defspec)

## Type: DefSpec

`{DefSpec Object}` is a globally unique identifier for a definition in a repository at a specific revision. It is the same as the Srclib DefSpec.

```
{
    // Repo is the repository URI in which the definition is located.
    Repo: "github.com/gorilla/mux",

    // Commit is the Git commit ID (not branch) of the repository.
    Commit: "6ba0a61e811b8b846aacc704d5cb306b71c5d858",
    
    // UnitType
    UnitType: "GoPackage",
    
    // Unit
    Unit: "github.com/gorilla/mux",
    
    // Path
    Path: "NewRouter",
}
```

## Type: Symbol

`{Symbol Object}` represents information on a symbol in code.

```
{
    // DefSpec is the DefSpec for this symbol.
    DefSpec: {DefSpec Object}

    // Name of the symbol. This need not be unique.
    Name: "NewRouter",

    // Kind is the kind of thing this definition is. This is
    // language-specific. Possible values include "type", "func",
    // "var", etc.
    Kind: "func",

    // File is the path to the file containing the symbol.
    File: "mux.go",

    // DocHTML is the docstring for the symbol, in the format 'text/html'.
    //
    // Note: You can't assume DocHTML has already been sanitized.
    DocHTML: "\u003cp\u003e\nNewRouter returns a new router instance.\n\u003c/p\u003e"
}
```

See also: [{DefSpec Object}](#type-defspec)
