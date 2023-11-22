import React, { useCallback } from 'react'

import { mdiAlertCircle } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Button, Icon } from '@sourcegraph/wildcard'

import type { WebhookByIDLogPageHeaderResult } from '../graphql-operations'

import { WEBHOOK_BY_ID_LOG_PAGE_HEADER } from './webhooks/backend'
import { PerformanceGauge } from './webhooks/PerformanceGauge'

import styles from './WebhookInfoLogPageHeader.module.scss'

export interface Props {
    webhookID: string
    onlyErrors: boolean

    onSetOnlyErrors: (onlyErrors: boolean) => void
}

export const WebhookInfoLogPageHeader: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    webhookID,
    onlyErrors,
    onSetOnlyErrors: onSetErrors,
}) => {
    const onErrorToggle = useCallback(() => onSetErrors(!onlyErrors), [onlyErrors, onSetErrors])

    const { data } = useQuery<WebhookByIDLogPageHeaderResult>(WEBHOOK_BY_ID_LOG_PAGE_HEADER, {
        variables: { webhookID },
    })
    const errorCount = data?.webhookLogs.totalCount ?? 0

    return (
        <div className={styles.grid}>
            <div className={styles.errors}>
                <PerformanceGauge
                    count={errorCount}
                    countClassName={errorCount > 0 ? 'text-danger' : undefined}
                    label="recent error"
                />
            </div>
            <div className={styles.button}>
                <Button variant="danger" onClick={onErrorToggle} outline={!onlyErrors}>
                    <Icon
                        className={classNames(styles.icon, onlyErrors && styles.enabled)}
                        aria-hidden={true}
                        svgPath={mdiAlertCircle}
                    />
                    <span className="ml-1">Show errors</span>
                </Button>
            </div>
        </div>
    )
}
