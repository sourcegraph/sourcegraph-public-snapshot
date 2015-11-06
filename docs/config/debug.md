+++
title = "Debug mode"
navtitle = "Debug mode"
+++

To troubleshoot your Sourcegraph installation, you can run the server in debug mode by setting `DEBUG=1` in the server's shell environment as follows:

# Mac OS X

On OS X, start Sourcegraph as:

```
DEBUG=1 src serve
```

# Ubuntu Linux and Cloud installations

If you are running Sourcegraph on Ubuntu Linux or one of the supported cloud providers, you can edit the `/etc/sourcegraph/config.env` file to export the `DEBUG` variable in the server's environment.