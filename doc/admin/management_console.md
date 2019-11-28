# Management console

The management console is a separate service used to edit Sourcegraph's [critical configuration](config/critical_config.md).

Critical configuration includes things like authentication providers, the external URL, and the license key. This configuration is separate from the regular [site configuration](config/site_config.md), because an error here could make Sourcegraph inaccessible, except through the management console.

## Accessing the management console

### When running Sourcegraph in a single Docker container

The management console is built-in to the same Docker image and published on port 2633:

```
$ docker ps
CONTAINER ID        IMAGE                              PORTS
394ff36a8c3c        sourcegraph/server:3.10.1           0.0.0.0:2633->2633/tcp, 0.0.0.0:7080->7080/tcp
```

Usually, you can access it through the public internet via https://my.server.ip:2633, or https://localhost:2633 when testing locally.

### When running Sourcegraph in a cluster deployment

The management console is a separate service running in your cluster. You will need to either port-forward / VPN into your cluster in order to access the service via a browser, or you may expose the service to the public internet (which is generally secure) and access it with a browser through the public internet. For example, using Kubernetes you can forward port 2633 of the management console service to your local machine:

```
$ kubectl port-forward svc/management-console 2633:2633
```

Then visit https://localhost:2633 to access the management console.

## Troubleshooting

### I am getting "The server sent an invalid response" errors from my browser, why?

Ensure you are connecting via `https://` and _not_ `http://`.

This type of browser error often indicates you are connecting via HTTP instead of HTTPS. The management console **only** serves over HTTPS for security reasons.

### I am seeing TLS / SSL warnings in my browser, why?

You must click the **Advanced** option in your browser and continue anyway (**Proceed to localhost (unsafe)**).

The management console uses self-signed TLS certificates by default. The first time you connect, your browser will warn you that your connection is insecure and/or that the TLS certificate is invalid.

The self-signed TLS certificate in use ensures that your interaction with the management console cannot be sniffed via [MITM attacks](https://en.wikipedia.org/wiki/Man-in-the-middle_attack). If desired, you can configure the management console to use your own TLS certificates (but you do not need to).

### What is the password to my management console?

-> Visit Sourcegraph's **Site admin** area (https://sourcegraph.example.com/site-admin) to retrieve your management console password (**enter any username**):

![image](https://user-images.githubusercontent.com/3173176/50871227-3eac6700-1378-11e9-8ba7-4c712e622039.png)

Once dismissed, you will not be able to retrieve the password again, but you can generate a new one.

This password is automatically generated for you to ensure it is very long and secure, even if the management console is exposed to the public internet.

### How can I reset my management console password?

Resetting the management console password can be done manually via the database.

#### (Option 1): If you have access to Sourcegraph still

Open a `psql` prompt on your Sourcegraph instance (see ["How do I access my Sourcegraph database?"](faq.md#how-do-i-access-the-sourcegraph-database)) and run:

```sql
UPDATE global_state SET mgmt_password_plaintext='', mgmt_password_bcrypt='';
```

When you next visit sourcegraph.example.com/site-admin, a new password will be generated and presented to you.

#### (Option 2) If you _don't_ have access Sourcegraph

First, determine what your new password will be. If your management console is exposed to the public internet, it is important that this be a **very** long and random password (e.g. 128 characters in length). For this example, we will use `abc123`.

Second, bcrypt your password on any machine:

1. Install Python
2. `pip install bcrypt`
3. Encrypt your password:

    a. (**If you have Python 2**):

    ```bash
    PASSWORD='abc123' python -c "import bcrypt; import os; print(bcrypt.hashpw(os.environ['PASSWORD'], bcrypt.gensalt(15)))"
    ```

    b. (**If you have Python 3**): 

    ```bash
    PASSWORD='abc123' python -c "import bcrypt; import os; print(bcrypt.hashpw(os.environ['PASSWORD'].encode('utf-8'), bcrypt.gensalt(15)))"
    ```

Finally, open a `psql` prompt on your Sourcegraph instance (see ["How do I access my Sourcegraph database?"](faq.md#how-do-i-access-the-sourcegraph-database)) and run:

```sql
UPDATE global_state SET mgmt_password_bcrypt='my-encrypted-password';
```

You may now sign into the management console using your plaintext password `abc123`.

### How can I use my own TLS certificates with the management console?

The management console looks for TLS certificates in the following location inside the Docker container:

- `/etc/sourcegraph/management/cert.pem`
- `/etc/sourcegraph/management/key.pem`

If you are using `sourcegraph/server` and the regular Docker flag:

```
--volume ~/.sourcegraph/config:/etc/sourcegraph
```

This means you can simply place them in `~/.sourcegraph/config/management/`  on the host.

Restart the container once you have copied the files there for the changes to take effect.

### Can I disable HTTPS on the management console?

It is **unsafe** to do so as anyone who can MITM your traffic to the management console can steal the admin password and act on your behalf.

If you understand the risks and still wish to, you can set the environment variable `UNSAFE_NO_HTTPS` to `true` on the Docker container. This will entirely disable HTTPS (the port will remain the same).
