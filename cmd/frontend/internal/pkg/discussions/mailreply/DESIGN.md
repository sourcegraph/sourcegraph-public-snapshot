## General behavior

Any reply is accepted, as long as it has a text form. We treat the text as Markdown. With email clients such as Gmail, things such as e.g. bulleted lists, bold, etc. buttons work OK because they produce this format of text already.

This feature _is optional_, as it requires giving Sourcegraph access to an IMAP server with support for sub-addressing (e.g. `foo+bar@me.com`, see https://tools.ietf.org/html/rfc5233). It is activated when `email.imap` is configured.

## Authentication model

It is easy to know that the person we're sending emails to is correct (because they have verified their email) but it is very hard to know whether or not an email coming from someone is actually that person. For example, it is trivial to spoof an email `From` header. On top of this, existing email authentication methods which may prevent this are quite beastly (lookup PGP, DKIM, SPF, etc).

The authentication method I chose here is simple and based on my reverse engineering of GitHub's model:

- When we send an email notification to you, the `Reply-To` header includes a random token. For example, `notifications+SOMESECRET@sourcegraph.com`.
- The token grants anyone with it access to post to _that thread_ as _that user_, indefinitely.
- If you reply to the email, the frontend worker service reads it. Only emails containing a token that we previously generated (and stored in postgres) are accepted.

Possible attack vectors include:

- **Someone gets one of your tokens**
  - The easiest way this can happen is by you forwarding an email notification to someone. This allows them to post as you in the discussion thread. GitHub's email notifications have the same flaw (just respond to the `reply+TOKEN@reply.github.com` address). Otherwise, it cannot happen generally unless someone has access to your email (which implicitly means they have access to your Sourcegraph account anyway due to password reset).
  - The token is only good for that one thread, it doesn't allow e.g. posting to other threads or performing other actions under your account.
- **Someone guesses one of your tokens**
  - The token is a SHA256 produced from 128 bytes of crypto/rand data. Should be basically impossible to guess or brute force.

## Comparison with other services

As mentioned above, I wrote this primarily following what information I could glean from GitHub's model. However, there is little to no information online about how to write a system such as this.

As Keegan pointed out to me, Phabricator uses an almost identical model for this (see https://secure.phabricator.com/book/phabricator/article/configuring_inbound_email/). My comparison of our implementation here versus theirs is:

1.  Our system is reply-only, theirs allows for much more actions such as creating new bugs by sending an email to an address. But it is not clear to me how (or if they do) secure this, or if it is even valuable in general.
2.  They also acknowledge the risk that leaking emails could allow others to act on a user's behalf. For this reason, they do not allow some dangerous actions such as e.g. accepting a revision into the codebase via email. We will need to keep this in mind and generally restrict what operations can be done via email (for example, we should keep this in mind if we ever have code discussions hooks).
3.  Their security model is nearly identical to ours. They use the same model that we do (reply-to token provides access to a single object as an arbitrary user). They also came to the same conclusion around: _"Phabricator does not currently attempt to verify "From" addresses because this is technically complex, seems unreasonably difficult in the general case [...]"_.
4.  They support many more email providers: Mailgun, Postmark, Sendgrid, and Local MTA (but discouraged). I think most organizations have an IMAP server, and it spares us a lot of work to have to support these other providers for today, so I only support IMAP right now. We only use a single inbox, so we are compatible with e.g. a standard company Gmail / Google Apps setup.
