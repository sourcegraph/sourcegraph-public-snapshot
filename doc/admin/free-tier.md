# Sourcegraph Enterprise Free Plan FAQ

The new Sourcegraph Enterprise free plan will be enforced starting in Sourcegraph 4.5. You can find details on Sourcegraph's [various tiers and licenses here](https://docs.sourcegraph.com/getting-started/oss-enterprise).

## What's changing?

Previously, Sourcegraph Enterprises's free plan provided the following:
- SSO (OAuth, SAML, OpenID Connect, etc.)
- Unlimited private repositories can be synced
- 5 changesets for Batch Changes

In Sourcegraph 4.0, the free plan was adjusted to the following:
- SSO is no longer available
- 1 private repository can be synced
- 10 changesets for Batch Changes

However, SSO and private repository restrictions were not rolled out in Sourcegraph 4.0. These changes will be rolled out in Sourcegraph 4.5, releasing February 22, 2023.

## How will this impact new and existing Sourcegraph Enterprise installations?

For new Sourcegraph installations running Sourcegraph 4.5 or later, the new free plan will automatically be in place.

For existing Sourcegraph installations running Sourcegraph 4.4 or older, the old free plan will stay in place.

Any Sourcegraph instances without a license key that are upgraded to 4.5 or later will automatically move to the new free plan. Synced private repositories (in excess of the 1 repository limit) will no longer sync to the instance. **Please take note of this before upgrading your instance to Sourcegraph 4.5**.

## Contacting Sourcegraph

If you have any questions, or would like to upgrade to a Business-tier Sourcegraph license, you can [contact us here](https://about.sourcegraph.com/contact).
