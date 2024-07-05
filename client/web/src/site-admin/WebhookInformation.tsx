import type { FC } from 'react'

import classNames from 'classnames'

import { Icon } from '@sourcegraph/wildcard'

import { CopyableText } from '../components/CopyableText'
import { defaultExternalServices } from '../components/externalServices/externalServices'
import type { WebhookFields } from '../graphql-operations'

import styles from './WebhookInformation.module.scss'

export interface WebhookInformationProps {
    webhook: WebhookFields
}

export const WebhookInformation: FC<WebhookInformationProps> = props => {
    const { webhook } = props

    const IconComponent = defaultExternalServices[webhook.codeHostKind].icon

    const codeHostKindName = defaultExternalServices[webhook.codeHostKind].defaultDisplayName

    return (
        <table className={classNames(styles.table, 'table')}>
            <tbody>
                <tr>
                    <th className={styles.tableHeader}>Code host</th>
                    <td>
                        <Icon inline={true} as={IconComponent} aria-label="Code host logo" className="mr-1" />
                        {codeHostKindName}
                    </td>
                </tr>
                <tr>
                    <th className={styles.tableHeader}>URN</th>
                    <td>{webhook.codeHostURN}</td>
                </tr>
                <tr>
                    <th className={styles.tableHeader}>Webhook endpoint</th>
                    <td className={styles.contentCell}>
                        <CopyableText text={webhook.url} size={60} />
                    </td>
                </tr>
                <tr>
                    <th className={styles.tableHeader}>Secret</th>
                    <td className={styles.contentCell}>
                        {webhook.secret === null ? (
                            <span className="text-muted">
                                <em>No secret</em>
                            </span>
                        ) : (
                            <CopyableText text={webhook.secret} secret={true} />
                        )}
                    </td>
                </tr>
            </tbody>
        </table>
    )
}
