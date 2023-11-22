import React, { useCallback } from 'react'

import { mdiAlertCircle } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Button, Select, Icon } from '@sourcegraph/wildcard'

import type { WebhookLogPageHeaderResult } from '../../graphql-operations'

import { type SelectedExternalService, WEBHOOK_LOG_PAGE_HEADER } from './backend'
import { PerformanceGauge } from './PerformanceGauge'

import styles from './WebhookLogPageHeader.module.scss'

export interface Props {
    externalService: SelectedExternalService
    onlyErrors: boolean

    onSelectExternalService: (externalService: SelectedExternalService) => void
    onSetOnlyErrors: (onlyErrors: boolean) => void
}

export const WebhookLogPageHeader: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    externalService,
    onlyErrors,
    onSelectExternalService: onExternalServiceSelected,
    onSetOnlyErrors: onSetErrors,
}) => {
    const onErrorToggle = useCallback(() => onSetErrors(!onlyErrors), [onlyErrors, onSetErrors])
    const onSelect = useCallback(
        (value: string) => {
            onExternalServiceSelected(value)
        },
        [onExternalServiceSelected]
    )

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
            <div className={styles.selectService}>
                <Select
                    aria-label="External service"
                    className="mb-0"
                    onChange={({ target: { value } }) => onSelect(value)}
                    value={externalService}
                >
                    <option key="all" value="all">
                        All webhooks
                    </option>
                    <option key="unmatched" value="unmatched">
                        Unmatched webhooks
                    </option>
                    {data?.externalServices.nodes.map(({ displayName, id }) => (
                        <option key={id} value={id}>
                            {displayName}
                        </option>
                    ))}
                </Select>
            </div>
            <div className={styles.errorButton}>
                <Button variant="danger" onClick={onErrorToggle} outline={!onlyErrors}>
                    <Icon
                        className={classNames(styles.icon, onlyErrors && styles.enabled)}
                        aria-hidden={true}
                        svgPath={mdiAlertCircle}
                    />
                    <span className="ml-1">Only errors</span>
                </Button>
            </div>
        </div>
    )
}
