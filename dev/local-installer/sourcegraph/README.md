# Sourcegraph local installer

This README walks you through how to run an instance of Sourcegraph on your local machine.

## System requirements

* Install [Docker](https://www.docker.com/community-edition) version 1.12.0 or later (run `docker version` to check).
* Install [Docker Compose](https://docs.docker.com/compose/install/). (On macOS and Windows, this is automatically installed with Docker.)

## Credentials

Sourcegraph is currently available to a restricted set of users. If you are part of this set, you should have received a username and password to the Sourcegraph Docker container registry. Before starting, run the following command to authorize your Docker client to fetch from this registry:

```
docker login -u <username> -p <password> docker.sourcegraph.com
```

## Install

1. `cd` into the directory containing this README and run `docker-compose pull && docker-compose up`.
1. Visit http://localhost:3080/github.com/gorilla/mux.
1. **That's it**—you can start exploring code on your local instance of Sourcegraph.

For example, open [`mux.go`](http://localhost:3080/github.com/gorilla/mux/-/blob/mux.go) and try clicking on some function names or search for symbols using the '/#' hotkey combination.<br>
*Note: you may have to wait about 60 seconds the very first time for jump-to-definition and tooltips to work.*

### Add your private repositories

Before adding your private repositories, please note the "Privacy" section of this README.

1. Stop the Sourcegraph instance with `Ctrl-C` in the terminal running `docker-compose up`.
1. Run `cp env.example .env`

   (The ".env" file sets default values for environment variables when running Docker Compose.)

1. Uncomment the `GIT_PARENT_DIRECTORY` line in `.env` and set it to a parent directory containing your private repositories.
1. Run `docker-compose up`
1. Visit `http://localhost:3080`. You should now see your local repositories listed on that page.

### Share it with your team

The real value of Sourcegraph is using it to ground technical discussions and share knowledge among your team. To share your instance of Sourcegraph with others,

1. Run `ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1'`
1. One of the IPs listed will be your machine's local IP. Sourcegraph will be accessible to teammates in your local network at URLs of the form `http://<local-ip>:3080`.
1. If you decide to deploy Sourcegraph at a domain name, update the `SRC_APP_HOST` value in `.env` to reflect the user-visible domain and restart Sourcegraph.

If you run into any issues with installation, please email support@sourcegraph.com. We typically respond within 24 hours.

## Restarting, resetting, and updates

To stop Sourcegraph, `Ctrl-C` the terminal that is running `docker-compose up`. Alternatively, you can `cd` into the Sourcegraph directory (the one that contains `docker-compose.yml` and this README) and run `docker-compose down`. You can restart Sourcegraph with `docker-compose restart`.

Sourcegraph persists data to the `.data` directory. To reset Sourcegraph, stop your Sourcegraph instance, delete the `.data` directory, and start Sourcegraph back up.

To update Sourcegraph run `docker-compose pull` in the Sourcegraph directory.

## Customization

### Remote repository host

Sourcegraph can index repositories stored on a remote Git host (e.g., GitHub Enterprise, Bitbucket Server, Gitlab, Gitolite, raw Git). We highly recommend you run through the installation instructions on local repositories above before running through the following steps to index remote Git repositories.

To index remote repositories:

1. Uncomment the appropriate lines in your `.env` file to define `SSH_KEYPAIR_FOLDER`, `ORIGIN_MAP`, and `ENSURE_REPOS_REMOTE`.
1. Run `docker-compose restart` in your Sourcegraph directory.
1. Visit `http://localhost:3080`. You should now see your remote repositories listed.

Note: you should only use Sourcegraph to index code that you trust. Do not index any repositories with untrusted code.

#### GitHub.com

Sourcegraph can index many repository hosts, including GitHub.com. If you would like your local Sourcegraph instance to index your organization's repositories on GitHub.com:

1. Follow the [instructions here](https://github.com/blog/1509-personal-api-tokens) to create a personal GitHub API token.
1. Replace `${GITHUB_USERNAME}`, `${GITHUB_PERSONAL_ACCESS_TOKEN}`, `${YOUR_ORG_NAME}` in the following bash command and run it.
   ```
   curl -u ${GITHUB_USERNAME}:${GITHUB_PERSONAL_ACCESS_TOKEN} "https://api.github.com/orgs/${YOUR_ORG_NAME}/repos?page=1&per_page=100" | grep "full_name" | awk -F '[": ]+' '{ print "github.com/"$3 }' | xargs echo
   ```
1. Set `ENSURE_REPOS_REMOTE` in your `.env` file to be the value of the output.

### SSL/TLS

1. Designate a TLS certificate. If you are serving at `localhost`, you can use the existing `sg.cer`/`sg.key` cert and key in the `config` directory.
   Otherwise, you'll need to generate a new TLS certificate. You can use a tool like [https://github.com/deckarep/EasyCert](https://github.com/deckarep/EasyCert).
   Move the generated certificate and key files to `config/sg.cer` and `config/sg.key`.
1. In `.env`, uncomment the appropriate `SRC_TLS_CERT`, `SRC_TLS_KEY`, and `CORS_ORIGIN` lines.
1. In `.env`, set `SRC_APP_HOST=https://localhost` (optionally replacing "localhost" with whatever your hostname is) and `SRC_APP_PORT=3443`.

### Chrome extension

1. Make sure you have the latest version of the extension installed.
1. Navigate to `chrome://extensions`, find the Sourcegraph extension item, and click the `Options` link.
1. Set the Sourcegraph URL to be the URL of your Sourcegraph instance.

**Note:** if your code host uses SSL/TLS, your Sourcegraph instance probably needs to use SSL/TLS (see the section on SSL/TLS).

### Java

#### Gradle

If you use Gradle as your build system, Sourcegraph by default statically analyzes Gradle files to extract the necessary build metadata. There are some cases in which this static analysis is insufficient.
In such cases, use the [Sourcegraph Gradle plugin](https://github.com/sourcegraph/sourcegraph-gradle-plugin).
Use the Gradle plugin to generate `javaconfig.json` configuration files, add these to your repository, and then try viewing the project in Sourcegraph.

#### Custom artifact host (Artifactory, Nexus)

Sourcegraph can be configured to work with your private Artifactory or Nexus instance. Simply set the following variables in your `.env` file:

```
PRIVATE_ARTIFACT_REPO_ID=${ARTIFACT_REPOSITORY_ID_FROM_YOUR_POM_OR_GRADLE}
PRIVATE_ARTIFACT_REPO_USERNAME=${YOUR_USERNAME}
PRIVATE_ARTIFACT_REPO_PASSWORD=${YOUR_PASSWORD}
```

## Privacy

By default, Sourcegraph collects usage data at the JavaScript layer and transmits this over HTTPS to a server controlled by Sourcegraph (the company). This data is similar to what is collected by many web applications and lets our team identify bugs and prioritize product improvements. This data DOES NOT include the contents of private source code files, but *does* include user actions like the following snippet:
```
{
    event_action: CLICK,
    event_category: Home,
    event_label: RepoButtonClicked,
    language: go,
    platform: Web,
    repo: github.com/gorilla/mux,
    page_title: HomePage
    path_name: github.com/gorilla/mux/mux.go,
}
```
Note that the above includes repository names and file names viewed in Sourcegraph. To see exactly what data is sent to Sourcegraph, open the JavaScript console in your browser and view network traffic to the `production` endpoint.

To disable all tracking, contact Sourcegraph support (support@sourcegraph.com).

## FAQ

### Why am I unable to view repositories from GitHub.com?

Sourcegraph indexes repositories from GitHub.com on demand, in response to user actions. By default, it uses the credentials in your `$HOME/.ssh` directory. If your GitHub SSH key is not in `$HOME/.ssh`, then auto-indexing GitHub.com repositories may fail.

### Why doesn't Sourcegraph auto-clone repositories from X code host?

Sourcegraph currently auto-clones only GitHub.com repositories. Repositories hosted elsewhere (including your local machine) are added via a separate code path. Local repositories are automatically added based on the local file path you specify to the containing directory (`GIT_PARENT_DIRECTORY` in your `.env` file). Remote repositories are specified by the `ENSURE_REPOS_REMOTE` variable in your `.env`.

## Troubleshooting

Contact support@sourcegraph.com with questions that aren't answered by the FAQ. Include the output of `docker-compose ps`, a copy of your `.env` file, and the logs from `docker-compose up`.
