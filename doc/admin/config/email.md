# Configure email sending / SMTP server

Sourcegraph uses an SMTP server of your choosing to send emails for:

* [Code Monitoring](../../code_monitoring/index.md) notifications
* Password reset requests
* Email verification when built-in authentication is enabled
* Inviting other users to Sourcegraph itself, or to an organization/team on Sourcegraph

## Configuring Sourcegraph to send email via Amazon AWS / SES

To use Amazon SES with Sourcegraph, first [follow these steps to create an SES account for Sourcegraph](https://docs.aws.amazon.com/ses/latest/dg/send-email-smtp-software-package.html).

Navigate to your site configuration (e.g. https://sourcegraph.com/site-admin/configuration) and fill in the configuration:

```json
  "email.address": "from@domain.com",
  "email.smtp": {
    "authentication": "PLAIN",
    "username": "<SES SMTP username>",
    "password": "<SES SMTP password>",
    "host": "email-smtp.us-west-2.amazonaws.com",
    "port": 587
  },
```

Please note that the configured `email.address` (the from address) must be a verified address with SES, see [this page for details](https://docs.aws.amazon.com/ses/latest/dg/verify-addresses-and-domains.html).

[Send a test email](#sending-a-test-email) to verify it is configured properly.

## Configuring Sourcegraph to send email via Google Workspace / GMail

To use Google Workspace with Sourcegraph, you will need to [create an SMTP Relay account](https://support.google.com/a/answer/2956491). Be sure to choose `Require SMTP Authentication` and `Require TLS encryption` in step 7.

Navigate to your site configuration (e.g. https://sourcegraph.com/site-admin/configuration) and fill in the configuration:

```json
  "email.address": "test@domain.com",
  "email.smtp": {
    "authentication": "PLAIN",
    "username": "test@domain.com",
    "password": "<YOUR SECRET>",
    "host": "smtp-relay.gmail.com",
    "port": 587
  },
```

Make sure that `test@domain.com` in both places of the configuration matches the email address of the account you created, and that `<YOUR SECRET>` is replaced with the account password.

[Send a test email](#sending-a-test-email) to verify it is configured properly.

## Configuring Sourcegraph to send email using another provider

Other providers such as Mailchimp and Sendgrid may also be used with Sourcegraph. Any valid SMTP server account should do. For those two providers in specific, you may follow their documentation:

* https://mailchimp.com/developer/transactional/docs/smtp-integration/
* https://docs.sendgrid.com/for-developers/sending-email/getting-started-smtp

Once you have an SMTP account, simply navigate to your site configuration (e.g. https://sourcegraph.com/site-admin/configuration) and fill in the configuration:

```json
  "email.address": "from@domain.com",
  "email.smtp": {
    "authentication": "PLAIN",
    "username": "test@domain.com",
    "password": "<YOUR SECRET>",
    "host": "smtp-server.example.com",
    "port": 587
  },
```

A few helpful tips:

* `email.address` is the email Sourcegraph will send email `FROM`. Make sure your account has privilege to send mail from that address.
* You should use a TLS/SSL enabled port, either `587` or `2525`. Any port number is allowed, though.
* The following `authentication` types are allowed: `PLAIN` (default), `CRAM-MD5`, and `none`
* If a `HELO` domain is required, simply add `"domain": "<the domain>",` under `email.smtp`
* TLS certificate verification can be disabled if your SMTP server does not have a valid TLS cert. To do so, set `"noVerifyTLS": true,` (this is not recommended and may be a security risk)

[Send a test email](#sending-a-test-email) to verify it is configured properly.

## Sending a test email

(Added in Sourcegraph v3.38)

To verify email sending is working correctly, visit the GraphQL API console at e.g. https://sourcegraph.example.com/api/console and then run the following query replacing `test@example.com` with your personal email address:

```graphql
mutation {
  sendTestEmail(to: "test@example.com")
}
```

If everything is successfully configured, you should see a response like:

```json
{
  "data": {
    "sendTestEmail": "Sent test email to \"test@example.com\" successfully! Please check it was received."
  }
}
```

Otherwise, you should see an error with more information:

```json
{
  "data": {
    "sendTestEmail": "Failed to send test email: mail: ..."
  }
}
```

If you need further assistance, please let us know at <mailto:support@sourcegraph.com>.
