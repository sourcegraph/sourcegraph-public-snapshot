# Sourcgraph mini
Sourcegraph can be packaged in a small format intended for self-service pilots and demonstration purposes on-premises within a customer network.

## Prerequisities
[Install docker](https://docs.docker.com/engine/installation/).

[Install docker-compose](https://docs.docker.com/compose/install/).

Sourcegraph should have provided you with a username and password for our docker container registry. To set up the docker registry credentials, use the following command:

```docker login -u <username> -p <password> docker.sourcegraph.com```

Try skipping to the "Run" section, and verifying that everything is working, before configuring further.

## Setup
First, try running Sourcegraph using the default settings (see Run section below). By default, Sourcegraph can be tested on public and private GitHub.com repositories (it uses your computer's ~/.ssh folder to try and clone repos).

Sourcegraph mini can be configured to read repositories from your local filesystem or from a remote git server.

### Local filesystem
Define the `GIT_DIRECTORY` environment argument in your `.env` file. You may need to copy `env.example` to `.env`. Next, uncomment out the `ORIGIN_MAP` argument. This argument let's Sourcegraph know that <SOURCEGRAPH_URL>/local/repo_name can be found in the folder mounted on the git-server docker image (the folder is mounted to /local/). Skip to "Set up repositories".

### Remote git server
Define the `ORIGIN_MAP` argument, which instructs Sourcegraph how to clone repositories beginning with a given URL prefix. For instance, if your git server is located at gitserver.companyintranet.com, then your `ORIGIN_MAP` could look like `gitserver.companyintranet.com/!git@gitserver.companyintranet.com:%`. In general, this is harder to do than local filesystem, so recommended to start with local filesystem. After this step, skip to "Set up repositories".

### Set up repositories
Edit the text file `repos.txt`. Add any repositories you would like Sourcegraph to clone, using full repository paths. For local filesystem repositories, this would look like `local/any_dirs/repo_name`, whereas for remote git server repositories, this would look like gitserver.company.com/any_dirs/repo_name . Make sure that Sourcegraph is running (see the Run section), and next run `./setup.sh path/to/repos.txt`, and SQL rows will be inserted into the docker container for each repository. Refresh the Sourcegraph home page, and you should see the repositories you've added.

## Run
To start Sourcegraph, `cd` into the mini directory, and run `docker-compose up`.

Once the process completes, try visiting the Sourcegraph application URL specified in the .env file. By default, Sourcegraph will be accessible via http://localhost:3080 . To ensure that code intelligence is working, try visiting a repository located on GitHub, such as http://localhost:3080/github.com/gorilla/mux/-/blob/mux.go , and hovering over the `Router` token (may take ~30 seconds the first time, since golang/go must be cloned in the background).

** You may see some errors about Sourcegraph connection to PostgreSQL and Redis on startup. This is caused by out of order startup of services, and shouldn't be a problem. **

## Stop
To stop Sourcegraph, `cd` into the mini directory, and run `docker-compose down`.

## Update
To update to the latest Sourcegraph images, run `docker-compose pull`.

## Reset
Sourcegraph mini persists PostgreSQL and git data in the mini/.data directory. To reset Sourcegraph mini, stop Sourcegraph mini, delete this directory, and start Sourcegraph mini. The directory will be recreated.

## Privacy
Sourcegraph collects usage data via JavaScript loaded on the web UI for Sourcegraph mini to ensure customers are having success with the product. The JavaScript usage tracker DOES NOT collect user code, _just_ repository names and file names. This data is sent from the user's browser, and consists of the following fields:
```
{
    event_action: i.e. CLICK,
    event_category: i.e. Home,
    event_label: i.e. RepoButtonClicked,
    language: i.e. go,
    platform: Web,
    repo: i.e. github.com/gorilla/mux,
    page_title: i.e. HomePage
    path_name: i.e. github.com/gorilla/mux/mux.go,
}
```
To see what data is sent to Sourcegraph, open up your JavaScript console, and view network traffic to the `production` endpoint.

## Packaging this directory (for Sourcegraph employees)
The following command will download a copy of this directory from `master`. The first time you issue this command, you will need to generate a [GitHub Personal Access Token](https://github.com/settings/tokens). After entering it the first time, it will be saved to your ~/.subversion repository.
`svn export https://github.com/sourcegraph/sourcegraph/trunk/dev/mini mini/ && sed -ie 's/TRACKING_APP_ID:/TRACKING_APP_ID: OnPremInstance/g' mini/docker-compose.yml && zip -r mini.zip mini/ && rm -rf mini/`

## FAQs

### Unable to clone repositories from GitHub.com

Sourcegraph requires access to GitHub.com via the git-server container for open-source dependency resolution. By default, the ~/.ssh directory is mounted to the git-server ~/.ssh directory. If your SSH keys are not located in .ssh, or are not set up to be used with your GitHub.com account, then cloning repositories from GitHub may fail.

### Running setup.sh fails with message "No container found for postgres_1"

Make sure that Sourcegraph is up by following the instructions in "Run" before running `setup.sh`.

## Troubleshooting
Please contact support@sourcegraph.com with questions about Sourcegraph mini. Include the output of `docker-compose ps`, a copy of your `.env` file, and any logs that might be relevant.
