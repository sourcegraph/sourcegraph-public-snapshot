# sg tail

A small utility that connects to a running `sg start --tail` and provides a better UI to read logs.

## Usage

In your usual terminal session:

```
sg tail
```

In another terminal session:

```
cd sourcegraph
sg start --tail
```

### CLI

Flags:

- `--only-name [name]`: starts `sg tail` with a new tab focused, that only displays logs from service whose name starts with `[name]`.

### Keybindings

Press `h` or `?` when `sg tail` is running to see the inline help.
