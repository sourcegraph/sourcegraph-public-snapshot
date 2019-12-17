# LightStep integration with Sourcegraph

[LightStep](https://lightstep.com) is an application performance management (APM) tool that supports distributed tracing and the [OpenTracing](https://opentracing.io/) standard.

Sourcegraph integrates with LightStep both for users (to easily view live traces when navigating code) and site admins (to monitor a self-hosted Sourcegraph instance).

Feature | Supported?
------- | ----------
For users: [View live traces for your code](#view-live-traces-for-your-code) | ✅
For site admins: [Monitoring a self-hosted Sourcegraph instance](#instrumenting-a-self-hosted-sourcegraph-instance) | ✅

# View live traces for your code

Any user can [enable the LightStep extension for Sourcegraph](https://sourcegraph.com/extensions/sourcegraph/lightstep) to view live traces for OpenTracing spans in their own code.

![Screenshot](https://storage.googleapis.com/sourcegraph-assets/LightStep_Sourcegraph.png)

# Monitoring a self-hosted Sourcegraph instance

In the site configuration, site admins can [configure LightStep tracing](../admin/config/site_config.md) using the `lightstepAccessToken` and `lightstepProject` configuration properties.

> NOTE: In Sourcegraph versions before v3.11, these options were located in the [critical configuration](../admin/config/critical_config.md).