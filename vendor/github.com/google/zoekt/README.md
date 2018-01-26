
    "Zoekt, en gij zult spinazie eten" - Jan Eertink

    ("seek, and ye shall eat spinach" - My primary school teacher)

This is a fast text search engine, intended for use with source
code. (Pronunciation: roughly as you would pronounce "zooked" in English)

INSTRUCTIONS
============

Downloading:

    go get github.com/google/zoekt/

Indexing:

    go install github.com/google/zoekt/cmd/zoekt-index
    $GOPATH/bin/zoekt-index .

Searching

    go install github.com/google/zoekt/cmd/zoekt
    $GOPATH/bin/zoekt 'ngram f:READ'

Indexing git repositories:

    go install github.com/google/zoekt/cmd/zoekt-git-index
    $GOPATH/bin/zoekt-git-index -branches master,stable-1.4 -prefix origin/ .

Indexing repo repositories:

    go install github.com/google/zoekt/cmd/zoekt-{repo-index,mirror-gitiles}
    zoekt-mirror-gitiles -dest ~/repos/ https://gfiber.googlesource.com
    zoekt-repo-index \
       -name gfiber \
       -base_url https://gfiber.googlesource.com/ \
       -manifest_repo ~/repos/gfiber.googlesource.com/manifests.git \
       -repo_cache ~/repos \
       -manifest_rev_prefix=refs/heads/ --rev_prefix= \
       master:default_unrestricted.xml

Starting the web interface

    go install github.com/google/zoekt/cmd/zoekt-webserver
    $GOPATH/bin/zoekt-webserver -listen :6070

A more organized installation on a Linux server should use a systemd unit file,
eg.

    [Unit]
    Description=zoekt webserver

    [Service]
    ExecStart=/zoekt/bin/zoekt-webserver -index /zoekt/index -listen :443  --ssl_cert /zoekt/etc/cert.pem   --ssl_key /zoekt/etc/key.pem
    Restart=always

    [Install]
    WantedBy=default.target


SEARCH SERVICE
==============

Zoekt comes with a small service management program:

    go install github.com/google/zoekt/cmd/zoekt-indexserver

    cat << EOF > config.json
    [{"GithubUser": "username"},
     {"GitilesURL": "https://gerrit.googlesource.com", Name: "zoekt" }
    ]
    EOF

    $GOPATH/bin/zoekt-server -mirror_config config.json

This will mirror all repos under 'github.com/username' as well as the
'zoekt' repository. It will index the repositories.

It takes care of fetching and indexing new data and cleaning up logfiles.

The webserver can be started from a standard service management framework, such
as systemd.


SYMBOL SEARCH
=============

It is recommended to install [Universal
ctags](https://github.com/universal-ctags/ctags) to improve ranking,
in particular, version `db3d9a6` or newer, which supports seccomp
sandboxing.


ACKNOWLEDGEMENTS
================

Thanks to Alexander Neubeck for coming up with this idea, and helping me flesh
it out.


DISCLAIMER
==========

This is not an official Google product
