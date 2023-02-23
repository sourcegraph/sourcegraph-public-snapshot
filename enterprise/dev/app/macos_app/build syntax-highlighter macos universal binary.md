# notes for cross-compiling Rust-based tools (syntect_server and others)

install `rustup` to enable adding alternate targets. Might be able to do this another way, but I didn't want to mess with `asdf`.

```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

It complained because I already have rust installed via `asdf`. I just chose to install anyway.

set PATH to use the newly-installed `cargo` and `rustup` (instead of the asdf-managed ones)

```
export PATH="$HOME/.cargo/bin:$PATH"
```

That should have been part of `~/.profile` etc..., but it doesn't seem to work correctly.

`rustup target list` shows the available targets. There are many; maybe filter using `| grep darwin` for example.

`rustup target add x86_64-apple-darwin` to add another darwin target

compile using

```
cargo build --release --target=aarch64-apple-darwin
cargo build --release --target=x86_64-apple-darwin
```

Can see the compiled binaries in `target/aarch64-apple-darwin/release` and `target/x86_64-apple-darwin/release`.

Use `file ...` to see that they are the expected platform type.

Combine into a universal binary (MacOS) using

```
mkdir -p target/universal/release
lipo target/aarch64-apple-darwin/release/syntect_server target/x86_64-apple-darwin/release/syntect_server -create -output target/universal/release/syntect_server
```

confirm platforms using `file target/universal/release/syntect_server` and check symbolic links (should be only system ones) using `otool -L target/universal/release/syntect_server`

alternately, [this docker-based cross compiler](https://github.com/joseluisq/rust-linux-darwin-builder) looks promising
