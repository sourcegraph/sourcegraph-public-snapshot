# Load Templates from Binary Data

This example shows how to load templates from binary data instead of template files.

You can run this example web application by executing the following command.

```sh
$ go run main.go asset.go
```

`asset.go` is created by executing `make_asset.sh`. This shell script executes [go-bindata](https://github.com/jteeuwen/go-bindata) and generates binary data from the Ace template.

**You can compile your web application into one binary file by using this function.**
