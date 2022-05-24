import React from 'react'

import { Link, Typography } from '@sourcegraph/wildcard'

import { ExternalServiceKind } from '../../../graphql-operations'

import styles from './modalHints.module.scss'

const MachineUserRecommendation = (
    <p>
        We recommend setting up a machine user to provide restricted access to repositories.{' '}
        <Link to="https://docs.sourcegraph.com/cloud/access_tokens_on_cloud" target="_blank" rel="noopener noreferrer">
            Learn more
        </Link>
        .
    </p>
)

export const scopes: Partial<Record<ExternalServiceKind, React.ReactFragment>> = {
    [ExternalServiceKind.GITHUB]: (
        <small className="text-muted">
            Use an access token with <Typography.Code className={styles.codeInline}>repo</Typography.Code>,{' '}
            <Typography.Code className={styles.codeInline}>read:org</Typography.Code>, and{' '}
            <Typography.Code className={styles.codeInline}>user:email</Typography.Code> scopes.
            {MachineUserRecommendation}
        </small>
    ),
    [ExternalServiceKind.GITLAB]: (
        <small className="text-muted">
            Use an access token with <Typography.Code className={styles.codeInline}>read_user</Typography.Code>,{' '}
            <Typography.Code className={styles.codeInline}>read_api</Typography.Code> and{' '}
            <Typography.Code className={styles.codeInline}>read_repository</Typography.Code> scopes.
            {MachineUserRecommendation}
        </small>
    ),
}

export const getMachineUserFragment = (serviceName: string): React.ReactFragment => (
    <div className={styles.alertBodyBg + ' p-2 border border-2 rounded'}>
        <div className="px-2 py-1">
            <Typography.H4>
                We recommend setting up a machine user on {serviceName} to provide restricted access to repositories.{' '}
                <Link
                    to="https://docs.sourcegraph.com/cloud/access_tokens_on_cloud"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Learn more
                </Link>
                .
            </Typography.H4>

            <span className="text-muted">
                Using your own personal access token may reveal your public and private repositories to other members of
                your organization.
            </span>
        </div>
    </div>
)
