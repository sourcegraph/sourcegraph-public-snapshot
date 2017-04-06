# Sourcgraph mini
Sourcegraph can be packaged in a small format intended for self-service pilots and demonstration purposes on-premises within a customer network.

## Prerequisities
[Install docker](https://docs.docker.com/engine/installation/).

[Install docker-compose](https://docs.docker.com/compose/install/).

Set up the docker registry credentials for the Sourcegraph docker registry. They will be provided by Sourcegraph.

```docker login -u <username> -p <password> docker.sourcegraph.com```

## Setup
Copy the `.env.example` file to `.env`, and modify settings for your environment. More details available in the configuration file.

## Run
To start Sourcegraph, `cd` into the mini directory, and run `docker-compose up`.
