# Sourcgraph mini
Sourcegraph can be packaged in a small format intended for self-service pilots and demonstration purposes on-premises within a customer network.

## Prerequisities
[Install docker](https://docs.docker.com/engine/installation/).

[Install docker-compose](https://docs.docker.com/compose/install/).

Sourcegraph should have provided you with a username and password for our docker container registry. To set up the docker registry credentials, use the following command:

```docker login -u <username> -p <password> docker.sourcegraph.com```

## Setup
First, try running Sourcegraph using the default settings (see below). To configure Sourcegraph mini for your own code host or local directory, read env.example. Copy env.examle to ".env" in order for `docker-compose up` to apply the settings.

## Run
To start Sourcegraph, `cd` into the mini directory, and run `docker-compose up`.

Once the process completes, try visiting the Sourcegraph application URL specified in the .env file. By default, Sourcegraph will be accessible via http://localhost:3080 . To ensure that code intelligence is working, try visiting a repository located on GitHub, such as http://localhost:3080/github.com/gorilla/mux/-/blob/mux.go , and hovering over the `Router` token (may take ~30 seconds the first time, since golang/go must be cloned in the background).

** You may see some errors about Sourcegraph connection to PostgreSQL and Redis on startup. This is caused by out of order startup of services, and shouldn't be a problem. **

## Stop
To stop Sourcegraph, `cd` into the mini directory, and run `docker-compose down`.

## Reset
Sourcegraph mini persists PostgreSQL and git data in the mini/.data directory. To reset Sourcegraph mini, stop Sourcegraph mini, delete this directory, and start Sourcegraph mini. The directory will be recreated.

## Packaging this directory
The following command will download a copy of this directory from `master`. The first time you issue this command, you will need to generate a [GitHub Personal Access Token](https://github.com/settings/tokens). After entering it the first time, it will be saved to your ~/.subversion repository.
`svn export https://github.com/sourcegraph/sourcegraph/trunk/dev/mini mini/ && zip -r mini.zip mini/ && rm -rf mini/`

## Privacy
Sourcegraph collects usage data for Sourcegraph mini to ensure customers are having success with the product. Sourcegraph mini DOES NOT collect data or metadata about a user's code, except for the what language mode is being used. This data is sent from the user's browser, and consists of the following fields:
```
{
    // FILL IN
}
```
To see what data is sent to Sourcegraph, open up your JavaScript console, and type `features.eventLogDebug.enable();`. Contact Sourcegraph if this will be a problem for your organization to trial Sourcegraph mini.
