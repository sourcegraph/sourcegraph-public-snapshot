# unused

_unused_ checks Go code for unused constants, variables, functions and
types.

## Install

	go get honnef.co/go/tools/cmd/unused

## Usage

	unused -help

## Usage Tips

- When running _unused_ on multiple packages, it will first try to
  check them all at once, because that's faster. If any of the
  packages don't compile, however, _unused_ will check each package
  individually.

  The first step can, depending on the number of packages, use a lot
  of memory. For the entire standard library, it uses roughly 800 MB.
  For a GOPATH with thousands of packages, it can quickly use several
  gigabytes. If that is an issue, consider using something like this
  instead:

  ```
  for pkg in $(go list your_selection); do unused "$pkg"; done
  ```

  This will effectively skip the first step and always check every
  package individually.

## What counts as used/unused?

_unused_ checks for unused constants, functions, types and optionally
struct fields. They will be considered used or unused under the
following conditions:

- Unexported package-level objects will be reported as unused if there
  are no explicit references to them.

- Unexported methods will be reported as unused if there are no
  explicit references to them and if they don't implement any
  interfaces.

- The `main` function is considered as used if it's in the `main`
  package.

- `init` functions are always considered as used.

- Exported objects in function scope are treated like unexported
  objects.

- Exported functions in tests are treated like unexported functions,
  unless they're test, benchmark or example functions.

- Struct fields will be considered as unused if there are no explicit
  references to them. Unkeyed composite literals with >=1 elements
  mark all fields of the struct as used.

- Neither the checks for methods nor for struct fields are aware of
  the reflect package and may thus produce false positives.

## Whole program analysis

Optionally via the `-exported` flag, _unused_ can analyse all
arguments as a single program and report unused exported identifiers.
This can be useful for checking "internal" packages, or large software
projects that do not export an API to the public, but use exported
methods between components.

Do note that in the whole-program analysis, all arguments must
type-check. It is not possible to check packages individually in this
mode.

## Examples

```
$ time unused cmd/go
/usr/lib/go/src/cmd/go/build.go:1327:6: func hasString is unused
/usr/lib/go/src/cmd/go/build.go:2328:6: func toolVerify is unused
/usr/lib/go/src/cmd/go/generate.go:375:21: func identLength is unused
/usr/lib/go/src/cmd/go/get.go:474:5: var goTag is unused
/usr/lib/go/src/cmd/go/get.go:513:6: func cmpGoVersion is unused
/usr/lib/go/src/cmd/go/go_test.go:426:23: func grepCountStdout is unused
/usr/lib/go/src/cmd/go/go_test.go:432:23: func grepCountStderr is unused
/usr/lib/go/src/cmd/go/main.go:406:5: var logf is unused
/usr/lib/go/src/cmd/go/main.go:431:6: func runOut is unused
/usr/lib/go/src/cmd/go/pkg.go:91:2: field forceBuild is unused
/usr/lib/go/src/cmd/go/pkg.go:688:2: const toRoot is unused
/usr/lib/go/src/cmd/go/testflag.go:278:6: func setIntFlag is unused
unused cmd/go  3.33s user 0.25s system 447% cpu 0.799 total
```

```
$ time unused $(go list github.com/prometheus/prometheus/... | grep -v /vendor/)
/home/dominikh/prj/src/github.com/prometheus/prometheus/promql/engine_test.go:11:5: var noop is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/retrieval/discovery/dns.go:39:2: const interval is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/retrieval/discovery/dns.go:69:2: field m is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/retrieval/discovery/nerve.go:31:2: const nerveNodePrefix is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/retrieval/discovery/serverset.go:33:2: const serversetNodePrefix is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/retrieval/scrape.go:41:2: const ingestedSamplesCap is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/retrieval/scrape.go:49:2: var errSkippedScrape is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/retrieval/targetmanager.go:184:2: field providers is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/storage/local/delta.go:394:2: field error is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/storage/local/delta.go:398:3: field error is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/storage/local/doubledelta.go:500:2: field error is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/storage/local/doubledelta.go:504:3: field error is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/storage/remote/opentsdb/client.go:40:2: var illegalCharsRE is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/util/stats/timer.go:56:2: field child is unused
/home/dominikh/prj/src/github.com/prometheus/prometheus/util/treecache/treecache.go:25:2: field zkEvents is unused
unused $(go list github.com/prometheus/prometheus/... | grep -v /vendor/)  5.70s user 0.43s system 535% cpu 1.142 total
```

```
$ time unused -exported github.com/kr/pretty/...
/home/dominikh/prj/src/github.com/kr/pretty/formatter.go:14:2: const limit is unused
/home/dominikh/prj/src/github.com/kr/pretty/formatter.go:322:6: func tryDeepEqual is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:20:6: func Errorf is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:28:6: func Fprintf is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:37:6: func Log is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:45:6: func Logf is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:54:6: func Logln is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:63:6: func Print is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:71:6: func Printf is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:80:6: func Println is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:88:6: func Sprintf is unused
/home/dominikh/prj/src/github.com/kr/pretty/pretty.go:92:6: func wrap is unused
unused -exported github.com/kr/pretty/...  1.23s user 0.19s system 253% cpu 0.558 total
```
