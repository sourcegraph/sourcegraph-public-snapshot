# Index a popular public Go repository

<style>
.markdown-body pre.chroma {
  font-size: 0.75em;
}

img.screenshot {
    max-width: 600px;
    margin: 1em;
    margin-bottom: 0.5em;
    border: 1px solid lightgrey;
    border-radius: 10px;
}

img.center {
  display: block;
  margin: auto
}
</style>

### Introduction
This tutorial shows you how to add LSIF data for a Go repository on your private Sourcegraph instance to get precise code intelligence results.

Every time you add a repository to your Sourcegraph instance you automatically get fuzzy code intelligence results out of the box.
  <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/basic-code-intel-result.png" class="screenshot center">

This can, in most cases, be good enough, but you may want to get precise results. By adding LSIF data for your repository you’ll be able to get precise results like shown below:
  <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/precise-code-intel-result.png" class="screenshot center">


### Prerequisites
We recommend that use the latest version of Sourcegraph and that you have a basic understanding of how code intelligence works. See the following documents for more information:

1. ["Sourcegraph quickstart"](../../index.md)
1. ["Introduction to code intelligence"](../explanations/introduction_to_code_intelligence.md)

Before starting this tutorial, you should first complete the following:

1. ["Add a Go repo to Sourcegraph"](../how-to/add_a_go_repository.md)
1. ["Enable LSIF in your global settings"](../how-to/enable_lsif.md)


For the purposes of the tutorial we’ll be using the following repository: [https://github.com/moby/moby](https://github.com/moby/moby) 

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

Install the indexer by running

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
<img src="https://sourcegraphstatic.com/docs/images/code-intelligence/run-lsif-go-indexer.png" class="screenshot center">


2. A **_dump.lsif_** file should have been generated, this is what we are going to upload to our Sourcegraph instance.
<img src="https://sourcegraphstatic.com/docs/images/code-intelligence/lsif-go-dump.png" class="screenshot center">



### Upload the LSIF data file to your Sourcegraph instance



1. Go back to your terminal and make sure you are at the repo’s root folder (where the dump.lsif file is). Run the following command:
    ```
    src -endpoint=<your sourcegraph endpoint> lsif upload
    ```
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/lsif-upload2.png" class="screenshot center">


2. If you open the processing status URL you should see the current status of your upload. Wait for it to complete.
  <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/index-status.png" class="screenshot center">


### Congratulations!

You’ve indexed your first Go repo! You can now go to any Go file in that repo and get precise code intelligence results.

To learn what else you can do with Code Intelligence, see ["Code intelligence"](../index.md).