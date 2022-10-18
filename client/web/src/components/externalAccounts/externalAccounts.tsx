import React from 'react';

import AlertIcon from 'mdi-react/AlertIcon';
import GithubIcon from 'mdi-react/GithubIcon';
import GitLabIcon from 'mdi-react/GitlabIcon';

import { ExternalAccountKind } from '@sourcegraph/shared/src/graphql-operations';

export interface AddExternalAccountOptions {
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

const GITHUB_DOTCOM: AddExternalAccountOptions = {
    kind: ExternalAccountKind.GITHUB,
    title: 'GitHub',
    icon: GithubIcon,
}

const GITLAB_DOTCOM: AddExternalAccountOptions = {
    kind: ExternalAccountKind.GITLAB,
    title: 'GitLab',
    icon: GitLabIcon,
}

const SAML: AddExternalAccountOptions = {
    kind: ExternalAccountKind.SAML,
    title: 'SAML',
    icon: AlertIcon,
}

export const defaultExternalAccounts: Record<ExternalAccountKind, AddExternalAccountOptions> = {
    [ExternalAccountKind.GITHUB]: GITHUB_DOTCOM,
    [ExternalAccountKind.GITLAB]: GITLAB_DOTCOM,
    [ExternalAccountKind.SAML]: SAML,
}

