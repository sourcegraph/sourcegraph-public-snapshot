# Adapting a language server for use with Sourcegraph'

Servers that speak the Language Server Protocol (LSP) can provide [code intelligence](index.md) everywhere that you view code (on Sourcegraph, on GitHub and other code hosts) by creating a [Sourcegraph extension](https://github.com/sourcegraph/sourcegraph-extension-api).

This documentation page is intended for language server developers who are adapting an existing language server (or building a new language server) to provide code intelligence on Sourcegraph, because there are a few extra considerations when running language servers in a distributed service environment like Sourcegraph. 

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

## Authentication

To make sure the language server is authenticated when fetching URIs, the extension can authenticate the URL by embedding the users session token into the userinfo section of the URI:

```
          userinfo        host
          ┌─┴────────┐ ┌────┴────────┐
  https://sessiontoken:sourcegraph.com/.api/raw/github.com/ReactiveX/rxjs.tar#src/Observable.ts
  └─┬─┘ └───────┬────────────────────┘└─┬───────────────────────────────────┘└──┬─────────────┘
  scheme     authority                 path                                    fragment
```

## Deployment

Language servers can be deployed anywhere as long as the extension has a URL to connect to.
The README of the extension typically provides examples to deploy the corresponding language server.
