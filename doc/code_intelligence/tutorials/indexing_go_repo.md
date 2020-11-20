# Index a popular Go repository

<style>
.markdown-body pre.chroma {
  font-size: 0.75em;
}

img.screenshot {
    display: block;
    margin: 1em auto;
    max-width: 600px;
    margin-bottom: 0.5em;
    border: 1px solid lightgrey;
    border-radius: 10px;
}

</style>

### Introduction
This tutorial shows you how to add LSIF data for a Go repository on your Sourcegraph instance to get precise code intelligence results.

Every time you add a repository to your Sourcegraph instance you automatically get fuzzy code intelligence results out of the box.
  <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/basic-code-intel-result.png" class="screenshot">

This can, in most cases, be good enough, but you may want to get precise results. By adding LSIF data for your repository you’ll be able to get precise results like shown below:
  <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/precise-code-intel-result.png" class="screenshot">


### Prerequisites
We recommend that use the latest version of Sourcegraph and that you have a basic understanding of how code intelligence works. See the following documents for more information:

1. ["Sourcegraph quickstart"](../../index.md)
1. ["Introduction to code intelligence"](../explanations/introduction_to_code_intelligence.md)

For the purposes of the tutorial we’ll be using the following repository: [https://github.com/moby/moby](https://github.com/moby/moby) 

Before starting this tutorial, you should first complete the following:

1. ["Add the repo to Sourcegraph"](../how-to/add_a_repository.md)
1. ["Enable LSIF in your global settings"](../how-to/enable_lsif.md)


### Clone the Go repository locally

Clone the Go repository by running: 
```
git clone https://github.com/moby/moby 
```

### Install Sourcegraph CLI

Install the Sourcegraph CLI with


*   **Linux:** 

    ```
    curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src

    chmod +x /usr/local/bin/src
    ```


*   **macOS:** 

    ```
    curl -L https://sourcegraph.com/.api/src-cli/src_darwin_amd64 -o /usr/local/bin/src

    chmod +x /usr/local/bin/src

    ```

    Or

      ```
      brew install sourcegraph/src-cli/src-cli
      ```



### Install the lsif-go indexer

Install the indexer by running:

*   **Linux:**

    ```
    curl -L  https://github.com/sourcegraph/lsif-go/releases/download/v1.2.0/src_linux_amd64 -o /usr/local/bin/lsif-go

    chmod +x /usr/local/bin/lsif-go
    ```


*   **macOS:**

    ```
    curl -L  https://github.com/sourcegraph/lsif-go/releases/download/v1.2.0/src_darwin_amd64 -o /usr/local/bin/lsif-go

    chmod +x /usr/local/bin/lsif-go

    ```



### Run the indexer in the repo

1. Navigate to the **_repo’s root folder_** and run `lsif-go -v`

2. A `dump.lsif` file should have been generated. This is the file you are going to upload to your Sourcegraph instance.
<img src="https://sourcegraphstatic.com/docs/images/code-intelligence/lsif-go-dump.png" class="screenshot center">



### Upload the LSIF data file to your Sourcegraph instance



1. Go back to your terminal and make sure you are at the repo’s root folder (where the `dump.lsif` file is). Run:
    ```
    src -endpoint=<your sourcegraph endpoint> lsif upload
    ```
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/lsif-upload2.png" class="screenshot center">


2. If you open the processing status URL you will see the current status of your upload. Wait for it to complete.
  <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/index-status.png" class="screenshot center">


### Congratulations!

You’ve indexed your first Go repo! You can now go to any Go file in that repo and get precise code intelligence results.

To learn what else you can do with Code Intelligence, see ["Code intelligence"](../index.md).