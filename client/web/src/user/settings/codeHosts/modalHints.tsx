import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { ExternalServiceKind } from '../../../graphql-operations'

export const hints: Partial<Record<ExternalServiceKind, React.ReactFragment>> = {
    [ExternalServiceKind.GITHUB]: (
        <small>
            <Link
                to="https://github.com/settings/tokens/new?description=Sourcegraph.com&scopes=user:email,repo,read:org"
                target="_blank"
                rel="noopener noreferrer"
            >
                Create a new access token
            </Link>
            <span className="text-muted">
                {' '}
                with <code className="user-code-hosts-page__code--inline">repo</code>,{' '}
                <code className="user-code-hosts-page__code--inline">read:org</code> and{' '}
                <code className="user-code-hosts-page__code--inline">user:email</code> scopes.
            </span>
        </small>
    ),
    [ExternalServiceKind.GITLAB]: (
        <small>
            <Link to="https://gitlab.com/-/profile/personal_access_tokens" target="_blank" rel="noopener noreferrer">
                Create a new access token
            </Link>
            <span className="text-muted">
                {' '}
                with <code className="user-code-hosts-page__code--inline">read_user</code>,{' '}
                <code className="user-code-hosts-page__code--inline">read_api</code> and{' '}
                <code className="user-code-hosts-page__code--inline">read_repository</code> scopes.
            </span>
        </small>
    ),
}
