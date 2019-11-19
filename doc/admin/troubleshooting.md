# Administration troubleshooting

### Docker Toolbox on Windows: `New state of 'nil' is invalid`

If you are using Docker Toolbox on Windows to run Sourcegraph, you may see an error in the `frontend` log output:

```bash
frontend |
     frontend |
     frontend |
     frontend |     New state of 'nil' is invalid.
```

After this error, no more `frontend` log output is printed.

This problem is caused by [docker/toolbox#695](https://github.com/docker/toolbox/issues/695#issuecomment-356218801) in Docker Toolbox on Windows. To work around it, set the environment variable `LOGO=false`, as in:

```bash
docker container run -e LOGO=false ... sourcegraph/server
```

See [sourcegraph/sourcegraph#398](https://github.com/sourcegraph/sourcegraph/issues/398) for more information.

### Submitting a metrics dump

If you encounter performance or instability issues with Sourcegraph, we may ask you to submit a metrics dump to us. This allows us to inspect the performance and health of various parts of your Sourcegraph instance in the past and can often be the most effective way for us to identify the cause of your issue.

The metrics dump includes non-sensitive aggregate statistics of Sourcegraph like CPU & memory usage, number of successful and error requests between Sourcegraph services, and more. It does NOT contain sensitive information like code, repository names, user names, etc.

#### Single-container `sourcegraph/server` deployments

To create a metrics dump from a single-container `sourcegraph/server` deployment, run this command on the host machine:

```sh
cd ~/.sourcegraph/data/prometheus && tar -czvf /tmp/sourcegraph-metrics-dump.tgz .
```

If needed, you can download the metrics dump to your local machine (current directory) using `scp`:

```sh
scp -r username@hostname:/tmp/sourcegraph-metrics-dump.tgz .
```

Please then upload the `sourcegraph-metrics-dump.tgz` for Sourcegraph support to access it. If desired, we can send you a shared private Google Drive folder for the upload as it can sometimes be a few gigabytes.
