# xlang notes

## Shocking things!

* Everything is done in memory (except reading source files from GOROOT for the Go langserver at the moment).
  * An entire repo "clone" plus in-memory analysis of docker/machine takes 1.3s to complete, end-to-end, and ~27 MB of memory to keep resident.
* There is no shared state other than what is passed over the JSON-RPC connection between Sourcegraph and the build/lang server.. For example, all of the files that the lang server needs are sent to it using the LSP textDocument/didOpen request. This seems crazy, but it's fast. (There is a benchmark that shows the speed; it's around 15 MB/s unoptimized. For comparison, all of the .go files in the Go stdlib amount to 5 MB, so we can "clone" the Go stdlib in 300msec.)
  * To integrate a new build/lang server, Sourcegraph just needs a socket address (TCP, Unix socket, or even stdio will suffice), and it doesn't need any control over the environment. No shared state!
* Disk is used as a cache for repos, but the system is fast even when it needs to fetch them all on the fly. It fetches GitHub repos (which is where most Go repos live) using GitHub's zip archive APIs and keeps them in memory.
* Deps are fetched on the fly based on whatever the currently viewed source file needs. For Go, this is easy; I use code from gddo.

Changes from current system:
* The Go lang server now uses go/types (go/loader) instead of shelling out to `godef`, etc.
* The Go lang processor is now called the Go build server, and the Go build and lang servers both speak standard LSP now.
* Sourcegraph does not need to set up the file system or environment on the build/lang server anymore, beyond sending it LSP messages.
* Monaco speaks LSP in HTTP request bodies now.

Try it out:

1. Run `go test ./xlang -v` and see it run integration tests that perform full clones, dep fetches, and analysis of several real-world, large Go repositories. (Run with `-short` to just run the non-integration tests.)
1. Run `localStorage.xlang=true` in a JavaScript console on http://localhost:3080 in your browser. You can verify it's set by seeing if the hovers issue `/hover-info` HTTP requests (xlang is NOT enabled) or `/xlang/initialize`, `/xlang/hover` requests (xlang IS enabled).
1. Go to http://localhost:3080/github.com/golang/go@6129f37367686edf7c2732fbb5300d5f28203743/-/blob/src/strings/strings.go#L250 in your browser and mouse over some symbols. The full source build and analysis of the Go stdlib takes ~3.5s.
1. Go to http://localhost:3080/github.com/coreos/fuze@7df4f06041d9daba45e4c68221b9b04203dff1d8/-/blob/config/convert.go#L48 in your browser. Mouse around.
1. OK, well those don't have external deps...so let's try one that does: http://localhost:3080/github.com/fsouza/go-dockerclient@1123a1e9fcff4684f9ec2f488a430f8fefe5fab1/-/blob/container.go#L143.
1. If you want to start from scratch, `rm -rf /tmp/xlang-git-clone-cache /tmp/github-cache` and restart the `make serve-dev` process. No state is retained when you start up again.

Note: Go-to-def and hover are implemented for Go only right now. Find Local References is not implemented.

To dig into the code:

1. The Go bit is really 2 "LSP" servers: a Go build server, which wraps (and communicates LSP to/from) a Go language server. Read the Architecture section for more about those. The only deviation from LSP is that the Go build server sends some extra GOPATH, GOROOT, etc., initialize parameters to the Go lang server.
1. The heavy lifting in the LSP proxy is in xlang/client_proxy.go (which services/httpapi/xlang.go multiplexes web clients onto) and xlang/server_proxy.go (which manages build servers and routes client requests to the right build servers).

Diagram:

  Web client --> LSP data request in an HTTP POST body --> services/httpapi/xlang.go handler --> xlang/client_proxy.go --> xlang/server_proxy.go --> Go build server --> Go lang server

# Architecture

Each intermediate layer proxies the layer below it.

* Proxy: in main process (not necessary to be in the same process in production, but it's a small and general service)
* Build server: separate server (which is aware of things like dependencies and file system path structure for Go projects)
* Language server: separate server (which does typechecking, hover, go-to-def, etc.; the basic LSP stuff)

The build and language servers are often run in the same process space. Sourcegraph itself just communicates with the build server (which, for simple language support, can just be a language server directly).

## Example: life of a gotodef request

First, the user clicks to go-to-definition on "render" in the expression "ReactDOM.render" in file ui/web_modules/sourcegraph/editor/Editor.tsx line 10 column 20 at rev master in repo github.com/sourcegraph/sourcegraph. The `ui/` directory contains a TypeScript project. (The xlang branch only supports Go, not TypeScript, right now, but I chose TypeScript to make it a bit more illustrative. Go is easy.)

The client (web browser) sends a series of LSP requests to the server:

1. initialize rootPath=git://github.com/sourcegraph/sourcegraph?master
1. textDocument/definition textDocument=git://github.com/sourcegraph/sourcegraph?master#ui/web_modules/sourcegraph/editor/Editor.tsx position=10:20

The HTTP API receives these requests and checks that the user can access the repo. It sends the following requests to the proxy:

* `--> initialize rootPath=git://github.com/sourcegraph/sourcegraph?master` mode=typescript
* `--> textDocument/definition textDocument=git://github.com/sourcegraph/sourcegraph?master#ui/web_modules/sourcegraph/editor/Editor.tsx position=10:20`

The proxy receives these requests and determines which workspace root directory the file is in (always the root `/` for now) and which build/language server to use (using the non-standard "mode" on the LSP initialize request).

(Future: To determine the workspace root directory in the future, it will use a list of rules that are provided by the build servers that have registered with it: For example, the TypeScript language server can register that the presence of a tsconfig.json file indicates that the subtree is a TypeScript workspace. Because the file ui/tsconfig.json exists in the repo, the workspace is the "ui" subtree. But let's say it's at the root for now, which will work fine.)

The proxy sends the following requests to the build server:

* `--> initialize rootPath=file:///`
* `--> textDocument/didOpen uri=file:///ui/web_modules/sourcegraph/editor/Editor.tsx text=...` requests for every single file in the repo that the build/lang servers could possibly need (e.g., all `.go` files), all rooted at `file:///`
* `--> textDocument/definition textDocument=file:///ui/web_modules/sourcegraph/editor/Editor.tsx position=10:20`

The build server receives these requests and stores the files received in `textDocument/didOpen` requests in an in-memory virtual file system (or on disk if absolutely necessary).

It then determines the configuration and environment in which to run the language server. It scans the tsconfig.json file to see that the TypeScript SDK is version 2.0, and it performs dependency analysis to determine that the Editor.tsx file depends on 13 other files in the workspace, plus files from 5 external npm packages (that, unlike in reality, are not committed to the repository). Providing only certain dependencies' files is a huge win for performance and resource consumption.

The build server then sends the following requests to the language server:

* `--> initialize rootPath=file:/// typescript.sdk=2.0`
* `--> textDocument/definition textDocument=file:///ui/web_modules/sourcegraph/editor/Editor.tsx position=10:20`

The language server receives these requests and behaves as a standard LSP server. It returns the following response to the build server:

* `<-- textDocument/definition Location: uri=file:///ui/node_modules/react/lib/ReactClass.js range=30:40-45`

The build server receives this response and notices that it refers to a path inside an npm package (react@15.1.0). It performs some analysis on the definition location and then returns the following response to the proxy:

* `<-- textDocument/definition Location[]: [uri=file:///ui/node_modules/react/lib/ReactClass.js range=30:40-45, uri=npm://react@15.1.0?file=lib/ReactMount.js&range=30:40-45&name=render&containerName=ReactMount]` (note the addition of the second Location)

The proxy receives this response and converts any `file:///` URIs back to `git://` URLs and sends the following response to the client:

* `<-- textDocument/definition Location[]: [uri=git://github.com/sourcegraph/sourcegraph?master#ui/node_modules/react/lib/ReactClass.js range=30:40-45, uri=npm://react@15.1.0?file=lib/ReactMount.js&range=30:40-45&name=render&containerName=ReactMount]` (note that the first Location is now a `git://` URL)

The client gives the user the option of jumping to either the locally vendored definition (even if it's not checked into the Git repo) or the definition in the npm package.

## Example: indexing for global defs/refs

See the above gotodef example for the format of the Location `uri` field for references to external definitions. To build a global (cross-repository) index mapping every definition to its external references, we can perform the following steps.

First. collect the results of running a "textDocument/definition" LSP request (which is Sourcegraph-specific) on every identifier in every file in the default branch of every repository. (Future optimization: create a new "workspace/references" LSP method.) Filter these results to only those entries that refer to definitions in an external repository.

Next, build a reverse index of the abstract def locations to all of the concrete locations that refer to them:

```
npm://react@15.1.0?file=lib/ReactMount.js&range=30:40-45&name=render&containerName=ReactMount
  <- uri=git://github.com/sourcegraph/sourcegraph?master#ui/web_modules/sourcegraph/editor/Editor.tsx position=10:20
  <- uri=git://github.com/sourcegraph/sourcegraph?master#ui/web_modules/sourcegraph/repo/RepoMain.tsx position=50:60
  ...

npm://lodash@4.15.0/?file=partition.js&range=70:80-85&name=partition
  <- uri=git://github.com/reactjs/redux?master#src/compose.js position=60:20
  ...

...
```

To look up all references to a given def, consult this index. The index is built such that a query of the form `npm://react?name=render&containerName=ReactMount` matches the first entry in the example index above (and any similar variants, such as a different version number or character range).

Let's walk through this from the user's POV. Suppose you're browsing the github.com/facebook/react repository and want to see global refs to ReactMount.render. You place your cursor on the definition of ReactMount.render and choose "Find External References." The build server knows that your current file is in a workspace that produces an npm package named "react" (because it scans the package.json file), and a quick AST scan of the file lets us produce the querystring `name=render&containerName=ReactMount`.


# Design

(The rest of this doc is probably obvious to most people reading it. You can skip it.)

A repo has many workspaces. For example:

* The sourcegraph repo has the main Go workspace, plus the JavaScript workspace in `./ui`.
* The github.com/Microsoft/vscode-languageserver-node repo has separate JavaScript workspaces `client`, `jsonrpc`, `server`, and `types`, each of which publish their own separate (but interdependent) npm packages.

A workspace is identified by:

  repo + rev + (sub)path + config

where:

* repo is a string like "github.com/a/b"
* rev refers to a revision, such as a branch or commit ID (or, in the future, a scratch area for collaborating WIP changes)
* subpath is either the root (".") or the path to a subdirectory in the repo (such as "ui" for the Sourcegraph JavaScript UI code)
* config is 

We make a few assumptions to make workspaces immutable and reusable (although in the future we will relax these assumptions):

* users can't change files in the workspace directly; the initial workspace files are taken solely from the repo (e.g., the LSP textDocument/didOpen can't be used to substitute user-supplied contents for a file in the repo)
* the workspace's language server can only access resources (such as private dependencies) that are available to    anyone who has read access to the repo (i.e., the language server doesn't assume the privileges of the current or initial user)

Note: These aren't yet being enforced in the xlang branch!

## Building workspaces

Language servers do not deal with installing dependencies, compiling code, and other build tasks. But these steps are required for the language server to work. This presents a few challenges:

* How can we determine what steps must be run to build the workspace?
* How can we execute the build steps:
  * securely (ideally, without running untrusted code)
  * quickly (e.g., without taking 20 minutes to download 1 GB of JARs each time)
  * efficiently (e.g., how do we reuse the dependencies we fetched from a similar commit and avoid doing that work all over again?)
  * with the correct level of access (e.g., if a Java project depends on JARs hosted in an internal Artifactory behind a VPN, how can we fetch the JARs?)

VSCode uses the `.vscode` directory to specify how a workspace should be built, compiled, etc. See https://github.com/Microsoft/vscode/tree/master/.vscode for an example. Our solution should be conceptually compatible with this.

#### Separation of build tasks and language analysis

Why did Microsoft choose to use `.vscode` for build tasks and LSP for language analysis? Build and analysis initially seem like they're related and should be handled by the same component of any system we build. But they aren't.

* A single language can have any number of build tools with very different behavior (e.g., Ant, Maven, Gradle, and multiple versions of each), but compilers usually work nearly identically (by design).
* They never need to communicate synchronously. Their sole communication channel is the file system. So, they can be run sequentially and their results can be cached in lieu of rerunning them.

Because of this, we should keep build and language analysis separate (even if they run in the same process for speed).

### Build security

Key goal: don't run untrusted code. We're OK compromising on correctness to avoid running untrusted code.

Usually dependency installation is where you must run untrusted code (e.g., shelling out to `mvn install`, `npm install`, etc.). These tools all basically just download and unpack archives from the web into a certain directory. Most of them also expose their functionality via a library API, not just through a CLI tool. So, we will write something that programmatically returns a list of files and their contents given a pom.xml, package.json, etc., file.

For Go, this is easy; we just use go/types, go/loader, etc. It's doable for other languages as well. Most people who write language package managers are dev tools folks, who understand the importance of making functionality available as a library. Also, we have experience doing dependency resolution using these libraries from srclib.

There will be some cases in which we must run untrusted code, but I don't think that will be necessary for the first several languages we need to support. When it becomes necessary, there are other workarounds we can do, like giving people an easy way to upload their local build products from CI or from their own machines to supplement Sourcegraph's analysis. (This works around the VPN issue as well.)

### Build speed

Key goal: keep as much in memory as possible, and cache results that can be shared across runs (or even across workspaces).

Avoiding executing code means the build can be performed in memory in the same process and can use paths on disk that are only isolated from each other at the application level (i.e., containers or VMs are not necessary). Win!

We will also build a service that makes it super fast to get external dependencies. For example, you could ask it for a specific JAR from Maven Central or a snapshot of a Go package (including only the files that you need, not all the Git history and unrelated files).

### Build efficiency

Similar to build speed. As long as we can cache results, we're good here.

### Build access to external resources

We need to make it so either:

* you can push private external resources to Sourcegraph from your local workspace or from CI
* Sourcegraph can access resources behind your VPN

We also need to make it easy to debug, so that you can see which external resources Sourcegraph is using when it performs the build.

For now, as described above, we can release a CLI tool for developers to upload build products from their own machines or from CI.

# Adding new build/language servers

## Creating a new build/lang server

* The big thing is that it should not depend on the file system. It's fine to use the file system as an overlay, but it should accept files and their contents sent in `textDocument/didOpen` LSP requests. Sourcegraph sends every relevant file to your server.

  If you truly can't make your server work in-memory, then you can write the `textDocument/didOpen` file contents to a temporary directory. (But check with other people first to make sure there isn't a better solution.)
* All file URIs will be rooted at `file:///`, so if the repo has a `a.js` and `b/c.js`, your server will see `textDocument/didOpen` requests with URIs `file:///a.js` and `file:///b/c.js`.
* If you want to emit a cross-repository reference, you can emit an LSP `Location` with a URI of the form `git://github.com/foo/bar?HEAD#path/to/the/file.txt`. The `?HEAD` is there to be explicit that you don't know which commit ID to refer to. (See the rest of this document for how we will do smarter, non-positional cross-repository references. But this kind will work for now, and I don't want anyone to block on stuff we haven't nailed down with Universe yet.)

Otherwise, just speak LSP and it'll work with both Sourcegraph and VS Code!

## Hooking up a build/lang server to Sourcegraph

1. Figure out the mode ID of your language. Consult [the language ID table](https://code.visualstudio.com/docs/languages/overview#_language-id), or figure out what existing VSCode extensions are using. If you can't figure it out, just make something up and be consistent.
1. Make sure our temporarily hacky `getModeByFilename` func in `EditorService.tsx` includes a branch for the language/mode with the proper file extensions.
1. In the `sourcegraph/editor/Editor.tsx` file's `Editor` constructor, add an if-condition branch for the mode ID of the language so that the hover/def/ref providers are registered.
1. Run Sourcegraph (`make serve-dev` or `src`) with environment variables of the form `LANGSERVER_xyz=addr-or-program`, where `xyz` is the mode (e.g., `typescript`) and `addr-or-program` is either `tcp://addr:port`, `unix:///path/to/socket`, or `/path/to/my/executable/program/that/speaks/over/stdio`.
1. Open up Sourcegraph to a file in the language, and it will work. Check the JavaScript console for the LSP requests and responses.

Here's what I am running with:

```
LANGSERVER_JAVASCRIPT=tcp://localhost:2089 LANGSERVER_TYPESCRIPT=tcp://localhost:2089 SG_UNIVERSE_REPO=all SG_FEATURE_NOSRCLIB=t SG_FEATURE_UNIVERSE=t make serve-dev
```

Notes:

1. `typescript` and `javascript` are separate modes in Monaco, so you'll have to register the JS/TS lang server twice, one for each mode (see the above example).
1. Until the JS/TS lang server supports running in stdio, you need to register its TCP listen address, and you need to run it in a separate terminal (see its README.md for more info).
1. The Go build/lang server currently runs in-process for simplicity, so you needn't register it. That will change soon.
1. All you need to provide Sourcegraph is a single entrypoint. If you have separate build and lang servers (see the rest of this doc for the distinction), then you still only give Sourcegraph one entrypoint, since the build server wraps the lang server.
1. VSCode/Monaco have a concept of "modes," which is basically a language. A mode's ID is a string like `go`, `javascript`, `python`, `typescript`, etc. We will reuse this existing taxonomy.


# Further work

* Figure out what https://github.com/Microsoft/vscode/blob/master/src/vs/buildunit.json is for. Is it for statically declaring the products of this workspace (like Google's Blaze/Bazel)? Probably need to ask someone at/inside MSFT.
