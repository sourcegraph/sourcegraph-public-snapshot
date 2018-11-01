# Adapting a language server for use with Sourcegraph'

Sourcegraph provides [code intelligence](index.md) through extensions. One way for extensions to provide code intelligence is to connect to a language server that speaks the Language Server Protocol (LSP) standard. However, there are a few assumptions that a language server targeting editor clients may make that are not true in a distributed service environment like Sourcegraph. This documentation page is intended for language server developers who are adapting an existing language server (or building a new language server) to provide code intelligence on Sourcegraph.

## Connection protocol

The easiest way for an extension to talk from the browser to a language server is through WebSockets.

## File system access

LSP uses URIs to identify text documents. In an editor environment, these are usually `file:` URLs that the language server parses to read file and directory contents from disk.
When deploying a language server for a Sourcegraph extension, the language server does not have all the files of the repository on the local file system by default.
The extension has multiple options to access the file contents through the Sourcegraph API and can choose whatever is the most suitable for the language server.

### Using the raw archive API

The TypeScript extension for example sends a `rootUri` pointing to the HTTP archive endpoint of the Sourcegraph raw API, with the session token encoded in the auth section of the URL:

```url
https://sessiontoken:sourcegraph.com/.api/raw/github.com/ReactiveX/rxjs.tar
```

The TypeScript server is able to fetch this archive and extract it to disk on initialization.
In further requests, the files are then identified as contents of the archive:

```url
https://sessiontoken:sourcegraph.com/.api/raw/github.com/ReactiveX/rxjs.tar#src/Observable.ts
```

This method is easy to implement and compatible with many language servers.

### Using the raw contents API

The xxx extension sends URLs pointing to the Sourcegraph raw API for individual files:

```url
https://sessiontoken:sourcegraph.com/.api/raw/github.com/ReactiveX/rxjs/-/src/Observable.ts
```

The server can retrieve the contents of the file with a simple HTTP GET.

A `GET` on a directory returns the entries in that directory separated by newlines (LF):

```
GET https://sessiontoken:sourcegraph.com/.api/raw/github.com/ReactiveX/rxjs/-/src
```

```
Content-Type: text/plain

https://sessiontoken:sourcegraph.com/.api/raw/github.com/ReactiveX/rxjs/-/src/operators
https://sessiontoken:sourcegraph.com/.api/raw/github.com/ReactiveX/rxjs/-/src/Observable.ts
```

### Authentication

To make sure the language server is authenticated when fetching URIs, the extension can authenticate the URL by embedding the users session token into the userinfo section of the URI:

```
          userinfo        host
          ┌─┴────────┐ ┌────┴────────┐
  https://sessiontoken:sourcegraph.com/.api/raw/github.com/ReactiveX/rxjs.tar#src/Observable.ts
  └─┬─┘ └───────┬────────────────────┘└─┬───────────────────────────────────┘└──┬─────────────┘
  scheme     authority                 path                                    fragment
```
