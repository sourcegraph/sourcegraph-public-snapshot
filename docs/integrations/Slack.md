+++
title = "Slack"
+++

You can configure Sourcegraph to send notifications to Slack by setting these flags in `/etc/sourcegraph/config.ini` and restarting your
Sourcegraph server with `sudo restart src`:

```
[serve.Slack]
WebhookURL = https://hooks.slack.com/services/XXXXXX/YYYYY
DefaultChannel = dev-bot
```
