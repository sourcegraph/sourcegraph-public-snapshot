import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { ExternalServiceKind } from '../../../graphql-operations'

import styles from './modalHints.module.scss'

const MachineUserRecommendation = (
    <p>
        We recommend setting up a machine user to provide restricted access to repositories.{' '}
        <Link to="https://docs.sourcegraph.com/cloud/TBD" target="_blank" rel="noopener noreferrer">
            Learn more
        </Link>
        .
    </p>
)

export const scopes: Partial<Record<ExternalServiceKind, React.ReactFragment>> = {
    [ExternalServiceKind.GITHUB]: (
        <small>
            Use an access token
            <span className="text-muted">
                {' '}
                with <code className={styles.codeInline}>repo</code>,{' '}
                <code className={styles.codeInline}>read:org</code>, and{' '}
                <code className={styles.codeInline}>user:email</code> scopes.
            </span>{' '}
            {MachineUserRecommendation}
        </small>
    ),
    [ExternalServiceKind.GITLAB]: (
        <small>
            Use an access token
            <span className="text-muted">
                {' '}
                with <code className={styles.codeInline}>read_user</code>,{' '}
                <code className={styles.codeInline}>read_api</code> and{' '}
                <code className={styles.codeInline}>read_repository</code> scopes.
            </span>{' '}
            {MachineUserRecommendation}
        </small>
    ),
}

export const getMachineUserFragment = (serviceName: string): React.ReactFragment => (
    <div className="p-2 bg-light border border-2 rounded">
        <div className="px-2 py-1">
            <h4>
                We recommend setting up a machine user on {serviceName} to provide restricted access to repositories.
                <Link to="https://docs.sourcegraph.com/cloud/TBD" target="_blank" rel="noopener noreferrer">
                    {' '}
                    Learn more
                </Link>
                .
            </h4>
            Using your own personal access token may reveal your public and private repositories to other members of
            your organization.
        </div>
    </div>
)
