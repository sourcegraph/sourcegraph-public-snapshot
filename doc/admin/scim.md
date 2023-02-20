# SCIM

**⚠️ DISCLAIMER: SCIM SUPPORT IS UNDER DEVELOPMENT. THE ENDPOINT DOES NOT COMPLY WITH THE SCIM STANDARD YET AND IT IS NOT READY FOR PRODUCTION USE.**

SCIM (System for Cross-domain Identity Management) is a standard for provisioning users and groups in an organization. It is supported by many IdP (identity providers) such as Okta, OneLogin, and Azure Active Directory.

Sourcegraph supports SCIM for provisioning and de-provisioning _users_, but not groups at this time.

You can use any IdP that supports SCIM, but we’ve only tested the endpoint with Okta and Azure Active Directory.

## How to use

To use SCIM, you must have an existing IdP configured with your Sourcegraph instance.
To configure it, add the following line to your [site configuration](config/site_config.md):

```
"scim.authToken": "{your token}"
```

Currently, we only support Bearer token authentication.

Then you can set up your IdP to to use the SCIM endpoint. The API is at `https://sourcegraph.company.com/.api/scim/v2`, so the “Users” endpoint is at `https://sourcegraph.company.com/.api/scim/v2/Users`.

