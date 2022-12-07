# codebot

## Dev

Proxy:

```
gow run proxy.go --remote "http://129.146.104.152:5000"
```

The proxy in the long-term is probably unnecessary. Right now, it
makes some modifications to talk to https://github.com/moyix/fauxpilot
to make things work.

Frontend:

```
code ./vscode-codegen
```

Then run the extension. Pause for autocomplete or hit `alt-\` to
generate a few examples.
