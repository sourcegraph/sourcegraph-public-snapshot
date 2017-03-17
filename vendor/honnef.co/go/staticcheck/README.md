Staticcheck is `go vet` on steroids, applying a ton of static analysis
checks you might be used to from tools like ReSharper for C#.

**This project is supported by
[patrons on Patreon](https://www.patreon.com/dominikh). If you, too,
would like to support the project's development, consider
[becoming a patron](https://www.patreon.com/dominikh).**


## Installation

Staticcheck requires Go 1.6 or later.

    go get honnef.co/go/staticcheck/cmd/staticcheck

## Usage

Invoke `staticcheck` with one or more filenames, a directory, or a package named
by its import path. Staticcheck uses the same
[import path syntax](https://golang.org/cmd/go/#hdr-Import_path_syntax) as
the `go` command and therefore
also supports relative import paths like `./...`. Additionally the `...`
wildcard can be used as suffix on relative and absolute file paths to recurse
into them.

The output of this tool is a list of suggestions in Vim quickfix format,
which is accepted by lots of different editors.

## Purpose

The main purpose of staticcheck is editor integration, or workflow
integration in general. For example, by running staticcheck when
saving a file, one can quickly catch simple bugs without having to run
the whole test suite or the program itself.

The tool shouldn't report any errors unless there are legitimate
bugs - or very dubious constructs - in the code.

It is similar in nature to `go vet`, but has more checks that catch
bugs that would also be caught easily at runtime, to reduce the number
of edit, compile and debug cycles.

## Checks

The following things are currently checked by staticcheck:

| Check      | Description                                                                                                                                           |
|------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| **SA1???** | **Various misuses of the standard library**                                                                                                           |
| SA1000     | Invalid regular expression                                                                                                                            |
| SA1001     | Invalid template                                                                                                                                      |
| SA1002     | Invalid format in time.Parse                                                                                                                          |
| SA1003     | Unsupported argument to functions in encoding/binary                                                                                                  |
| SA1004     | Suspiciously small untyped constant in time.Sleep                                                                                                     |
| SA1005     | Invalid first argument to exec.Command                                                                                                                |
| SA1006     | Printf with dynamic first argument and no further arguments                                                                                           |
| SA1007     | Invalid URL in net/url.Parse                                                                                                                          |
| SA1008     | Non-canonical key in http.Header map                                                                                                                  |
| SA1010     | `(*regexp.Regexp).FindAll` called with n == 0, which will always return zero results                                                                  |
| SA1011     | Various methods in the `strings` package expect valid UTF-8, but invalid input is provided                                                            |
| SA1012     | A nil `context.Context` is being passed to a function, consider using context.TODO instead                                                            |
| SA1013     | `io.Seeker.Seek` is being called with the `whence` constant as the first argument, but it should be the second                                        |
| SA1014     | Non-pointer value passed to Unmarshal or Decode                                                                                                       |
| SA1015     | Using `time.Tick` in a way that will leak. Consider using `time.NewTicker`, and only use `time.Tick` in tests, commands and endless functions         |
| SA1016     | Trapping a signal that cannot be trapped                                                                                                              |
| SA1017     | Channels used with signal.Notify should be buffered                                                                                                   |
| SA1018     | `strings.Replace` called with n == 0, which does nothing                                                                                              |
| SA1019     | Using a deprecated function, variable, constant or field                                                                                              |
| SA1020     | Using an invalid `host:port` pair with a `net.Listen`-related function                                                                                |
| SA1021     | Using bytes.Equal to compare two net.IP                                                                                                               |
|            |                                                                                                                                                       |
| **SA2???** | **Concurrency issues**                                                                                                                                |
| SA2000     | `sync.WaitGroup.Add` called inside the goroutine, leading to a race condition                                                                         |
| SA2001     | Empty critical section, did you mean to `defer` the unlock?                                                                                           |
| SA2002     | Called testing.T.FailNow or SkipNow in a goroutine, which isn't allowed                                                                               |
| SA2003     | Deferred Lock right after locking, likely meant to defer Unlock instead                                                                               |
|            |                                                                                                                                                       |
| **SA3???** | **Testing issues**                                                                                                                                    |
| SA3000     | TestMain doesn't call os.Exit, hiding test failures                                                                                                   |
| SA3001     | Assigning to `b.N` in benchmarks distorts the results                                                                                                 |
|            |                                                                                                                                                       |
| **SA4???** | **Code that isn't really doing anything**                                                                                                             |
| SA4000     | Boolean expression has identical expressions on both sides                                                                                            |
| SA4001     | `&*x` gets simplified to `x`, it does not copy `x`                                                                                                    |
| SA4002     | Comparing strings with known different sizes has predictable results                                                                                  |
| SA4003     | Comparing unsigned values against negative values is pointless                                                                                        |
| SA4004     | The loop exits unconditionally after one iteration                                                                                                    |
| SA4005     | Field assignment that will never be observed. Did you mean to use a pointer receiver?                                                                 |
| SA4006     | A value assigned to a variable is never read before being overwritten. Forgotten error check or dead code?                                            |
| SA4008     | The variable in the loop condition never changes, are you incrementing the wrong variable?                                                            |
| SA4009     | A function argument is overwritten before its first use                                                                                               |
| SA4010     | The result of `append` will never be observed anywhere                                                                                                |
| SA4011     | Break statement with no effect. Did you mean to break out of an outer loop?                                                                           |
| SA4012     | Comparing a value against NaN even though no value is equal to NaN                                                                                    |
| SA4013     | Negating a boolean twice (`!!b`) is the same as writing `b`. This is either redundant, or a typo.                                                     |
| SA4014     | An if/else if chain has repeated conditions and no side-effects; if the condition didn't match the first time, it won't match the second time, either |
| SA4015     | Calling functions like math.Ceil on floats converted from integers doesn't do anything useful                                                         |
| SA4016     | Certain bitwise operations, such as `x ^ 0`, do not do anything useful                                                                                |
| SA4017     | A pure function's return value is discarded, making the call pointless                                                                                |
|            |                                                                                                                                                       |
| **SA5???** | **Correctness issues**                                                                                                                                |
| SA5000     | Assignment to nil map                                                                                                                                 |
| SA5001     | Defering `Close` before checking for a possible error                                                                                                 |
| SA5002     | The empty `for` loop (`for {}`) spins and can block the scheduler                                                                                     |
| SA5003     | Defers in infinite loops will never execute                                                                                                           |
| SA5004     | `for { select { ...` with an empty default branch spins                                                                                               |
| SA5005     | The finalizer references the finalized object, preventing garbage collection                                                                          |
| SA5006     | Slice index out of bounds                                                                                                                             |
| SA5007     | Infinite recursive call                                                                                                                               |
|            |                                                                                                                                                       |
| **SA9???** | **Dubious code constructs that have a high probability of being wrong**                                                                               |
| SA9000     | Storing non-pointer values in sync.Pool allocates memory                                                                                              |
| SA9001     | `defer`s in `for range` loops may not run when you expect them to                                                                                     |
| SA9002     | Using a non-octal `os.FileMode`  that looks like it was meant to be in octal.                                                                         |

## Ignoring checks

staticcheck allows disabling some or all checks for certain files. The
`-ignore` flag takes a whitespace-separated list of
`glob:check1,check2,...` pairs. `glob` is a glob pattern matching
files in packages, and `check1,check2,...` are checks named by their
IDs.

For example, to ignore assignment to nil maps in all test files in the
`os/exec` package, you would write `-ignore
"os/exec/*_test.go:SA5000"`

Additionally, the check IDs support globbing, too. Using a pattern
such as `os/exec/*.gen.go:*` would disable all checks in all
auto-generated files in the os/exec package.

Any whitespace can be used to separate rules, including newlines. This
allows for a setup like the following:

```
$ cat stdlib.ignore
sync/*_test.go:SA2001
testing/benchmark.go:SA3001
runtime/string_test.go:SA4007
runtime/proc_test.go:SA5004
runtime/lfstack_test.go:SA4010
runtime/append_test.go:SA4010
errors/errors_test.go:SA4000
reflect/all_test.go:SA4000

$ staticcheck -ignore "$(cat stdlib.ignore)" std
```
