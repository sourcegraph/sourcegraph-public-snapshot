+++
title = "Frequently Asked Questions"
navtitle = "FAQ"
+++

## How to get Sourcegraph logs

If you deployed using a cloud install script on Digital Ocean, AWS, or EC2
(or if running with upstart on Linux):

```bash
$ less /var/log/upstart/src.log
```

If you're running Sourcegraph locally on a Mac, logs are printed
to stdout.

## How to get Sourcegraph version

```bash
$ src version
```
