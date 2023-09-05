export const content = [
    {
        id: 'licenseKey',
        isComplete: true,
        title: 'Set license key',
        description: 'Please set your license key',
        link: '#',
    },
    {
        id: 'externalURL',
        isComplete: false,
        title: 'Set external URL',
        description: 'Must be set in order for Sourcegraph to work correctly.',
        link: '#',
    },
    {
        id: 'emailSmtp',
        isComplete: false,
        title: 'Set up SMTP',
        description: 'Must be set in order for Sourcegraph to send emails.',
        link: '#',
    },
    {
        id: 'externalServices',
        isComplete: false,
        title: 'Connect a code host',
        description: 'You must connect a code host to set up user authentication and use Sourcegraph.',
        link: '#',
    },
    {
        id: 'authProviders',
        isComplete: false,
        title: 'Set up user authentication',
        description: 'We recommend that enterprise instances use SSO or SAML to authenticate users.',
        link: '#',
    },
    {
        id: 'usersPermissions',
        isComplete: false,
        title: 'Set user permissions',
        description:
            'We recommend limiting permissions based on repository permissions already set in your code host(s).',
        link: '#',
    },
]
