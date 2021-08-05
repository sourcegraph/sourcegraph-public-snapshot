#Configuring Enviroment Variables
Depending on your deployment type, configuring your enviroment variables is done in different ways. NOTE: Ensure that you are not adding code that already exists , instead opting to adjust the code that already exists

##Sourcegraph/server
Add the following to your docker run command:

```
docker run [...]
-e (YOUR CODE)
sourcegraph/server:3.30.3
```

##Docker compose
Add/modify the environment variables to all of the sourcegraph-frontend-* services and the sourcegraph-frontend-internal service in [docker-compose.yaml](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/3.21/docker-compose/docker-compose.yaml):
```
sourcegraph-frontend-0:
  # ...
  environment:
    # ...
    - (YOUR CODE)
    # ...
```
See [“Environment variables in Compose”](https://docs.docker.com/compose/environment-variables/) for other ways to pass these environment variables to the relevant services (including from the command line, a .env file, etc.).

##Kubernetes
Update the environment variables in the sourcegraph-frontend deployment YAML file
