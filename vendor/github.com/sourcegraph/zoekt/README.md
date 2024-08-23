
    "Zoekt, en gij zult spinazie eten" - Jan Eertink

    ("seek, and ye shall eat spinach" - My primary school teacher)

This is a fast text search engine, intended for use with source
code. (Pronunciation: roughly as you would pronounce "zooked" in English)

**Note:** This is a [Sourcegraph](https://github.com/sourcegraph/zoekt) fork
of [github.com/google/zoekt](https://github.com/google/zoekt). It is now the
main maintained source of Zoekt.

# INSTRUCTIONS

## Downloading

    go get github.com/sourcegraph/zoekt/

## Indexing

### Directory

    go install github.com/sourcegraph/zoekt/cmd/zoekt-index
    $GOPATH/bin/zoekt-index .

### Git repository

    go install github.com/sourcegraph/zoekt/cmd/zoekt-git-index
    $GOPATH/bin/zoekt-git-index -branches master,stable-1.4 -prefix origin/ .

### Repo repositories

    go install github.com/sourcegraph/zoekt/cmd/zoekt-{repo-index,mirror-gitiles}
    zoekt-mirror-gitiles -dest ~/repos/ https://gfiber.googlesource.com
    zoekt-repo-index \
        -name gfiber \
        -base_url https://gfiber.googlesource.com/ \
        -manifest_repo ~/repos/gfiber.googlesource.com/manifests.git \
        -repo_cache ~/repos \
        -manifest_rev_prefix=refs/heads/ --rev_prefix= \
        master:default_unrestricted.xml

## Searching

### Web interface

    go install github.com/sourcegraph/zoekt/cmd/zoekt-webserver
    $GOPATH/bin/zoekt-webserver -listen :6070

### JSON API

You can retrieve search results as JSON by sending a GET request to zoekt-webserver.

    curl --get \
        --url "http://localhost:6070/search" \
        --data-urlencode "q=ngram f:READ" \
        --data-urlencode "num=50" \
        --data-urlencode "format=json"

The response data is a JSON object. You can refer to [web.ApiSearchResult](https://sourcegraph.com/github.com/sourcegraph/zoekt@6b1df4f8a3d7b34f13ba0cafd8e1a9b3fc728cf0/-/blob/web/api.go?L23:6&subtree=true) to learn about the structure of the object.

### CLI

    go install github.com/sourcegraph/zoekt/cmd/zoekt
    $GOPATH/bin/zoekt 'ngram f:READ'

## Installation
A more organized installation on a Linux server should use a systemd unit file,
eg.

    [Unit]
    Description=zoekt webserver

    [Service]
    ExecStart=/zoekt/bin/zoekt-webserver -index /zoekt/index -listen :443  --ssl_cert /zoekt/etc/cert.pem   --ssl_key /zoekt/etc/key.pem
    Restart=always

    [Install]
    WantedBy=default.target


# SEARCH SERVICE

Zoekt comes with a small service management program:

    go install github.com/sourcegraph/zoekt/cmd/zoekt-indexserver

    cat << EOF > config.json
    [{"GithubUser": "username"},
     {"GithubOrg": "org"},
     {"GitilesURL": "https://gerrit.googlesource.com", "Name": "zoekt" }
    ]
    EOF

    $GOPATH/bin/zoekt-indexserver -mirror_config config.json

This will mirror all repos under 'github.com/username', 'github.com/org', as
well as the 'zoekt' repository. It will index the repositories.

It takes care of fetching and indexing new data and cleaning up logfiles.

The webserver can be started from a standard service management framework, such
as systemd.


# SYMBOL SEARCH

It is recommended to install [Universal
ctags](https://github.com/universal-ctags/ctags) to improve
ranking. See [here](doc/ctags.md) for more information.


# ACKNOWLEDGEMENTS

Thanks to Han-Wen Nienhuys for creating Zoekt. Thanks to Alexander Neubeck for
coming up with this idea, and helping Han-Wen Nienhuys flesh it out.


# FORK DETAILS

Originally this fork contained some changes that do not make sense to upstream
and or have not yet been upstreamed. However, this is now the defacto source
for Zoekt. This section will remain for historical reasons and contains
outdated information. It can be removed once the dust settles on moving from
google/zoekt to sourcegraph/zoekt. Differences:

- [zoekt-sourcegraph-indexserver](cmd/zoekt-sourcegraph-indexserver/main.go)
  is a Sourcegraph specific command which indexes all enabled repositories on
  Sourcegraph, as well as keeping the indexes up to date.
- We have exposed the API via
  [keegancsmith/rpc](https://github.com/keegancsmith/rpc) (a fork of `net/rpc`
  which supports cancellation).
- Query primitive `BranchesRepos` to efficiently specify a set of repositories to
  search.
- Allow empty shard directories on startup. Needed when starting a fresh
  instance which hasn't indexed anything yet.
- We can return symbol/ctag data in results. Additionally we can run symbol regex queries.
- We search shards in order of repo name and ignore shard ranking.
- Other minor changes.

Assuming you have the gerrit upstream configured, a useful way to see what we
changed is:

``` shellsession
$ git diff gerrit/master -- ':(exclude)vendor/' ':(exclude)Gopkg*'
```

# DISCLAIMER

This is not an official Google product
