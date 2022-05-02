# How to set up the Sourcegraph local development environment

This document guides you through the installation and basic configuration of components required to set up a working local environment for Sourcegraph on a Mac, so that you can run them on your own computer, and make changes to the code. Sourcegraph uses [sg cli](https://docs.sourcegraph.com/dev/background-information/sg) which is the Sourcegraph developer tool, to manage the local development environment.

## Pre-requisites

In order to use [SG CLI](https://github.com/sourcegraph/src-cli), you will need to install various packages account requirements including:

* Docker
* MacOS 
* Personal Github account with [SSH key set up](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent) 
* Approximately 1gb disk space for cloned Sourcegraph repo

SG CLI will automatically attempt to install and set up the following if they are not installed:

* Homebrew
* Git
* Gnu-sed
* pcre
* sqlite
* jq
* bash
* rosetta
* PostgresSQL database
* Redis database
* Proxy for local development

## Install the sg CLI

Open your Mac terminal and install sg by running the following:

`curl --proto '=https' --tlsv1.2 -sSLf https://install.sg.dev | sh`

Ensure sg successfully installed by running:

`sg logo`

## Set up development environment

Run: 

`sg setup`

This will automatically check for required dependencies and guide you through steps to install them. Make sure all the checks have cleared before moving forward. 

**Note: if you are not a sourcegraph employee, you will not need access to the private repo.**

## Start the server

The server will continuously compile code. Start and run Sourcegraph with one of two ways:

### Run Enterprise version for Sourcegraph employees: 

`sg start`

### Run open source version: 

`sg start oss`

Navigate to your local Sourcegraph instance at https://sourcegraph.test:3443 

## Next steps for new developers

To make code changes, make a new branch and navigate to it. This is where you will make your code changes. Example:

`git checkout -b your_new_branch_name`

Be sure to review the steps in [Contributing to Sourcegraph](https://github.com/northyg/sourcegraph/blob/main/CONTRIBUTING.md)

