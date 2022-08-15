import React from 'react'

import { Link, Code, H4, Text } from '@sourcegraph/wildcard'

import { ExternalServiceKind } from '../../../graphql-operations'

import styles from './modalHints.module.scss'

const MachineUserRecommendation = (
    <Text>
        We recommend setting up a machine user to provide restricted access to repositories.{' '}
        <Link to="https://docs.sourcegraph.com/cloud/access_tokens_on_cloud" target="_blank" rel="noopener noreferrer">
            Learn more
        </Link>
        .
    </Text>
)

export const scopes: Partial<Record<ExternalServiceKind, React.ReactNode>> = {
    [ExternalServiceKind.GITHUB]: (
        <small className="text-muted">
            Use an access token with <Code className={styles.codeInline}>repo</Code>,{' '}
            <Code className={styles.codeInline}>read:org</Code>, and{' '}
            <Code className={styles.codeInline}>user:email</Code> scopes.
            {MachineUserRecommendation}
        </small>
    ),
    [ExternalServiceKind.GITLAB]: (
        <small className="text-muted">
            Use an access token with <Code className={styles.codeInline}>read_user</Code>,{' '}
            <Code className={styles.codeInline}>read_api</Code> and{' '}
            <Code className={styles.codeInline}>read_repository</Code> scopes.
            {MachineUserRecommendation}
        </small>
    ),
}

export const getMachineUserFragment = (serviceName: string): React.ReactNode => (
    <div className={styles.alertBodyBg + ' p-2 border border-2 rounded'}>
        <div className="px-2 py-1">
            <H4>
                We recommend setting up a machine user on {serviceName} to provide restricted access to repositories.{' '}
                <Link
                    to="https://docs.sourcegraph.com/cloud/access_tokens_on_cloud"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Learn more
                </Link>
                .
            </H4>

            <span className="text-muted">
                Using your own personal access token may reveal your public and private repositories to other members of
                your organization.
            </span>
        </div>
    </div>
)
