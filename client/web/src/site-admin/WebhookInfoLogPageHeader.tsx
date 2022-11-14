import React, { useCallback } from 'react'

import { mdiAlertCircle } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Button, Icon } from '@sourcegraph/wildcard'

import { WebhookLogPageHeaderResult } from '../graphql-operations'

import { WEBHOOK_LOG_PAGE_HEADER } from './webhooks/backend'
import { PerformanceGauge } from './webhooks/PerformanceGauge'

import styles from './WebhookInfoLogPageHeader.module.scss'

export interface Props {
    onlyErrors: boolean

    onSetOnlyErrors: (onlyErrors: boolean) => void
}

export const WebhookInfoLogPageHeader: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    onlyErrors,
    onSetOnlyErrors: onSetErrors,
}) => {
    const onErrorToggle = useCallback(() => onSetErrors(!onlyErrors), [onlyErrors, onSetErrors])

    const { data } = useQuery<WebhookLogPageHeaderResult>(WEBHOOK_LOG_PAGE_HEADER, {})
    const errorCount = data?.webhookLogs.totalCount ?? 0

    return (
        <div className={styles.grid}>
            <div className={styles.errors}>
                <PerformanceGauge
                    count={data?.webhookLogs.totalCount}
                    countClassName={errorCount > 0 ? 'text-danger' : undefined}
                    label="recent error"
                />
            </div>
            <div className={styles.services}>
                <PerformanceGauge count={data?.externalServices.totalCount} label="external service" />
            </div>
            <div className={styles.buttons}>
                <Button variant="danger" onClick={onErrorToggle} outline={!onlyErrors}>
                    <Icon
                        className={classNames(styles.icon, onlyErrors && styles.enabled)}
                        aria-hidden={true}
                        svgPath={mdiAlertCircle}
                    />
                    <span className="ml-1">Show errors</span>
                </Button>
                <Button variant="success" className="ml-2">
                    <span className="ml-1">Send test event</span>
                </Button>
            </div>
        </div>
    )
}
