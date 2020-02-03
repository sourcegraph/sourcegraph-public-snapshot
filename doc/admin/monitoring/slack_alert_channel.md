# Sourcegraph monitoring: setting up a Slack alert channel

1. You need to access the Grafana instance running inside your Sourcegraph instance directly (as opposed to through the 
site admin Monitoring page). Please follow [these instructions](../monitoring_and_tracing.md#Accessing Grafana directly)
 to do that.

1. You need to have Grafana admin privileges to set up a new alert channel. To do that append `/login` to the direct
 Grafana access URL you set up in step 1. If you are doing this the first time, use username/password `admin/admin` and
 then change to a password of your choice. Otherwise use your admin username/password combination.
 
1. Click on the `Bell` icon on the left sidebar and select `Notification channels.`

1. Click on `Add Channel.` and select Slack from the `Type` pulldown. Give the channel a name (this is to identify it 
in the Grafana alert channels list and does not specify to which Slack chnannel the alerts are going).

1. In a separate browser tab go to `https://api.slack.com/apps` to create a new Slack App.

    1. Click `Create an App` button, give the new app a name and click `Create App` button.
    1. Click `Incoming Webhooks` and toggle `Activate Incoming Webhooks` to on.
    1. Click `Add New Webhook to Workspace`.
    1. Pick a channel where this Slack App will post and authorize.
    1. Copy the web hook.

1. Paste the web hook into the channel form Url field.

1. Test by clicking `Send Test` button. You should see a test alert in the Slack channel you selected when you created
the web hook.

1. If everything works, save your new alert channel.

> NOTE: Alerts have a link back to the relevant Grafana panel. In order for these links to work properly Grafana needs
> to know under which external URL it is running (note: this is usually different from the direct access URL you used
> earlier). Set the environment variable `GF_SERVER_ROOT_URL` to your Sourcegraph instance external URL followed
> by the path `/-/debug/grafana`.

## Other alert channel types

Other alert channel types are configured in a way similar to the Slack alert channel type described above. Choose the
appropriate channel type and provide the necessary information for that type.
