import React, { type FC } from 'react'

import classNames from 'classnames'

import type { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { Icon, Link, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import type { RateLimiterState } from './backend'

import styles from '../../site-admin/WebhookInformation.module.scss'

interface ExternalServiceInformationProps {
    /**
     * Icon to show in the external service "button".
     */
    icon: React.ComponentType<React.PropsWithChildren<{ className?: string }>>
    kind: ExternalServiceKind
    displayName: string
    rateLimiterState?: RateLimiterState | null
    codeHostID: string
    reposNumber: number
    syncInProgress: boolean
    gitHubApp?: {
        id: string
        name: string
    } | null
}

export const RateLimiterStateInfo: FC<{ rateLimiterState: RateLimiterState }> = props => {
    const { rateLimiterState } = props
    const rateLimiterDebug = Object.entries(rateLimiterState).map(([key, value]) => (
        <div key={key}>
            {key}: {value.toString()}
        </div>
    ))

    return (
        <tr>
            <th className={styles.tableHeader}>Rate limit</th>
            {rateLimiterState.infinite ? (
                <td>
                    <Tooltip content={rateLimiterDebug}>
                        <span>No rate limit</span>
                    </Tooltip>
                </td>
            ) : (
                <td>
                    <Tooltip content={rateLimiterDebug}>
                        <span>
                            {(rateLimiterState.limit / rateLimiterState.interval).toFixed(2)} requests per second
                        </span>
                    </Tooltip>
                </td>
            )}
        </tr>
    )
}

export const ExternalServiceInformation: FC<ExternalServiceInformationProps> = props => {
    const { icon, kind, displayName, codeHostID, reposNumber, syncInProgress, gitHubApp, rateLimiterState } = props

    return (
        <table className={classNames(styles.table, 'table')}>
            <tbody>
                <tr>
                    <th className={styles.tableHeader}>Code host kind</th>
                    <td>
                        <Icon inline={true} as={icon} aria-label="Code host logo" className="mr-2" />
                        {kind}
                    </td>
                </tr>
                {gitHubApp && (
                    <tr>
                        <th className={styles.tableHeader}>GitHub App</th>
                        <td>
                            <Link to={`/site-admin/github-apps/${encodeURIComponent(gitHubApp.id)}`}>
                                {gitHubApp.name}
                            </Link>
                        </td>
                    </tr>
                )}
                <tr>
                    <th className={styles.tableHeader}>Display name</th>
                    <td>{displayName}</td>
                </tr>
                <tr>
                    <th className={styles.tableHeader}>Repositories</th>
                    <td>
                        <Tooltip content="Click to see the list of repositories">
                            <Link to={`/site-admin/repositories?codeHost=${encodeURIComponent(codeHostID)}`}>
                                {reposNumber}
                            </Link>
                        </Tooltip>
                        {syncInProgress && (
                            <span className="text-muted font-italic ml-2">
                                Syncing list of repositories from code host...
                                <LoadingSpinner inline={true} />
                            </span>
                        )}
                    </td>
                </tr>
                {rateLimiterState && <RateLimiterStateInfo rateLimiterState={rateLimiterState} />}
            </tbody>
        </table>
    )
}
