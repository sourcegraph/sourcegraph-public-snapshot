export const content = [
    {
        id: 'licenseKey',
        isComplete: false,
        title: 'Set license key',
        description: 'Please set your license key',
        link: 'site-admin/configuration?actionItem=setLicenseKey',
    },
    {
        id: 'externalURL',
        isComplete: false,
        title: 'Set external URL',
        description: 'Must be set in order for Sourcegraph to work correctly.',
        link: 'site-admin/configuration?actionItem=setExternalURL',
    },
    {
        id: 'emailSmtp',
        isComplete: false,
        title: 'Set up SMTP',
        description: 'Must be set in order for Sourcegraph to send emails.',
        link: 'https://docs.sourcegraph.com/admin/config/email',
    },
    {
        id: 'externalServices',
        isComplete: false,
        title: 'Connect a code host',
        description: 'You must connect a code host to set up user authentication and use Sourcegraph.',
        link: 'site-admin/external-services',
    },
    {
        id: 'authProviders',
        isComplete: false,
        title: 'Set up user authentication',
        description: 'We recommend that enterprise instances use SSO or SAML to authenticate users.',
        link: 'https://docs.sourcegraph.com/admin/config/authorization_and_authentication',
    },
]
