# Prometheus provisioning

We use `docker-compose` to manage our prometheus stack. This allows you to
easily replicate the stack on your localhost. To deploy just run

```bash
make deploy
```

You can access prometheus on http://metrics.sgdev.org/ Note that this is only
available on the VPN network, so you need to connect through the bounce box
(see `../bounce/README.md`)

## Kubernetes
The current setup relies on AWS. However, it shouldn't be much effort to port
over once we need to.

## Local

To do this locally you need to set the correct environment variables, update
`docker-compose.yml` to work on OS X and update your `/etc/hosts` to use your
local instances:

```bash
export AWS_SECRET_ACCESS_KEY=...
export AWS_ACCESS_KEY=...
export DATABASE_URL=sqlite3:/promdash-data/file.sqlite3
docker-compose up
```

To point the domains at your local instance:

```
# /etc/hosts
192.168.99.100 prom.metrics.sgdev.org
192.168.99.100 dash.metrics.sgdev.org
192.168.99.100 push.metrics.sgdev.org
192.168.99.100 alert.metrics.sgdev.org
192.168.99.100 metrics.sgdev.org
```

where `192.168.99.100` is the ip of your docker machine (`boot2docker ip` or
`docker-machine ip docker-vm` on a OS X). Remember to undo this change when
you are done.


If it's the first time you are running, you may need to setup the DB

```bash
docker run -v $PWD/data/promdash:/promdash-data -e DATABASE_URL prom/promdash ./bin/rake db:migrate sqlite3:/promdash-data/file.sqlite3
```

## Storage

All data is bind-mounted to `./data`. Ensure you have enough space on your
filesystem (`500GB` should be sufficient for production use).
