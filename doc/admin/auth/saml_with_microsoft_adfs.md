# Configuring SAML with Microsoft Active Directory Federation Services (ADFS)

This document applies to the following versions of [Microsoft Active Directory Federation Services (ADFS)](https://docs.microsoft.com/en-us/windows-server/identity/active-directory-federation-services):

- ADFS 2.1 (Windows Server 2012)
- ADFS 3.0 (Windows Server 2012 R2)
- ADFS 4.0 (Windows Server 2016)

These instructions guide you through configuring Sourcegraph as a relying party (RP) of ADFS, which enables users to authenticate to Sourcegraph using their Active Directory credentials.

## 1. Add the SAML auth provider to Sourcegraph site config

1.  Set the `externalURL` in [site config](../config/site_config.md) to a URL that the ADFS server can reach.
1.  Add an entry to `auth.providers` that points to your ADFS server's SAML metadata URL (typically containing the path `/federationmetadata/2007-06/federationmetadata.xml`).
1.  Confirm there are no error messages in the `sourcegraph/server` Docker container logs (or the `sourcegraph-frontend` pod logs, if Sourcegraph is deployed to a Kubernetes cluster). The most likely error message indicating a problem is `Error prefetching SAML service provider metadata.`.

The example below demonstrates the properties that you must set. See the [SAML auth provider documentation](../config/site_config.md#saml) the full set of properties that the SAML auth provider supports.

```json
{
  // ...
  "externalURL": "https://sourcegraph.example.com",
  "auth.providers": [
    {
      "type": "saml",
      "identityProviderMetadataURL": "https://adfs.example.com/federationmetadata/2007-06/federationmetadata.xml"
    }
  ]
}
```

## 2. Add Sourcegraph as a relying party (RP) to ADFS

1.  In Windows Server's **Server Manager**, open **Tools > AD FS Management**.
1.  In the right sidebar, click the **Add Relying Party Trust...** action.
1.  Proceed through the "Add Relying Party Trust Wizard" as follows:
    - Welcome (Page 1)
      - Click **Start**.
    - Select Data Source (Page 2)
      - Import data about the relying party published online or on a local network: `https://sourcegraph.example.com/.auth/saml/metadata`
    - Specify Display Name (Page 3)
      - Leave everything unchanged and click **Next >**.
    - Choose Issuance Authorization Rules (Page 4)
      - Leave everything unchanged (**Permit all users to access this relying party** is selected) and click **Next >**.
    - Ready to Add Trust (Page 5)
      - Click **Next >**.
    - Ready to Add Trust (Page 5)
      - Click **Next >**.
    - Finish (Page 6)
      - Check the box for **Open the Edit Claim Rules dialog...**.
      - Click **Close**.

Next, in the "Edit Claim Rules for sourcegraph.example.com" window in the "Issuance Transform Rules" tab, add the 2 following rules.

- _Send User Info rule:_ Click **Add Rule...** and proceed through the "Add Transform Claim Rule Wizard" as follows:
  - Choose Rule Type (Page 1)
    - Claim rule template: `Send LDAP Attributes as Claims`
    - Click **Next >**.
  - Configure Claim Rule (Page 2)
    - Claim rule name: `Send User Info` (any value is OK)
    - Attribute store: `Active Directory`
    - Mapping of LDAP attributes to outgoing claim types:<br/> `E-Mail-Addresses` -> `E-Mail Address`<br/> `Display-Name` -> `Name`<br/><br/>
    - Click **Finish**.
- _Email to NameID rule:_ Click **Add Rule...** and proceed through the "Add Transform Claim Rule Wizard" as follows:
  - Choose Rule Type (Page 1)
    - Claim rule template: `Transform an Incoming Claim`
    - Click **Next >**.
  - Configure Claim Rule (Page 2):
    - Claim rule name: `Email to NameID`
    - Incoming claim type: `E-Mail Address`
    - Outgoing claim type: `Name ID`
    - Outgoing name ID format: `Persistent identifier`
    - Select **Pass through all claim values**.
      - Click **Finish**.

Click **OK** to apply the new claim rules and close the window.

## Authenticate to Sourcegraph using ADFS

All configuration is now complete. Let's test that it works.

1.  Visit `https://sourcegraph.example.com`. (If you are already authenticated from before configuring the SAML auth provider, sign out of Sourcegraph.)
1.  When prompted to sign into ADFS, provide the `alice` credentials and continue.
1.  Confirm that you are authenticated to Sourcegraph as `alice`.
