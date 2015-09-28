# Sourcegraph: the intelligent and hackable code platform

Sourcegraph makes your team more collaborative and efficient by
letting them:

* search, browse, and cross-reference code (like an IDE)
* view live usage examples for any function, type, etc.
* perform code reviews
* carry on persistent discussions on any piece of code

Status: **limited release** ([Go](https://golang.org) support only)

Your git repositories can live on Sourcegraph, or you can use it to
search and browse existing repositories.

Start using Sourcegraph for your team's code (see **Quickstart**
below), or try it out at [Sourcegraph.com](https://sourcegraph.com)
for public, open-source code.


## Installation

### Quickstart

The simplest way to get started with Sourcegraph is to
[download a pre-built static binary for your platform](https://sourcegraph.com/b86d5501a450ca38be78b112d88cb46d9bf27583/try-it#otherDownloadOptions).

Once you've downloaded and unzipped the binary:

1. `src serve`

2. Follow prompts at http://localhost:3000 to add your repositories, or see below to check out sample repositories.

**Sample repositories:** Want to try it out on some existing
repositories before you add your own? Run these commands, then go to
http://localhost:3000 to view them. When you visit a repository for
the first time, the build process will begin; wait for it to finish to
be able to search and browse code.

```
# Go
git clone https://github.com/gorilla/mux.git
ln -s mux ~/.sourcegraph/repos/github.com/gorilla/mux
```

### Team server

When you're ready to make Sourcegraph your team's git repository host,
create an Amazon EC2 instance or Digital Ocean Droplet running Ubuntu 14.04.

**Log in to your instance and follow these steps:**

1. `wget -O - https://sourcegraph.com/.download/install.sh | bash`

2. `sudo add-apt-repository -y ppa:evarlast/golang1.5 && sudo aptitude update && sudo aptitude install -y golang-go`

3. `sudo -iu sourcegraph src toolchain install go`

4. Open `/etc/sourcegraph/config.ini` and follow the instructions inside to set your server's hostname.

5. `sudo restart src`

6. Go to `http://<hostname-or-ip>:3000` and follow instructions to register your instance.

7. `src --endpoint=http://<hostname-or-ip>:3000 repo create <name-of-repo>`

(In step 7, you'll first be prompted to run `src login` -- follow those instructions then re-run that command.)

**On your laptop (from your repo directory):**

1. `git remote add origin http://<hostname-or-ip>:3000/<name-of-repo>`

2. `git push -u origin master`

(In step 2, provide your sourcegraph.com login credentials when prompted by git.)

### Grant access to teammates

When you first spin up a local Sourcegraph instance and log in with your
sourcegraph.com credentials, you will be granted admin access to that instance. After
that, only an admin user can grant access to others. Use the `src access ...` commands
to control access to your Sourcegraph instance:

1. `src access list`
	List all usernames that have access to your Sourcegraph, along with their permission
	levels (read/write/admin).

2. `src access grant [--admin] user1 user2 ...`
	Grant read+write access to users with sourcegraph.com usernames user1, user2, etc.
	To grant admin access to these users, set the `--admin` flag.

3. `src access revoke user1 user2 ...`
	Revoke all access from users user1, user2, etc.

NOTE that your CLI must be authenticated against your Sourcegraph instance with
your (admin) credentials to run these commands. Follow the instructions at:

	`http://APP_URL/~MYUSERNAME/.settings/auth`

### Install from source

Alternatively, you can install Sourcegraph from source. You'll need
[Go 1.4+](https://golang.org), a `$GOPATH` set up, and `go install`
needs to install into your `$PATH`.

```bash
mkdir -p $GOPATH/src/sourcegraph.com/sourcegraph
cd $GOPATH/src/sourcegraph.com/sourcegraph
git clone https://USER:PASSWORD@src.sourcegraph.com/sourcegraph

cd $GOPATH/src/src.sourcegraph.com/sourcegraph
make install

# Check that "src" was installed:
src --help
```

Next, install [srclib](https://srclib.org) toolchains to support
analyzing languages you're interested in. For example:

```
src toolchain install go
```

**Note:** Currently the only supported language is Go;
support for other languages is available in alpha.

Then run `src serve` and open http://localhost:3000 to access
Sourcegraph.

See README.dev.md for more information for Sourcegraph developers.

## Under the hood

Sourcegraph is built on several components:

* [srclib](https://srclib.org), a multi-language, hackable source code
  analysis toolchain
* The [Go](http://golang.org) programming language
* [gRPC](http://grpc.io), an HTTP2-based RPC protocol that uses
  Protocol Buffer service definitions
* [React](https://facebook.github.io/react/), a JavaScript library for
  building UIs.
* [Sourcegraph.com](https://sourcegraph.com), a public instance of
  Sourcegraph that provides information about open-source projects to
  your local Sourcegraph.


## Contributing to Sourcegraph

Want to make Sourcegraph better? Great! Check out
[CONTRIBUTING.md](CONTRIBUTING.md). We welcome all types of
contributions--code, documentation, assets, community support, and
user feedback.


## Security

Security is very important to us. If you discover a security-related
issue, please responsibly disclose it by emailing
security@sourcegraph.com and not by creating an issue.


## License

(TODO)
