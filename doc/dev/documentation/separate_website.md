# How to setup a separate website maintained by Sourcegraph

  > 💡 This documentation describes the steps engineers should follow to set up a separate website maintained by Sourcegraph. 

## 1. When do we need a separate website?

For projects like [langserver.org](http://langserver.org) or [lsif.dev](http://lsif.dev), we prefer to create separate websites from 
[Sourcegraph.com](http://sourcegraph.com). In these two examples, we provide a list of LSPs or LSIF indexers offered by us or other parties 
that can be used for Sourcegraph or other static analysis applications. Since the usage isn't strictly tied to Sourcegraph and we are not 
the sole providers, we publish this information on an independent website.

## 2. Set up domain and create content

### Secure the domain

First step is to request [@beyang](https://github.com/beyang) get a domain and add the site to Sourcegraph's Cloudflare account. 
    
> 💡 Make sure `www` redirects to `https`.

### Create the website content

Once the domain is secured, we need to setup the GitHub page, where the static page will be hosted as: `[projectname].github.io`.

We clone the repository locally, create a branch from master (make sure master is up-to-date!) and add the following files:

    - CNAME
    Containing:

        [project url]

    - README.md
    - index.html
    - style-addition.css

When all the content is created, we `push` our branch, create a **PR** and ask for reviews and merge to master.

Examples: [Langserver](https://github.com/langserver/langserver.github.io), [LSIF](https://github.com/lsif/lsif.github.io)

## 3. Update the DNS to point to the server

We use terraform to manage our DNS records. 

> 💡 If you don't have terraform installed and install via ```brew install terraform```. Ensure that your local `terraform` version is the same as the `CI` version to **avoid file locks**.  

To update DNS, follow the [DNS guide](https://github.com/sourcegraph/infrastructure/blob/master/dns/README.md), but push to a non-`master` branch and get it reviewed before merging.

> 💡 Important:
> - Make sure to pull the **latest version** of the infrastructure `master` before creating your branch.
> - Check that all of the domain's `A records` are also referenced in the terraform file.
>

From the `infrastructure` repository, with your branch checked out:

    $ terraform apply
    $ git commit -m 'dns: Updated terraform.tfstate' terraform.tfstate
    $ git push origin branch

Get the branch reviewed and merge to `master`.
