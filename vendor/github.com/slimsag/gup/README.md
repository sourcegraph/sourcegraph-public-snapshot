# gup: an opinionated Go binary updater [![GoDoc](https://godoc.org/github.com/slimsag/gup?status.svg)](https://godoc.org/github.com/slimsag/gup)

Gup is a library for self-updating Go binaries. That is, fetching an update to your Go binary from the internet and applying the update.

Unlike other libraries like [go-updater](https://github.com/inconshreveable/go-update), which gup uses internally, gup makes the following assumptions:

1. Updates should be downloaded from an HTTPS web server (e.g. a Google Cloud Storage bucket).
2. Security matters to you and you always want to verify the SHA256 and ECDSA binary signatures of updates.
3. You don't particularly care about the directory structure of your web server, i.e. you can give Gup control of an entire directory.

These assumptions make Gup much easier to use than the alternatives, which typically have security etc. as an add-on or paid feature.

## Generating Keys

First, you'll want to generate a public and private ECDSA P256 keypair. To do this, just run `gup genkey`:

```bash
$ gup genkey
# wrote private_key.pem
# wrote public_key.pem
Note: You may copy+paste the following public key into your Go program:

gup.Config.PublicKey = "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEm5DKV8SqS7HjIVtsFjGc93TYr/LA\naE9p72sF6qc1MjYMUoukScQoY0MysgEdekf/cmiWpKYwLc2rn8BnBRdz+w==\n-----END PUBLIC KEY-----\n"
```

As it suggests, you'll copy+paste the Go code it printed out into your Go program below.

## Example code

A basic program looks like this:

```Go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/slimsag/gup"
)

func main() {
	// Configure and start Gup.
	gup.Config.PublicKey = "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEm5DKV8SqS7HjIVtsFjGc93TYr/LA\naE9p72sF6qc1MjYMUoukScQoY0MysgEdekf/cmiWpKYwLc2rn8BnBRdz+w==\n-----END PUBLIC KEY-----\n"
	gup.Config.UpdateURL = "https://storage.googleapis.com/my-bucket/updates/$GUP"
	gup.Config.CheckInterval = 5 * time.Second // For production, you'll want to use something larger
	gup.Start()

	// Wait for updates to become available.
	<-gup.UpdateAvailable
	fmt.Println("an update is available!")

	// Perform the update.
	_, err := gup.Update()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("update successful, please relaunch the program")
}
```

## Creating index.json

The very first time you get started out with Gup, you'll want to create your "first patch" and the `index.json` file. This index file keeps track of all the versions of your program that are available, and it will also be uploaded to your webserver. To do this run build your binary and run `gup bundle myprogram`:

```bash
$ go build -o gup/bin/myprogram-latest ./example/
$ gup bundle gup/bin/myprogram-latest
wrote replacement patch bundle: gup/main-darwin-amd64-0.tgz
wrote index: gup/index.json
...
```

Now you'll want to upload the new files it wrote to the `gup/` folder to your webserver (`gup.Config.UpdateURL`). For example with Google Cloud Storage:

```bash
$ gsutil -m -h Cache-Control:"Cache-Control:private, max-age=0, no-transform" cp -a public-read -r gup/* gs://my-bucket/updates
```

## Creating a patch

To release a new version of your application, you'll create a patch using `gup patch`. This command will produce another `gup/*.tgz` patch bundle and update the `gup/index.json` to reflect the version being available. First, make a change to your application like adding `fmt.Println("MyApp started!")` at the top of `main`, then run:

```bash
$ mv gup/bin/myprogram-latest gup/bin/myprogram-prev
$ go build -o gup/bin/myprogram-latest ./example/
$ gup bundle ./gup/bin/myprogram-prev ./gup/bin/myprogram-latest 
wrote diff patch bundle: gup/main-darwin-amd64-1.tgz
updated index: gup/index.json
```

Now, in one terminal start the old program:

```bash
$ cp ./gup/bin/myprogram-prev myprogram-old
$ ./myprogram-old 
```

And in a second terminal, release the uploading these files to your web server:

```bash
$ gsutil -m -h Cache-Control:"Cache-Control:private, max-age=0, no-transform" cp -a public-read -r gup/* gs://my-bucket/updates
```

And you'll see:

```
$ ./myprogram-old 
an update is available!
2017/02/09 17:55:50 applying update main-darwin-amd64-1.tgz
update successful, please relaunch the program
```

Running the program once more, you'll see:

```
$ ./myprogram-old 
MyApp started!
```

## Notes

- You should backup _at least_ `private_key.pem`, `gup/index.json`, and your 'latest' binary (e.g. `gup/bin/myprogram-latest`).
  - Failure to do this means you will no longer be able to publish updates, as Gup needs all three of these to successfully publish a new update. 
- In a real application, you should use a higher `gup.Config.CheckInterval`, e.g. `5 * time.Minute`.
- While the above shows commands for Google Cloud Storage, any HTTPS file host will do.
- Gup patch bundle files (e.g. `main-darwin-amd64-0.tgz`) are designed to be easily introspected, e.g. try `tar -xzf -C tmp/ main-darwin-amd64-0.tgz`.
- The above didn't cover _replacement_ patches, which are different than the default binary patch mode. See `gup bundle -h` for details.
- When uploading files, you do not want to overwrite the entire directory (i.e. you want to keep old `.tgz` files that are in the directory). Also be aware of how your file host caches `index.json` as that will affect your application's ability to discover updates quickly.

