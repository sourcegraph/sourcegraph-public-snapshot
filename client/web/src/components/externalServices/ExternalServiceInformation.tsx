import React, { FC } from 'react'

import classNames from 'classnames'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { Icon, Link, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import styles from '../../site-admin/WebhookInformation.module.scss'

interface ExternalServiceInformationProps {
    /**
     * Icon to show in the external service "button".
     */
    icon: React.ComponentType<React.PropsWithChildren<{ className?: string }>>
    kind: ExternalServiceKind
    displayName: string
    codeHostID: string
    reposNumber: number
    syncInProgress: boolean
}

export const ExternalServiceInformation: FC<ExternalServiceInformationProps> = props => {
    const { icon, kind, displayName, codeHostID, reposNumber, syncInProgress } = props

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
            </tbody>
        </table>
    )
}
