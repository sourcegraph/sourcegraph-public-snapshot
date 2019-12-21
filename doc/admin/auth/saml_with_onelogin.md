# Configuring SAML with OneLogin

When configuring SAML authentication via [OneLogin](https://www.onelogin.com/saml), ensure that it sends the `email` attribute to Sourcegraph:

1.  Go to the OneLogin app's **Parameters** tab.
1.  Click **Add parameter**. In the **New Field** screen:
    - **Field name:** `email`
    - **Flags:** Check the box labeled **Include in SAML assertion**
1.  Click **Save**.

Confirm there are no error messages in the `sourcegraph/server` Docker container logs (or the `sourcegraph-frontend` pod logs, if Sourcegraph is deployed to a Kubernetes cluster). The most likely error message indicating a problem is `Error prefetching SAML service provider metadata`. See [SAML troubleshooting](/admin/auth/saml#saml-troubleshooting) for more tips.
