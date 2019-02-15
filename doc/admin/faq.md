# Adminstration FAQ

## How do I expose my Sourcegraph instance to a different host port when running locally?

Change the `docker` `--publish` argument to make it listen on the specific interface and port on your host machine. For example, `docker run ... --publish 0.0.0.0:80:7080 ...` would make it accessible on port 80 of your machine. For more information, see "[Publish or expose port](https://docs.docker.com/engine/reference/commandline/run/#publish-or-expose-port--p---expose)" in the Docker documentation.

The other option is to deploy and run Sourcegraph on a cloud provider. For an example, see documentation to [deploy to Google Cloud](install/docker/google_cloud.md).

## How do I access the Sourcegraph database?

For single-node deployments (`sourcegraph/server`), follow these steps on the machine that is running the Sourcegraph Docker container:

1.  Get the Docker container ID for Sourcegraph:
    ```
    $ docker ps
    CONTAINER ID        IMAGE
    d039ec989761        sourcegraph/server:VERSION
    ```
2.  Open a PostgreSQL interactive terminal:
    ```
    docker exec -it d039ec989761 psql -U postgres sourcegraph
    ```
3.  Run your SQL query:
    ```
    select * from users;
    ```

For Kubernetes cluster deployments:

1.  Get the id of one pgsql pod:
    ```
    $ kubectl get pods -l app=pgsql
    NAME                     READY     STATUS    RESTARTS   AGE
    pgsql-76a4bfcd64-rt4cn   2/2       Running   0          19m
    ```
2.  Open a PostgreSQL interactive terminal:
    ```
    kubectl exec -it pgsql-76a4bfcd64-rt4cn -- psql -U sg
    ```
3.  Run your SQL query:
    ```
    select * from users;
    ```

## Troubleshooting

### Docker Toolbox on Windows: `New state of 'nil' is invalid`

If you are using Docker Toolbox on Windows to run Sourcegraph, you may see an error in the `frontend` log output:

```
     frontend |
     frontend |
     frontend |
     frontend |     New state of 'nil' is invalid.
```

After this error, no more `frontend` log output is printed.

This problem is caused by [docker/toolbox#695](https://github.com/docker/toolbox/issues/695#issuecomment-356218801) in Docker Toolbox on Windows. To work around it, set the environment variable `LOGO=false`, as in:

```shell
docker run -e LOGO=false ... sourcegraph/server
```

See [sourcegraph/sourcegraph#398](https://github.com/sourcegraph/sourcegraph/issues/398) for more information.
