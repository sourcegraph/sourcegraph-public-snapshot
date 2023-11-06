# scip-treesitter-cli

A command line tool that uses other scip-* crates to either

- produce a SCIP file containing global and local symbols discovered during analysis.
- evaluate one SCIP file against another, producing precision/recall summary

## Usage

**Indexing** (`index --help` for more details)

List of files:

```
scip-treesitter-cli index --language java --out ./index.scip file1.java file2.java ...
```

Entire folder (files discovered according to pre-defined set of extensions):

```
scip-treesitter-cli index --language java --out ./index.scip --workspace <some-folder>
```

**Evaluation** (`scip-evaluate --help` for more details)

```
scip-treesitter-cli scip-evaluate --candidate index-tree-sitter.scip --ground-truth index.scip
```


## Contributing

This is a standard Rust CLI project, with a single runnable entry point - the CLI itself.

1. Run tests: `cargo test`

   We use Insta for snapshot testing, if you're changing the output produced by the CLI,
   run `cargo test` and then `cargo insta review` to accept/reject changes in symbols

2. Run CLI: `cargo run -- index --language java --out ./index.scip file1.java file2.java ...`

On CI, Bazel is used instead of Cargo.

1. Run tests: `bazel test //docker-images/syntax-highlighter/crates/scip-treesitter-cli:unit_test`
2. Build CLI: `bazel build //docker-images/syntax-highlighter/crates/scip-treesitter-cli:scip-treesitter-cli`
