import { FC, useState } from 'react'

import { mdiEye } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon, Input } from '@sourcegraph/wildcard'

import { CopyableText } from '../components/CopyableText'
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
                        <CopyableText text={webhook.url} size={60} />
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
