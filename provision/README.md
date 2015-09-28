# Sourcegraph Provisioning

This has moved to the infrastructure repo. Please see the README there.

#### Viewing server logs

Deployed instances send their logs to
[Papertrail](http://papertrailapp.com/). You can search and tail the
logs with
[papertrail-cli](https://github.com/papertrail/papertrail-cli) (or
from the Papertrail web interface, if you prefer).

Install the `papertrail` CLI by running:

```
sudo gem install papertrail
echo "token: 1234" > ~/.papertrail.yml
```

Replace `1234` with your Papertrail API token, obtained from
https://papertrailapp.com/user/edit.

To tail the logs for the www.sourcegraph.com frontend servers:

```
papertrail -f --group www-production-frontend
```

To search the logs for the term `Repos.List`:

```
papertrail --group www-production-frontend Repos.List
```

See `papertrail --help` for full usage info.
