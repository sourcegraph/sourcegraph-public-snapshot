import { type FC, useCallback } from 'react'

import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Button, Checkbox, ErrorAlert, H3, Label, LoadingSpinner } from '@sourcegraph/wildcard'

import type {
    OutboundWebhookEventTypeFields,
    OutboundWebhookEventTypesResult,
    OutboundWebhookEventTypesVariables,
} from '../../../graphql-operations'
import { OUTBOUND_WEBHOOK_EVENT_TYPES } from '../backend'

import styles from './EventTypes.module.scss'

export interface EventTypesProps {
    className?: string
    onChange: (values: Set<string>) => void
    values: Set<string>
}

export const EventTypes: FC<EventTypesProps> = ({ className, onChange, values }) => {
    const { data, loading, error } = useQuery<OutboundWebhookEventTypesResult, OutboundWebhookEventTypesVariables>(
        OUTBOUND_WEBHOOK_EVENT_TYPES,
        {}
    )

    const deselectAll = useCallback(() => {
        onChange(new Set())
    }, [onChange])

    const selectAll = useCallback(() => {
        onChange(new Set(data?.outboundWebhookEventTypes.map(({ key }) => key)))
    }, [data, onChange])

    const onToggle = useCallback(
        (key: string, checked: boolean) => {
            const newValues = new Set(values)
            if (checked) {
                newValues.add(key)
            } else {
                newValues.delete(key)
            }

            onChange(newValues)
        },
        [onChange, values]
    )

    return (
        <div className={className}>
            <div className={classNames('mb-2', styles.heading)}>
                <H3>Event types</H3>
                {data?.outboundWebhookEventTypes &&
                    (values.size >= data.outboundWebhookEventTypes.length ? (
                        <Button variant="secondary" onClick={deselectAll}>
                            Deselect all
                        </Button>
                    ) : (
                        <Button variant="secondary" onClick={selectAll}>
                            Select all
                        </Button>
                    ))}
            </div>
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert error={error} />}
            {data?.outboundWebhookEventTypes.map(({ key, description }) => (
                <Checkbox
                    key={key}
                    id={key}
                    label={<EventTypeLabel eventType={{ key, description }} />}
                    checked={values.has(key)}
                    onChange={event => onToggle(key, event.target.checked)}
                />
            ))}
        </div>
    )
}

const EventTypeLabel: FC<{
    eventType: OutboundWebhookEventTypeFields
}> = ({ eventType }) => (
    <Label htmlFor={eventType.key}>
        <div>{eventType.key}</div>
        <small className="text-muted">{eventType.description}</small>
    </Label>
)
