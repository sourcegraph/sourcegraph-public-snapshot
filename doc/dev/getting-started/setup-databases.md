# Setting up databases

The sourcegraph codebase requires a few databases to be installed in order to run locally. We recommend and support a default approach that is suited for everyone, regardless of their familiarity with the topic. 

It's totally possible to use alternative ways to install databases if you need to: `sg` checks will still perform correctly, but you're on your own for the installation process: see the [alternative instructions section](#alternative-instructions).

## Recommended instructions

### MacOs

Requirements: 

- The `github.com/sourcegraph/sourcegraph` repository cloned in a folder of your choice.
- [`Homebrew`](https://brew.shell) is installed.

Instructions:

- Install _Postgres.app_:
  1. Open your browser and navigate to this page: [https://postgresapp.com](https://postgresapp.com)
  1. Follow the installation instructions.
- Install _Redis.app_:
  1. Open your browser and navigate to this page: [https://jpadilla.github.io/redisapp/](https://jpadilla.github.io/redisapp/)
  1. Click on the Download button and install the app in the archive.
  1. Open a terminal and type: `sudo mkdir -p /etc/paths.d && echo /Applications/Redis.app/Contents/Resources/Vendor/redis/bin | sudo tee /etc/paths.d/redisapp`

### Ubuntu

## Alternative instructions

### MacOs

#### Using Homebrew

1. `brew install postgres redis`
1. `brew services start postgresql`
1. `brew services start redis`

### Docker 

- Assuming that `docker` is installed on your system, you can run `sg run redis-postgres`  to start the databases.
