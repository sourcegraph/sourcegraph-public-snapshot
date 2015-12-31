+++
title = "Enterprise configuration"
linktitle = "Enterprise"
description = "Use Sourcegraph for enterprise"
+++

If you are running Sourcegraph within a large enterprise, you may
be interested in some of the following features.

Looking for something else? [Let us know!](mailto:help@sourcegraph.com)

# Centralized LDAP

To have Sourcegraph authenticate users via your directory server,
follow the [LDAP configuration instructions]({{< relref "config/authentication.md" >}}).

# Standalone operation

To airgap your Sourcegraph server from Sourcegraph.com, add the following to your `config.ini`:

```
[serve]
fed.is-root = true
```

[Read more about Sourcegraph data collection.]({{< relref "management/data_collection.md" >}})

# High scalability and availability

Sourcegraph runs on Sourcegraph.com for hundreds of thousands of repositories and
tens of thousands of users. It can be configured to use multiple storage systems including
key/value stores and SQL databases.

[Contact us](mailto:help@sourcegraph.com) if you have questions or need technical assistance.

## Horizontal scaling

To run multiple instances of Sourcegraph with a shared SQL database backend, you must share
an RSA key across your instances. [Read more about client authentication]({{< relref "dev/OAuth2.md" >}}).
Follow these instructions to enable horizontal scaling of Sourcegraph servers:

1. Generate RSA key pair:

	```
	openssl genrsa -out private.pem 2048
	```

2. Register client:

	```
	src -u https://sourcegraph.com registered-clients create -i private.pem --client-name="My Company" --client-uri="http://src.mycompany.com"
	```

3. Base64 encode the private key:

	```
	ENCODED_KEY=`cat private.pem | openssl base64 | tr -d '\n'`
	```

4. Set env variable for all instances:

	```
	echo "export SRC_ID_KEY_DATA=base64:${ENCODED_KEY}" >> /etc/sourcegraph/config.env
	```

# License?

Sourcegraph uses the [Fair Source License](https://fair.io) with a
user limit of 15. It is free to use Sourcegraph Enterprise for teams
with 15 or fewer developers. After your organization hits that user
limit, you must pay a licensing fee.

Sourcegraph Server (not Enterprise) is free to use for teams of any size.

Have questions? [Contact us.](mailto:help@sourcegraph.com)
