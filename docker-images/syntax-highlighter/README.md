# Syntect Server

This is an HTTP server that exposes the Rust [Syntect](https://github.com/trishume/syntect) syntax highlighting library for use by other services. Send it some code, and it'll send you syntax-highlighted code in response.

Technologies:

- [Syntect](https://github.com/trishume/syntect) -> Syntax highlighting of code.
- [Rocket.rs](https://rocket.rs) -> Web framework.
- [Serde](https://serde.rs/) -> JSON serialization / deserialization .
- [Rayon](https://github.com/nikomatsakis/rayon) -> data parallelism for `SyntaxSet` across Rocket server threads.
- [lazy_static](https://crates.io/crates/lazy_static) -> lazily evaluated static `ThemeSet` (like a global).

## Usage

```bash
docker run --detach --name=syntect_server -p 9238:9238 sourcegraph/syntect_server
```

You can then e.g. `GET` http://localhost:9238/health or http://host.docker.internal:9238/health to confirm it is working.

## API

- `POST` to `/` with `Content-Type: application/json`. The following fields are required:
  - `filepath` string, e.g. `the/file.go` or `file.go` or `Dockerfile`, see "Supported file extensions" section below.
  - `theme` string, e.g. `Solarized (dark)`, see "Embedded themes" section below.
  - `code` string, i.e. the literal code to highlight.
- The response is a JSON object of either:
  - A successful response (`data` field):
    - `data` string with syntax highlighted response. The input `code` string [is properly escaped](https://github.com/sourcegraph/syntect_server/blob/ee3810f70e5701b961b7249393dbac8914c162ce/syntect/src/html.rs#L6) and as such can be directly rendered in the browser safely.
    - `plaintext` boolean indicating whether a syntax could not be found for the file and instead it was rendered as plain text.
  - An error response (`error` field), one of:
    - `{"error": "invalid theme", "code": "invalid_theme"}`
    - `{"error": "resource not found", "code": "resource_not_found"}`
- `GET` to `/health` to receive an `OK` health check response / ensure the service is alive.

## Client

[gosyntect](https://github.com/sourcegraph/gosyntect) is a Go package + CLI program to make requests against syntect_server.

## Configuration

By default on startup, `syntect_server` will list all features (themes + file types) it supports. This can be disabled by setting `QUIET=true` in the environment.

## Development

1. [Install Rust **nightly**](https://rocket.rs/guide/getting-started/#installing-rust).
2. `git clone` this repository anywhere on your filesystem.
3. Use `cargo run` to download dependencies + compile + run the server.

## Building

Invoke `cargo build --release` and an optimized binary will be built (e.g. to `./target/release/syntect_server`).

## Building docker image

`./build.sh` will build your current repository checkout into a final Docker image.

## Publishing docker image

Run `./publish.sh` after merging your changes.

## Updating Sourcegraph

Once published, the image version will need to be updated in the following locations to make Sourcegraph use it:

- [`sourcegraph/sourcegraph > dev/syntect_server.sh`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/syntect_server.sh?subtree=true#L26:82)
- [`sourcegraph/sourcegraph > docker-images/syntax-highlighter/build.sh`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/docker-images/syntax-highlighter/build.sh?subtree=true#L9:29)
- [`sourcegraph/sourcegraph > cmd/server/Dockerfile`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/server/Dockerfile?subtree=true#L54:13)
- [`sourcegraph/sourcegraph > sg.config.yaml`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/sg.config.yaml?subtree=true#L206:7)

Additionally, it's worth doing a [search](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+sourcegraph/syntect_server:&patternType=literal) for other uses in case this list is stale.

## Code hygiene

- Use `cargo fmt` or an editor extension to format code.

## Adding themes

- Copy a `.tmTheme` file anywhere under `./syntect/testdata` (make a new dir if needed) [in our fork](https://github.com/slimsag/syntect).
- `cd syntect && make assets`
- In this repo, `cargo update -p syntect`.
- Build a new binary.

## Adding languages:

#### 1) Find an open-source `.tmLanguage` or `.sublime-syntax` file and send a PR to our package registry

https://github.com/slimsag/Packages is the package registry we use which holds all of the syntax definitions we use in syntect_server and Sourcegraph. Send a PR there by following [these steps](https://github.com/slimsag/Packages/blob/master/README.md#adding-a-new-language)

#### 2) Update our temporary fork of `syntect`

We use a temporary fork of `syntect` as a hack to get our `Packages` registry into the binary. Update it by creating a PR with two commits like:

- https://github.com/slimsag/syntect/commit/9976d2095e49fd91607026364466cd7b389b938e
- https://github.com/slimsag/syntect/commit/1182dd3bd7c82b6655d8466c9896a1e4f458c71e

#### 3) Update syntect_server to use the new version of `syntect`

Send a PR to this repository [with the result of running `cargo update -p syntect`](https://github.com/sourcegraph/syntect_server/commit/1c72addeac3cb54f2c1a7735e8c4ca75eb16d0b3).

#### 4) Publish a new image, use it in Sourcegraph

Run `./publish.sh` to build and release a new image of `syntect_server`, and then send a PR to the main Sourcegraph repository like [this](https://github.com/sourcegraph/sourcegraph/pull/15634/commits/2b8c2a09ab52dbf840495fe0200abd21619b2856). Once merged, it will automatically rollout to Sourcegraph.com and go in the next Sourcegraph release.

## Embedded themes:

- `InspiredGitHub`
- `Monokai`
- `Solarized (dark)`
- `Solarized (light)`
- `Sourcegraph`
- `Sourcegraph (light)`
- `TypeScript`
- `TypeScriptReact`
- `Visual Studio`
- `Visual Studio Dark`
- `base16-eighties.dark`
- `base16-mocha.dark`
- `base16-ocean.dark`
- `base16-ocean.light`

## Supported file extensions:

- Plain Text (`txt`)
- ASP (`asa`)
- HTML (ASP) (`asp`)
- ASP vb.NET (`vb`)
- HTML (ASP.net) (`aspx`, `ascx`, `master`)
- ActionScript (`as`)
- AppleScript (`applescript`, `script editor`)
- Batch File (`bat`, `cmd`)
- NAnt Build File (`build`)
- C# (`cs`, `csx`)
- C++ (`cpp`, `cc`, `cp`, `cxx`, `c++`, `C`, `h`, `hh`, `hpp`, `hxx`, `h++`, `inl`, `ipp`)
- C (`c`, `h`)
- CMake Cache (`CMakeCache.txt`)
- CMake Listfile (`CMakeLists.txt`, `cmake`)
- ACUCOBOL (``)
- COBOL (`cbl`, `cpy`, `cob`, `dds`, `ss`, `wks`, `pco`)
- OpenCOBOL (``)
- jcl (`jcl`)
- CSS (`css`, `css.erb`, `css.liquid`)
- Cap’n Proto (`capnp`)
- Cg (`cg`)
- Clojure (`clj`, `cljc`, `cljs`, `cljx`, `edn`)
- Coq (`v`)
- Crontab (`crontab`)
- CUDA C++ (`cu`, `cuh`)
- D (`d`, `di`)
- DMD Output (``)
- Dart Doc Comments (``)
- Dart (`dart`)
- Diff (`diff`, `patch`)
- Dockerfile (`Dockerfile`)
- DM (`dm`, `dme`)
- Elixir (EEx) (`ex.eex`, `exs.eex`)
- Elixir (`ex`, `exs`)
- HTML (EEx) (`html.eex`, `html.leex`)
- Regular Expressions (Elixir) (`ex.re`)
- SQL (Elixir) (`ex.sql`)
- Elm (`elm`)
- Erlang (`erl`, `hrl`, `Emakefile`, `emakefile`, `escript`)
- HTML (Erlang) (`yaws`)
- Solidity (`sol`)
- Vyper (`vy`)
- F Sharp (`fs`)
- friendly interactive shell (fish) (`fish`)
- Forth (`frt`, `fs`)
- ESSL (`essl`, `f.essl`, `v.essl`, `_v.essl`, `_f.essl`, `_vs.essl`, `_fs.essl`)
- GLSL (`vs`, `fs`, `gs`, `vsh`, `fsh`, `gsh`, `vshader`, `fshader`, `gshader`, `vert`, `frag`, `geom`, `tesc`, `tese`, `comp`, `glsl`)
- Git Attributes (`attributes`, `gitattributes`, `.gitattributes`)
- Git Commit (`COMMIT_EDITMSG`, `MERGE_MSG`, `TAG_EDITMSG`)
- Git Common (``)
- Git Config (`gitconfig`, `.gitconfig`, `.gitmodules`)
- Git Ignore (`exclude`, `gitignore`, `.gitignore`)
- Git Link (`.git`)
- Git Log (`gitlog`)
- Git Mailmap (`.mailmap`, `mailmap`)
- Git Rebase Todo (`git-rebase-todo`)
- Go (`go`)
- GraphQL (`graphql`, `graphqls`, `gql`, `graphcool`)
- Graphviz (DOT) (`dot`, `DOT`, `gv`)
- Groovy (`groovy`, `gvy`, `gradle`, `Jenkinsfile`)
- HLSL (`fx`, `fxh`, `hlsl`, `hlsli`, `usf`)
- HTML (`html`, `htm`, `shtml`, `xhtml`)
- Haskell (`hs`)
- Literate Haskell (`lhs`)
- INI (`cfg`, `conf`, `ini`, `lng`, `url`, `.buckconfig`, `.flowconfig`, `.hgrc`)
- REG (`reg`)
- JSON (`json`, `sublime-settings`, `sublime-menu`, `sublime-keymap`, `sublime-mousemap`, `sublime-theme`, `sublime-build`, `sublime-project`, `sublime-completions`, `sublime-commands`, `sublime-macro`, `sublime-color-scheme`, `ipynb`, `Pipfile.lock`)
- Java Server Page (JSP) (`jsp`)
- Java (`java`, `bsh`)
- Javadoc (``)
- Java Properties (`properties`)
- JS Custom - Default (`js`, `htc`)
- JS Custom - React (`js`, `jsx`)
- Regular Expressions (Javascript) (``)
- JS Custom (Embedded) (``)
- Julia (`jl`)
- Kotlin (`kt`, `kts`)
- LESS (`less`)
- BibTeX (`bib`)
- LaTeX Log (``)
- LaTeX (`tex`, `ltx`)
- TeX (`sty`, `cls`)
- Lisp (`lisp`, `cl`, `clisp`, `l`, `mud`, `el`, `scm`, `ss`, `lsp`, `fasl`)
- Lua (`lua`)
- MSBuild (`proj`, `targets`, `msbuild`, `csproj`, `vbproj`, `fsproj`, `vcxproj`)
- Make Output (``)
- Makefile (`make`, `GNUmakefile`, `makefile`, `Makefile`, `makefile.am`, `Makefile.am`, `makefile.in`, `Makefile.in`, `OCamlMakefile`, `mak`, `mk`)
- Man (`man`)
- Markdown (`md`, `mdown`, `markdown`, `markdn`)
- MultiMarkdown (``)
- MATLAB (`matlab`)
- Maven POM (`pom.xml`)
- Mediawiki (`mediawiki`, `wikipedia`, `wiki`)
- Move (`move`)
- Ninja (`ninja`)
- Nix (`nix`)
- OCaml (`ml`, `mli`)
- OCamllex (`mll`)
- OCamlyacc (`mly`)
- camlp4 (``)
- Objective-C++ (`mm`, `M`, `h`)
- Objective-C (`m`, `h`)
- PHP Source (``)
- PHP (`php`, `php3`, `php4`, `php5`, `php7`, `phps`, `phpt`, `phtml`)
- Regular Expressions (PHP) (``)
- Pascal (`pas`, `p`, `dpr`)
- Perl (`pl`, `pc`, `pm`, `pmc`, `pod`, `t`)
- Property List (XML) (``)
- Postscript (`ps`, `eps`)
- PowerShell (`ps1`, `psm1`, `psd1`)
- Protocol Buffer (`proto`)
- Puppet (`pp`, `epp`)
- Python (`py`, `py3`, `pyw`, `pyi`, `pyx`, `pyx.in`, `pxd`, `pxd.in`, `pxi`, `pxi.in`, `rpy`, `cpy`, `SConstruct`, `Sconstruct`, `sconstruct`, `SConscript`, `pyst`, `pyst-include`, `gyp`, `gypi`, `Snakefile`, `vpy`, `wscript`, `bazel`, `bzl`)
- Regular Expressions (Python) (``)
- R Console (``)
- R (`R`, `r`, `Rprofile`)
- Rd (R Documentation) (`rd`)
- HTML (Rails) (`rails`, `rhtml`, `erb`, `html.erb`)
- JavaScript (Rails) (`js.erb`)
- Ruby Haml (`haml`)
- Ruby on Rails (`rxml`, `builder`)
- SQL (Rails) (`erbsql`, `sql.erb`)
- Regular Expression (`re`)
- reStructuredText (`rst`, `rest`)
- Ruby (`rb`, `Appfile`, `Appraisals`, `Berksfile`, `Brewfile`, `capfile`, `cgi`, `Cheffile`, `config.ru`, `Deliverfile`, `Fastfile`, `fcgi`, `Gemfile`, `gemspec`, `Guardfile`, `irbrc`, `jbuilder`, `Podfile`, `podspec`, `prawn`, `rabl`, `rake`, `Rakefile`, `Rantfile`, `rbx`, `rjs`, `ruby.rail`, `Scanfile`, `simplecov`, `Snapfile`, `thor`, `Thorfile`, `Vagrantfile`)
- Cargo Build Results (``)
- Rust Enhanced (`rs`)
- Sass (`sass`, `scss`)
- SQL (`sql`, `ddl`, `dml`)
- Scala (`scala`, `sbt`, `sc`)
- Bourne Again Shell (bash) (`sh`, `bash`, `zsh`, `ash`, `.bash_aliases`, `.bash_completions`, `.bash_functions`, `.bash_login`, `.bash_logout`, `.bash_profile`, `.bash_variables`, `.bashrc`, `.profile`, `.textmate_init`, `.zlogin`, `.zlogout`, `.zprofile`, `.zshenv`, `.zshrc`, `PKGBUILD`, `.ebuild`, `.eclass`)
- Shell-Unix-Generic (``)
- commands-builtin-shell-bash (``)
- Smalltalk (`st`)
- Smarty (`tpl`)
- Starlark (`build_defs`, `BUILD.in`, `BUILD`, `WORKSPACE`, `bzl`, `sky`, `star`, `BUILD.bazel`, `WORKSPACE`, `WORKSPACE.bazel`)
- Stylus (`styl`, `stylus`)
- Swift (`swift`)
- HTML (Tcl) (`adp`)
- Tcl (`tcl`)
- TOML (`toml`)
- Terraform (`tf`, `tfvars`, `hcl`)
- Textile (`textile`)
- Thrift (`thrift`, `frugal`)
- TypeScript (`ts`)
- TypeScriptReact (`tsx`)
- VimL (`vim`, `.vimrc`, `.gvimrc`)
- Vue Component (`vue`)
- XML (`xml`, `xsd`, `xslt`, `tld`, `dtml`, `rng`, `rss`, `opml`, `svg`)
- YAML (`yaml`, `yml`, `sublime-syntax`)
- Zig (`zig`)
