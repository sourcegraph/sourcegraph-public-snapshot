# Swift: code intelligence configuration

This page describes additional configuration that may be needed for Swift code intelligence on certain codebases. To enable Swift code intelligence, see the [installation documentation](install/index.md).

---

**Note:** The [Swift language server](https://github.com/RLovelett/langserver-swift) currently only supports [Swift Package Manager](https://swift.org/package-manager/) projects (that is, ones with a `Package.swift` file).

Open the App Store and install Xcode.

Accept the license:

```
sudo xcodebuild -license
```

Make sure Swift 4.1+ is installed:

```
$ swift --version
Apple Swift version 4.1.2 (swift-4.1.2-RELEASE)
Target: x86_64-apple-darwin16.7.0
```

Download and install the Swift language server:

```
git clone https://github.com/RLovelett/langserver-swift
cd langserver-swift
git checkout 2ede196ab35381c0661eea2620f7daa8ab6a31d8
sudo mkdir -p /usr/local/bin
sudo make install
```

Install lsp-adapter (for connecting the Swift language server to Sourcegraph):

```
curl -L https://github.com/sourcegraph/lsp-adapter/releases/download/v0.0.7/lsp-adapter_darwin_amd64 -o lsp-adapter
chmod 755 lsp-adapter
```

Run the Swift language server with `./lsp-adapter`:

```
./lsp-adapter -proxyAddress 0.0.0.0:8080 langserver-swift
```

Add an entry to the `langservers` field in the Sourcegraph site configuration, replacing `example.com` with the IP address or hostname of the machine that's running the Swift language server:

```
    {
      "address": "tcp://example.com:8080",
      "language": "swift"
    },
```

Save the changes.

Once you open a Swift file (e.g. http://sourcegraph.example.com/github.com/RLovelett/langserver-swift/-/blob/Sources/LanguageServerProtocol/Types/Server.swift), you should see `lsp-adapter` logging requests. You should also see hover tooltips on variables, functions, classes, etc.
