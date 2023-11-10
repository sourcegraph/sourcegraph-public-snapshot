# scip-treesitter-cli

A command line tool that uses other scip-* crates to either

- produce a SCIP file containing global and local symbols discovered during analysis.
- evaluate one SCIP file against another, producing precision/recall summary

## Usage

### Indexing

Index a list of files:

```bash
scip-treesitter-cli index --language java --out ./index.scip file1.java file2.java ...
```

Index a folder recursively:

```bash
scip-treesitter-cli index --language java --out ./index.scip --workspace <some-folder>
```

### Evaluation

```bash
scip-treesitter-cli evaluate --candidate index-tree-sitter.scip --ground-truth index.scip
```

## Development

This is a standard Rust CLI project, with a single runnable entry point - the CLI itself.

The CI uses Bazel for building and testing,
but Cargo usage is also supported for convenience.

### Running tests

```bash
cargo test
```

```bash
bazel test //docker-images/syntax-highlighter/crates/scip-treesitter-cli:all
```

We use [Insta](https://insta.rs/) for snapshot testing.
If you're changing the output produced by the CLI,
run `cargo test` and then `cargo insta review`
to accept/reject changes in snapshots.

### Build the CLI for local testing

```bash
cargo build
```

```bash
bazel build //docker-images/syntax-highlighter/crates/scip-treesitter-cli
```

### Run the locally built CLI

```bash
cargo run -- index --language java --out ./index.scip file1.java file2.java ...
```

```bash
bazel run //docker-images/syntax-highlighter/crates/scip-treesitter-cli -- index --language java --out ./index.scip file1.java file2.java ...
```
