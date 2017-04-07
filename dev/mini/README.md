# Sourcgraph mini
Sourcegraph can be packaged in a small format intended for self-service pilots and demonstration purposes on-premises within a customer network.

## Prerequisities
[Install docker](https://docs.docker.com/engine/installation/).

[Install docker-compose](https://docs.docker.com/compose/install/).

Sourcegraph should have provided you with a username and password for our docker container registry. To set up the docker registry credentials, use the following command:

```docker login -u <username> -p <password> docker.sourcegraph.com```

## Setup
First, try running Sourcegraph using the default settings (see Run section below). By default, Sourcegraph can be tested on public and private GitHub.com repositories (it uses your computer's ~/.ssh folder to try and clone repos).

Sourcegraph mini can be configured to read repositories from your local filesystem or from a remote git server.

### Local filesystem
Define the `GIT_DIRECTORY` environment argument in your `.env` file. You may need to copy `env.example` to `.env`. Next, uncomment out the `ORIGIN_MAP` argument. This argument let's Sourcegraph know that <URL>/local/repo_name can be found in the folder mounted on the docker image (the folder is mounted to /local/).

### Remote git server
Define the `ORIGIN_MAP` argument, which instructs Sourcegraph how to clone repositories beginning with a given URL prefix. For instance, if your git server is located at gitserver.companyintranet.com, then your `ORIGIN_MAP` could look like `gitserver.companyintranet.com/!git@gitserver.companyintranet.com:%`

### Set up repositories
Edit the text file `repos.txt`. Add any repositories you would like Sourcegraph to clone, using full repository paths. For local filesystem repositories, this would look like `local/repo_name`, whereas for remote git server repositories, this would look like gitserver.company.com/repo_name . Next, run `./setup.sh path/to/repos.txt`, and SQL rows will be inserted into the docker container for each repository. Refresh the Sourcegraph home page, and you should see the repositories you've added.

## Run
To start Sourcegraph, `cd` into the mini directory, and run `docker-compose up`.

Once the process completes, try visiting the Sourcegraph application URL specified in the .env file. By default, Sourcegraph will be accessible via http://localhost:3080 . To ensure that code intelligence is working, try visiting a repository located on GitHub, such as http://localhost:3080/github.com/gorilla/mux/-/blob/mux.go , and hovering over the `Router` token (may take ~30 seconds the first time, since golang/go must be cloned in the background).

** You may see some errors about Sourcegraph connection to PostgreSQL and Redis on startup. This is caused by out of order startup of services, and shouldn't be a problem. **

## Stop
To stop Sourcegraph, `cd` into the mini directory, and run `docker-compose down`.

## Reset
Sourcegraph mini persists PostgreSQL and git data in the mini/.data directory. To reset Sourcegraph mini, stop Sourcegraph mini, delete this directory, and start Sourcegraph mini. The directory will be recreated.

## Privacy
Sourcegraph collects usage data for Sourcegraph mini to ensure customers are having success with the product. Sourcegraph mini DOES NOT collect data or metadata about a user's code, except for the what language mode is being used. This data is sent from the user's browser, and consists of the following fields:
```
{
    event_action: i.e. CLICK,
    event_category: i.e. Home,
    event_label: i.e. RepoButtonClicked,
    language: i.e. go,
    platform: Web,
    repo: i.e. github.com/gorilla/mux,
    path_name: i.e. github.com/gorilla/mux/mux.go,
}
```
To see what data is sent to Sourcegraph, open up your JavaScript console, and type `features.eventLogDebug.enable();`. Contact Sourcegraph if this will be a problem for your organization to trial Sourcegraph.

## Packaging this directory (for Sourcegraph employees)
The following command will download a copy of this directory from `master`. The first time you issue this command, you will need to generate a [GitHub Personal Access Token](https://github.com/settings/tokens). After entering it the first time, it will be saved to your ~/.subversion repository.
`svn export https://github.com/sourcegraph/sourcegraph/trunk/dev/mini mini/ && sed -ie 's/TRACKING_APP_ID:/TRACKING_APP_ID: OnPremInstance/g' mini/docker-compose.yml && zip -r mini.zip mini/ && rm -rf mini/`
