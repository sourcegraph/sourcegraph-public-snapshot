import { FC, useState } from 'react'

import { mdiContentCopy, mdiEye } from '@mdi/js'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'

import { Button, Icon, Input } from '@sourcegraph/wildcard'

import { WebhookFields } from '../graphql-operations'

import styles from './WebhookInformation.module.scss'

export interface WebhookInformationProps {
    webhook: WebhookFields
}

export const WebhookInformation: FC<WebhookInformationProps> = props => {
    const { webhook } = props
    const [secretShown, setSecretShown] = useState(false)

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
                        {webhook.url}
                        <Button size="sm" className={styles.copyButton} onClick={() => copy(webhook.url)}>
                            <Icon svgPath={mdiContentCopy} inline={true} aria-label="Copy webhook endpoint" />
                        </Button>
                    </td>
                </tr>
                <tr>
                    <th className={styles.tableHeader}>Secret</th>
                    <td className={styles.contentCell}>
                        <Input
                            name="secret"
                            type={secretShown ? 'text' : 'password'}
                            value={secretShown ? webhook.secret ?? '' : 'verysecretvalue'}
                            className={classNames(styles.input)}
                            inputClassName={classNames(styles.input)}
                            readOnly={true}
                        />
                        <Button onClick={() => setSecretShown(!secretShown)}>
                            <Icon svgPath={mdiEye} inline={true} aria-label="ToggleSecret" />
                        </Button>
                    </td>
                </tr>
            </tbody>
        </table>
    )
}
