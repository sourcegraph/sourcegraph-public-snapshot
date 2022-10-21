import React from 'react';

import AccountCircleIcon from 'mdi-react/AccountCircleIcon';
import GithubIcon from 'mdi-react/GithubIcon';
import GitLabIcon from 'mdi-react/GitlabIcon';

import { ExternalAccountKind } from '@sourcegraph/shared/src/graphql-operations';

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

const GITHUB_DOTCOM: ExternalAccount = {
    kind: ExternalAccountKind.GITHUB,
    title: 'GitHub',
    icon: GithubIcon,
}

const GITLAB_DOTCOM: ExternalAccount = {
    kind: ExternalAccountKind.GITLAB,
    title: 'GitLab',
    icon: GitLabIcon,
}

const SAML: ExternalAccount = {
    kind: ExternalAccountKind.SAML,
    title: 'SAML',
    icon: AccountCircleIcon,
}

export const defaultExternalAccounts: Record<ExternalAccountKind, ExternalAccount> = {
    [ExternalAccountKind.GITHUB]: GITHUB_DOTCOM,
    [ExternalAccountKind.GITLAB]: GITLAB_DOTCOM,
    [ExternalAccountKind.SAML]: SAML,
}

