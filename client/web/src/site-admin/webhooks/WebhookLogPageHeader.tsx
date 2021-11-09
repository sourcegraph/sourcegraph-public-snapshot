import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback } from 'react'

import { useQuery } from '@sourcegraph/shared/src/graphql/graphql'

import { WebhookLogPageHeaderResult } from '../../graphql-operations'

import { SelectedExternalService, WEBHOOK_LOG_PAGE_HEADER } from './backend'
import { PerformanceGauge } from './PerformanceGauge'
import styles from './WebhookLogPageHeader.module.scss'

export interface Props {
    externalService: SelectedExternalService
    onlyErrors: boolean

    onExternalServiceSelected?: (externalService: SelectedExternalService) => void
    onSetErrors?: (onlyErrors: boolean) => void
}

export const WebhookLogPageHeader: React.FunctionComponent<Props> = ({
    externalService,
    onlyErrors,
    onExternalServiceSelected,
    onSetErrors,
}) => {
    const onErrorToggle = useCallback(() => onSetErrors?.(!onlyErrors), [onlyErrors, onSetErrors])
    const onSelect = useCallback(
        (value: string) => {
            onExternalServiceSelected?.(value)
        },
        [onExternalServiceSelected]
    )

    const { data } = useQuery<WebhookLogPageHeaderResult>(WEBHOOK_LOG_PAGE_HEADER, {})
    const errorCount = data?.webhookLogs.totalCount ?? 0
    console.log(data)

    return (
        <div className="d-flex align-items-end">
            <PerformanceGauge
                className="mr-3"
                count={data?.webhookLogs.totalCount}
                countClassName={errorCount > 0 ? 'text-danger' : undefined}
                label="recent error"
            />
            <PerformanceGauge count={data?.externalServices.totalCount} label="external service" />
            <div className="flex-fill" />
            <div>
                <select
                    className={classNames('form-control', styles.control)}
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
                </select>
            </div>
            <div>
                <button
                    type="button"
                    className={classNames(
                        'btn',
                        'ml-3',
                        'd-flex',
                        'align-items-center',
                        styles.control,
                        onlyErrors ? 'btn-secondary' : 'btn-outline-secondary'
                    )}
                    onClick={onErrorToggle}
                >
                    <AlertCircleIcon className={styles.icon} />
                    <span className="ml-1">Only errors</span>
                </button>
            </div>
        </div>
    )
}
