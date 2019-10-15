# How to setup a separate website maintained by Sourcegraph

    💡 Use this template to describe the steps engineers should follow to setup a separate website maintained by Sourcegraph. 

# 1. When do we need a separate website?

For projects like [langserver.org](http://langserver.org) or [lsif.dev](http://lsif.dev), we prefer to create separate websites from 
[Sourcegraph.com](http://sourcegraph.com). In these two examples, we provide a list of LSPs or LSIF indexers offered by us or other parties 
that can be used for Sourcegraph or other static analysis applications. Since the usage isn't strictly tide to Sourcegraph and we are not 
the sole providers, we publish this information on an independent website.

# 2. Setup domain and create content

### Secure the domain

    💡 First step is to get a domain at Namecheap and add the site to our Cloudflare account (Beyang or Quinn would do this). Make sure `www` is redirecting to `https`.

### Create the website content

- Once the domain is secured, we need to setup the Github page, where the static page will be hosted as: `[projectname].github.io`.
- We clone the repo locally, create a branch from master (make sure master is up-to-date!) and add the following files:
    - CNAME
    Containing:

        [project url]

    - README.md
    - index.html
    - style-addition.css
- Once all content is created, we `push` our branch, create a **PR** and ask for reviews and merge to master.

Examples: [Langserver](http://github.com/langserver), [LSIF](http://github.com/lsif)

# 3. Update the DNS to point to the server

We use terraform to manage our DNS records. To update DNS, follow this guide, but push to `branch` not `master` and get it reviewed before merging. 

If you don't have terraform installed and install via ```brew install terraform```:

    💡 Ensure that your local `terrraform` version is the same as the `CI` version to **avoid file locks**.

[DNS guide](../../../../../../infrastructure/blob/master/dns/README.md)

    💡 Make sure to pull the **latest version** of the infrastructure `master` before creating your branch.

    💡 Check that all of the domain's `A records` are also referenced in the terraform file.

Inside your branch:

    $ terraform apply
    $ git commit -m 'dns: Updated terraform.tfstate' terraform.tfstate
    $ git push origin branch

Get the branch reviewed and merge to `master`.
