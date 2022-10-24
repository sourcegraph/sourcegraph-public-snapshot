import React from 'react';

import AccountCircleIcon from 'mdi-react/AccountCircleIcon';
import GithubIcon from 'mdi-react/GithubIcon';
import GitLabIcon from 'mdi-react/GitlabIcon';

import {AuthProvider} from '../../jscontext';

export type ExternalAccountKind = Exclude<AuthProvider['serviceType'], 'http-header' | 'builtin'>

export interface ExternalAccount {
    kind: ExternalAccountKind

    /**
     * Title to show in the external account "button"
     */
    title: string

    /**
     * Icon to show in the external account "button"
     */
    icon: React.ComponentType<{ className?: string }>
}

const GITHUB: ExternalAccount = {
    kind: 'github',
    title: 'GitHub',
    icon: GithubIcon,
}

const GITLAB: ExternalAccount = {
    kind: 'gitlab',
    title: 'GitLab',
    icon: GitLabIcon,
}

const OPENID_CONNECT: ExternalAccount = {
    kind: 'openidconnect',
    title: 'OpenID Connect',
    icon: AccountCircleIcon,
}

const SAML: ExternalAccount = {
    kind: 'saml',
    title: 'SAML',
    icon: AccountCircleIcon,
}

export const defaultExternalAccounts: Record<ExternalAccountKind, ExternalAccount> = {
    ['github']: GITHUB,
    ['gitlab']: GITLAB,
    ['saml']: SAML,
    ['openidconnect']:OPENID_CONNECT,
}
