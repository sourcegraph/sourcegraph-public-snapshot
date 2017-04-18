# Sourcegraph Origin

Sourcegraph Origin is a downloadable distribution of Sourcegraph that can be run on any machine. It is designed to scale up to 10 repositories and 10 users. For larger teams, Sourcegraph Origin can be used for evaluation purposes.

## System requirements

Sourcegraph Origin should run on any OS where Docker can be installed (includes most modern macOS, Linux, and Windows systems).

* Install [Docker Compose](https://docs.docker.com/compose/install/).
* If you did not install Docker while installing Docker Compose, [install Docker](https://docs.docker.com/engine/installation/), as well.

## Credentials

Sourcegraph Origin is currently available to a restricted set of users. If you are part of this set, you should have received a username and password to the Sourcegraph Origin Docker container registry. Before starting, run the following command to authorize your Docker client to fetch from this registry:

```
docker login -u <username> -p <password> docker.sourcegraph.com
```

## Install

1. `cd` into the directory containing this README and run `docker-compose up`.<br>
*Note: you may see some messages in the Sourcegraph frontend logs about connecting to PostgreSQL or Redis on startup. These are usually innocuous.*
1. Visit http://localhost:3080/github.com/gorilla/mux. This adds a small open-source repository to your Sourcegraph Origin instance. You should see a message indicating the repository is cloning, followed by a file browser after the repository is cloned.
1. That's it—you can start exploring the code. For example, open [`mux.go`](http://localhost:3080/github.com/gorilla/mux/-/blob/mux.go) and try clicking on some function names or search for symbols using the '/#' hotkey combination.

Now, let's add your private repositories:

1. Stop the Sourcegraph instance with `Ctrl-C` in the terminal running `docker-compose up`.
1. Run `cp env.example .env`<br>
   (The ".env" file sets default values for environment variables when running Docker Compose.)
1. Uncomment the `GIT_PARENT_DIRECTORY` line in `.env` and set it to the local parent directory containing your private repositories.
1. Run `docker-compose up`
1. Visit `http://localhost:3080/local/\<repository-id\>`. The \<repository-id\> is the relative path from `GIT_PARENT_DIRECTORY` to the repository root directory on local disk.

## Restarting, resetting, and updates

To stop Sourcegraph, `Ctrl-C` the terminal that is running `docker-compose up`. Alternatively, you can `cd` into the directory and run `docker-compose down`.

To update Sourcegraph run `docker-compose pull` in that directory.

Sourcegraph Origin persists data to the `.data` directory. To reset Sourcegraph Origin, stop your Sourcegraph Origin instance, delete the `.data` directory, and restart the Sourcegraph Origin instance.

## Privacy

Sourcegraph Origin collects usage data in the web UI and transmits this over HTTPS to a server controlled by Sourcegraph. This data is similar to what is collected by many web applications and lets our team identify bugs and prioritize product improvements. This data DOES NOT include private source code, but *does* include the following data:
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
To see exactly what data is sent to Sourcegraph, open the JavaScript console in your browser, and view network traffic to the `production` endpoint.

To disable all tracking, contact Sourcegraph support (support@sourcegraph.com) and we will send simple instructions to do so.

## FAQ

### Unable to clone repositories from GitHub.com

Sourcegraph Origin indexes repositories from GitHub.com on demand, in response to user actions. By default, it uses the credentials in your `$HOME/.ssh` directory. If your GitHub SSH key is not in `$HOME/.ssh`, then auto-indexing GitHub.com repositories may fail.

## Troubleshooting

Contact support@sourcegraph.com with questions that aren't answered by the FAQ. Include the output of `docker-compose ps`, a copy of your `.env` file, and the logs from `docker-compose up`.
