# Management console

The management console is used to edit critical site configuration. This configuration is separate
from the regular site configuration, because an error here can make Sourcegraph completely
inaccessible (except through the management console). Critical configuration includes things like
authentication providers, the application URL, and the license key.

The management console is served on a separate port (2633) than the main application. This port is
not exposed to public traffic by default and can only be accessed by admins with access to the
deployment environment.

## Authentication

The management console does not use the authentication mechanism of the main application (because an
error in critical site configuration may make sign-in impossible). Site administrators authenticate
to the management console (after the port is exposed) via HTTP basic authentication. Admins should
obtain the password after initial setup by going to `/site-admin` and save this password to a safe
location.
