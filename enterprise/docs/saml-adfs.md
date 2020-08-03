# Testing SAML with Microsoft Active Directory Federation Services (ADFS) 2.1

Follow all the steps in this document in order.

## Create an Azure VM running Windows Server 2012

1.  Create an Azure VM using the [Windows Server 2012 Datacenter](https://portal.azure.com/#create/Microsoft.WindowsServer2012Datacenter) image.
    - ⚠ DO NOT use "Windows Server 2012 R2 Datacenter" (because R2 ships with ADFS 3.0, and this document is about testing ADFS 2.1 specifically).
    - After clicking on the link above (and, if necessary, signing into Azure and setting up a subscription):
      - Select a deployment model: `Resource Manager`
      - Click **Create**.
    - Basics (Page 1)
      - Name: `adfstest`
      - Username: `alice` (any value is OK)
      - Password: choose a strong password (the instance will be Internet-accessible)
      - Resource group: check the box for **Create new** and enter a name of `adfstest-rg` (any value is OK)
      - Leave everything else unchanged.
      - Click **OK**.
    - Size (Page 2)
      - Choose **B2s** (although smaller instances are probably OK).
      - Click **Select**.
    - Settings (Page 3)
      - Leave everything unchanged and click **OK**.
    - Summary (Page 4)
      - Click **Create**.
    - Create a user named `alice` with a secure password.
1.  Configure the firewall so you can access the VM on port 443.
    1.  Open **Settings > Networking** for the VM.
    1.  Click **Add inbound port rule** and create a rule with the following values:
        - Source: `IP Addresses`
        - Source IP addresses/CIDR ranges: `1.2.3.4/32` (where `1.2.3.4` is your external IPv4 address)
        - Destination port ranges: `*`
        - Name: `AllowAllFromMyIP` (any value is OK)
        - Leave all other values unchanged.
        - Click _Add_.
1.  On your own computer, make the Azure VM accessible at the hostname `adfstest.sgdev.org`:
    - In `/etc/hosts`, add an entry `5.6.7.8 adfstest.sgdev.org` (where `5.6.7.8` is the **Public IP** for the Azure VM).
1.  Connect to it using Remote Desktop
    - [Remmina](https://www.remmina.org/wp/) is a good Linux RDP client.

## Add/install the initial set of Active Directory roles

1.  In **Server Manager > Dashboard**, click **(2) Add roles and features**.
1.  Proceed through the "Add Roles and Features Wizard" as follows:
    - Before You Begin (Page 1)
      - Click **Next >**.
    - Installation Type (Page 2)
      - Leave everything unchanged (**Role-based or feature-based installation** is selected) and click **Next >**.
    - Server Selection (Page 3)
      - Leave everything unchanged and click **Next >**.
    - Server Roles (Page 4)
      - Check the following boxes:
        - **Active Directory Domain Services**
        - **Active Directory Federation Services**
        - **Active Directory Lightweight Directory Services**
        - ⚠️ DO NOT check the box for **Active Directory Certificate Services**. You will install this later. Installing it now will prevent you from completing the next section.
      - In the dialogs asking "Add features that are required for Active Directory ...?" that appear as you check the boxes, leave everything unchanged and click **Add Features**.
      - Click **Next >**.
    - Features (Page 5)
      - Leave everything unchanged and click **Next >**.
    - AD DS, AD FS, Role Services (Pages 6-8)
      - Leave everything unchanged and click **Next >** for each of these until you arrive at the **Configuration** page.
    - Confirmation (Page 9)
      - Check the box for **Restart the destination server automatically if required**.
      - In the "If a restart is required, ... Do you want to allow automatic restarts?" dialog, click **Yes**.
      - Click **Install**. The VM will restart after installation is complete.
1.  After the VM has restarted, reconnect to it using Remote Desktop.

1.  In the **Server Manager > Dashboard**, click **Add roles and features** and install all of the items whose name starts with "Active Directory".

## Configure an Active Directory domain

1.  In the left sidebar of the **Server Manager**, select **AD DS**.
1.  Click the **More...** link in the top right in the yellow alert that reads "Configuration required for Active Directory Domain Services at ADFSTEST".
1.  Click the **Promote this server to a domain controller** link in the "Action" column.
1.  Proceed through the "Active Directory Domain Services Configuration Wizard" as follows:
    - Deployment Configuration (Page 1)
      - Select the deployment operation: `Add a new forest`
      - Root domain name: `sgdev.org`
      - Click **Next >**. You may need to wait 1-2 minutes as it does something (no clue what). The progress bar is animated during this time.
    - Domain Controller Options (Page 2)
      - Type the Directory Services Restore Mode (DSRM) password: use the same password as for `alice`
      - Leave everything else unchanged.
      - Click **Next >**.
    - Domain Controller Options > DNS Options (Page 2.1)
      - Ignore the yellow alert that reads "A delegation for this DNS server cannot be created...".
      - Click **Next >**.
    - Additional Options (Page 3)
      - The NetBIOS domain name: `SGDEV`
      - Click **Next > **.
    - Paths (Page 4)
      - Leave everything unchanged. Click **Next >**.
    - Review Options (Page 5)
      - Click **Next >**.
    - Prerequisites Check (Page 6)
      - The following alerts in "View results" can be ignored:
        - "Windows Server 2012 domain controllers have a default for the security setting named 'Allow cryptography algorithms..."
        - "This computer has at least one physical network adapter that does not have static IP address(es)..."
        - "A delegation for this DNS server cannot be created because the authoritative parent zone..."
      - Click **Install**. The VM will restart after installation is complete.
1.  After the VM has restarted, reconnect to it using Remote Desktop.

## Install Active Directory Certificate Services

1.  In **Server Manager > Dashboard**, click **(2) Add roles and features**.
1.  Proceed through the "Add Roles and Features Wizard" as follows:
    - Before You Begin (Page 1)
      - Click **Next >**.
    - Installation Type (Page 2)
      - Leave everything unchanged (**Role-based or feature-based installation** is selected) and click **Next >**.
    - Server Selection (Page 3)
      - Leave everything unchanged and click **Next >**.
    - Server Roles (Page 4)
      - Check the box for **Active Directory Certificate Services**.
      - In the "Add features that are required for Active Directory Certificate Services?" dialog, leave everything unchanged and click **Add Features**.
      - Click **Next >**.
    - Features (Page 5)
      - Leave everything unchanged and click **Next >**.
    - AD CS (Page 6)
      - Click **Next >**.
    - Role Services (Page 6.1)
      - Leave everything unchanged and click **Next >**.
    - Confirmation (Page 7)
      - Click **Install**.
      - After the installation has finished, click **Close**.

## Configure Active Directory Certificate Services

1.  In the left sidebar of the **Server Manager**, select **AD CS**.
1.  Click the **More...** link in the top right in the yellow alert that reads "Configuration required for Active Directory Certificate Services at ADFSTEST".
1.  Click the **Configure Active Directory Certificate Services on the destination server** link in the "Action" column.
1.  Proceed through the "AD CS Configuration" wizard as follows:
    - Credentials (Page 1)
      - Leave everything unchanged and click **Next >**.
    - Role Services (Page 2)
      - Check the box for **Certification Authority**.
      - Click **Next >**.
    - Setup Type (Page 3)
      - Leave everything unchanged (**Enterprise CA** is selected) and click **Next >**.
    - CA Type (Page 4)
      - Leave everything unchanged (**Root CA** is selected) and click **Next >**.
    - Private Key (Page 5)
      - Leave everything unchanged (**Create new private key** is selected) and click **Next >**.
    - Cryptography (Page 5.1)
      - Select the hash algorithm for signing certificates issued by this CA: `SHA256`
      - Leave everything else unchanged.
      - Click **Next >**.
    - CA Name (Page 5.2)
      - Leave everything unchanged and click **Next >**.
    - Validity Period (Page 5.3)
      - Leave everything unchanged and click **Next >**.
    - Certificate Database (Page 6)
      - Leave everything unchanged and click **Next >**.
    - Confirmation (Page 7)
      - Click **Configure**.
    - Results (Page 8)
      - Click **Close**.

## Create a TLS certificate

1.  In the **Server Manager**, open **Tools > Internet Information Services (IIS) Manager**.
1.  In the left sidebar, select **ADFSTEST (SGDEV\alice)**.
1.  Double-click on the **Server Certificates** feature.
1.  In the right sidebar, click the **Create Self-Signed Certificate...** action.
    - Specify a friendly name for this certificate: `adfstest-cert`
    - Select a certificate store for the new certificate: `Personal`
    - Click **OK**.

<!--
TODO: These steps are not necessary.

1. In the right sidebar, click the **Create Domain Certificate...** action.
1. Proceed through the "Create Certificate" wizard as follows:
   - Distinguished Name Properties (Page 1)
     - Common name: `adfstest.sgdev.org`
     - Organization: `Sourcegraph` (any value is OK)
     - Organizational unit: `Test` (any value is OK)
     - City/locality: `San Francisco` (any value is OK)
     - State/province: `CA` (any value is OK)
     - Country/region: `US` (any value is OK)
     - Click **Next**.
   - Online Certification Authority (Page 2)
     - Specify Online Certification Authority:
       1. Click **Select...**.
       1. Choose the sole item (`sgdev-ADFSTEST-CA`)
       1. Click **OK**.
     - Friendly name: `adfstest-cert`
     - Click **Finish**.
-->

## Configure IIS to use HTTPS

1.  In the **Server Manager**, open **Tools > Internet Information Services (IIS) Manager**.
1.  In the left sidebar, select **ADFSTEST (SGDEV\alice) > Sites**.
1.  Select the **Default Web Site** site.
1.  In the right sidebar, click the **Bindings...** action.
1.  In the "Site Bindings" window, click **Add...**.
1.  In the "Add Site Binding" window, provide the following values and then click **OK**.
    - Type: `https`
    - IP address: `All Unassigned`
    - Port: `443`
    - Host name: `adfstest.sgdev.org`
    - Require Server Name Indication: leave unchecked
    - SSL certificate: `adfstest-cert`
1.  **Close** the "Site Bindings" window.

## Add user profile information to the `alice` account

1.  In the **Server Manager**, open **Tools > Active Directory Users and Computers**.
1.  In the left sidebar, select **Active Directory Users and Computers > sgdev.org > Users**.
1.  In the main list, right-click the **alice** entry and select **Properties** in the context menu.
1.  In the "alice Properties" window, provide the following values and click **OK**.
    - First name: `AliceFirstName` (any value is OK)
    - Last name: `AliceastName` (any value is OK)
    - Display name: `AliceDisplayName` (any value is OK)
    - E-mail: `alice@example.com` (any value is OK)

(The only required field is the email address (which Sourcegraph requires for all SAML users). Populating the other fields just helps confirm that other user profile information is passed to Sourcegraph in the SAML authentication flow.)

## Configure ADFS

1.  In the left sidebar of the **Server Manager**, select **AD FS**.
1.  Click the **More...** link in the top right in the yellow alert that reads "Configuration required for Federation Service at ADFSTEST".
1.  Click the **Run the AD FS Management snap-in** link in the "Action" column.
1.  In the "AD FS" window, click the **➡️ AD FS Federation Server Configuration Wizard** link.
1.  Proceed through the "AD FS Federation Server Configuration Wizard" as follows:
    - Welcome (Page 1)
      - Select **Create a new Federation Service** and click **Next >**.
    - Select Deployment Type (Page 2)
      - Select **New federation server farm** and click **Next >**.
    - Federation Service Name (Page 3)
      - Leave everything unchanged (the `adfstest-cert` SSL certificate is selected) and click **Next >**.
    - Specify Service Account (Page 4)
      - Service account
        1.  Click **Browse**.
        1.  In the "Select User" window in the **Enter the object name to select** text box, type `alice` and then click **OK**.
      - Password: use the same password as for `alice`
    - Summary (Page 5)
      - Click **Next >**.
    - Results (Page 5)
      - Wait for completion.
      - Ignore the warning for the "Configure service settings" step that reads "An error occurred during an attempt to modify the directory attributes for the specified service account. ...".
      - Click **Close**.

## Confirm that ADFS SAML metadata accessible

The ADFS SAML service should now be accessible from your computer. Confirm that, and then add the SSL certificate to your computer's certificate trust store.

1.  Confirm that the following command prints a large XML document:

    ```
    curl --insecure https://adfstest.sgdev.org/federationmetadata/2007-06/federationmetadata.xml
    ```

    If it fails, ensure that you configured the Azure VM firewall correctly and added the Azure VM's public IP to `/etc/hosts` (see above).

1.  Extract the `adfstest.sgdev.org` SSL certificate to a file:

    ```
    openssl s_client -connect adfstest.sgdev.org:443 -showcerts < /dev/null | openssl x509 > adfstest-cert.crt
    ```

1.  Add `adfstest-cert.crt` to your computer's certificate trust store.
    - Linux (Debian/Ubuntu): `sudo mv adfstest-cert.crt /usr/share/ca-certificates && sudo dpkg-reconfigure ca-certificates` and select `ask`, then check the box for `adfstest-cert.crt` and continue.
    - macOS: Use Keychain Access.
1.  Confirm that the SSL certificate is now trusted by your system by rerunning the `curl` command without `--insecure`:

    ```
    curl https://adfstest.sgdev.org/federationmetadata/2007-06/federationmetadata.xml
    ```

## Add ADFS as a SAML auth provider to Sourcegraph site config

1.  Make your Sourcegraph instance externally accessible over HTTPS (e.g., using [ngrok](https://ngrok.com)).

    🎗️ From here on, substitute `https://sourcegraph.example.com` with your instance's actual URL.

1.  Add the following to your Sourcegraph site configuration:

    ```json
    {
      ...,
      "externalURL": "https://sourcegraph.example.com",
      "auth.providers": [
        {
          "type": "saml",
          "displayName": "Windows Server 2012 ADFS 2.1 SAML (dev)",
          "identityProviderMetadataURL": "https://adfstest.sgdev.org/federationmetadata/2007-06/federationmetadata.xml"
        }
      ]
    }
    ```

    ⚠ Ensure that this SAML auth provider is the only auth provider with a `type` of `saml`.

1.  Confirm there are no error messages in the Sourcegraph logs (for Sourcegraph Data Center, the `sourcegraph-frontend` pod logs). The most likely error message indicating a problem is `Error prefetching SAML service provider metadata.`.

## Add Sourcegraph as an ADFS Relying Party (RP)

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
1.  In the "Edit Claim Rules for sourcegraph.example.com" window in the "Issuance Transform Rules" tab, add the 2 following rules.
    - _Send User Info rule:_ Click **Add Rule...** and proceed through the "Add Transform Claim Rule Wizard" as follows:
      - Choose Rule Type (Page 1)
        - Claim rule template: `Send LDAP Attributes as Claims`
        - Click **Next >**.
      - Configure Claim Rule (Page 2)
        - Claim rule name: `Send User Info` (any value is OK)
        - Attribute store: `Active Directory`
        - Mapping of LDAP attributes to outgoing claim types:
        - `E-Mail-Addresses` -> `E-Mail Address`
        - `Display-Name` -> `Name`
        - Click **Finish**.
    - _Email to NameID rule:_ Click **Add Rule...** and proceed through the "Add Transform Claim Rule Wizard" as follows:
      - Choose Rule Type (Page 1)
        - Claim rule template: `Transform an Incoming Claim`
        - Click **Next >**.
      - Configure Claim Rule (Page 2)
        - Claim rule name: `Email to NameID`
        - Incoming claim type: `E-Mail Address`
        - Outgoing claim type: `Name ID`
        - Outgoing name ID format: `Persistent identifier`
        - Select **Pass through all claim values**.
        - Click **Finish**.
1.  Click **OK** to apply the new claim rules and close the window.

## Authenticate to Sourcegraph using ADFS

All configuration is now complete. Let's test that it works.

1.  Visit `https://sourcegraph.example.com`. (If you are already authenticated from before configuring the SAML auth provider, sign out of Sourcegraph.)
1.  When prompted to sign into ADFS, provide the `alice` credentials and continue.
1.  Confirm that you are authenticated to Sourcegraph as `alice`.

✅ Done!
