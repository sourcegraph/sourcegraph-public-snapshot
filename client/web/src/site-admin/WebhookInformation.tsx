import type { FC } from 'react'

import classNames from 'classnames'

import { CopyableText } from '../components/CopyableText'
import type { WebhookFields } from '../graphql-operations'

import styles from './WebhookInformation.module.scss'

export interface WebhookInformationProps {
    webhook: WebhookFields
}

export const WebhookInformation: FC<WebhookInformationProps> = props => {
    const { webhook } = props

    return (
        <table className={classNames(styles.table, 'table')}>
            <tbody>
                <tr>
                    <th className={styles.tableHeader}>Code host</th>
                    <td>{webhook.codeHostKind}</td>
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
                        <CopyableText text={webhook.secret ?? ''} secret={true} />
                    </td>
                </tr>
            </tbody>
        </table>
    )
}
