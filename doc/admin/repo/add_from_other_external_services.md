# Add repositories from other external services

If your Git repositories are hosted in an external service for which a direct integration isn't yet available, you can still connect them with Sourcegraph.

Simply go to `https://sourcegraph.example.com/site-admin/external-services/add?kind=other` and [configure](https://docs.sourcegraph.com/admin/site_config/all#otherexternalserviceconnection-object) the list of Git clone URLs you want Sourcegraph to discover.

## Browser extension

The [Sourcegraph browser extension](../../integration/browser_extension.md) will work only with some of the external services for which it has a first class integration. If you're connecting your Git repositories here, it likely won't.
