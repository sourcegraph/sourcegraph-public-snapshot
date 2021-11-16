import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { ExternalServiceKind } from '../../../graphql-operations'

import styles from './modalHints.module.scss'

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
                with <code className={styles.codeInline}>repo</code>,{' '}
                <code className={styles.codeInline}>read:org</code> and{' '}
                <code className={styles.codeInline}>user:email</code> scopes.
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
                with <code className={styles.codeInline}>read_user</code>,{' '}
                <code className={styles.codeInline}>read_api</code> and{' '}
                <code className={styles.codeInline}>read_repository</code> scopes.
            </span>
        </small>
    ),
}
