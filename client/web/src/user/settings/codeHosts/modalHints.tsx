import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { ExternalServiceKind } from '../../../graphql-operations'

export const hints: Partial<Record<ExternalServiceKind, React.ReactFragment>> = {
    [ExternalServiceKind.GITHUB]: (
        <small>
            <Link to="https://github.com/settings/tokens/new?scopes=repo" target="_blank" rel="noopener noreferrer">
                Create a new personal access token
            </Link>
            <span className="text-muted">
                {' '}
                on <b>GitHub.com</b> with <code className="user-code-hosts-page__code--inline">repo</code> scope.
            </span>
        </small>
    ),
    [ExternalServiceKind.GITLAB]: (
        <small>
            <Link to="https://gitlab.com/-/profile/personal_access_tokens" target="_blank" rel="noopener noreferrer">
                Create a new personal access token
            </Link>
            <span className="text-muted">
                {' '}
                on <b>GitLab.com</b> with <code className="user-code-hosts-page__code--inline">read_user</code>,{' '}
                <code className="user-code-hosts-page__code--inline">read_api</code>, and{' '}
                <code className="user-code-hosts-page__code--inline">read_repository</code> scope.
            </span>
        </small>
    ),
}
