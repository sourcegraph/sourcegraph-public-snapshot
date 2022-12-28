import React from 'react'

import AccountCircleIcon from 'mdi-react/AccountCircleIcon'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitLabIcon from 'mdi-react/GitlabIcon'

import { AuthProvider } from '../../jscontext'

export type ExternalAccountKind = Exclude<
    AuthProvider['serviceType'],
    'http-header' | 'builtin' | 'sourcegraph-operator'
>

export interface ExternalAccount {
    /**
     * Title to show in the external account "button"
     */
    title: string

    /**
     * Icon to show in the external account "button"
     */
    icon: React.ComponentType<{ className?: string }>
}

export const defaultExternalAccounts: Record<ExternalAccountKind, ExternalAccount> = {
    github: {
        title: 'GitHub',
        icon: GithubIcon,
    },
    gitlab: {
        title: 'GitLab',
        icon: GitLabIcon,
    },
    openidconnect: {
        title: 'OpenID Connect',
        icon: AccountCircleIcon,
    },
    saml: {
        title: 'SAML',
        icon: AccountCircleIcon,
    },
    bitbucketCloud: {
        title: 'Bitbucket Cloud',
        icon: BitbucketIcon,
    },
}
