# Adminstration troubleshooting

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
