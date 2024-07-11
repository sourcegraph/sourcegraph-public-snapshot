# enterprise-portal

**WIP** - refer to [RFC 885 Sourcegraph Enterprise Portal (go/enterprise-portal)](https://docs.google.com/document/d/1tiaW1IVKm_YSSYhH-z7Q8sv4HSO_YJ_Uu6eYDjX7uU4/edit#heading=h.tdaxc5h34u7q) for more details.

There are some services that are expected to be running by the Enterprise Portal:

- PostgreSQL with a database named `enterprise-portal`
- Redis running on `localhost:6379`

To start the Enterprise Portal, run:

```zsh
sg run enterprise-portal
```

To customize the PostgreSQL and Redis connection strings, customize the following environment variables in your `sg.config.overwrite.yaml` file:

- `PGDSN`: PostgreSQL connection string
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port
