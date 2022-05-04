import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'

import { pluralize } from '@sourcegraph/common'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, Link, Icon } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import {
    EventStatus,
    MonitorActionEvents,
    MonitorTriggerEventWithActions,
    SearchPatternType,
} from '../../../../graphql-operations'

import { CollapsibleDetailsWithStatus } from './CollapsibleDetailsWithStatus'

import styles from './TriggerEvent.module.scss'

export const TriggerEvent: React.FunctionComponent<
    React.PropsWithChildren<{
        triggerEvent: MonitorTriggerEventWithActions
        startOpen?: boolean
        now?: () => Date
    }>
> = ({ triggerEvent, startOpen = false, now }) => {
    const [expanded, setExpanded] = useState(startOpen)

    const toggleExpanded = useCallback(() => setExpanded(expanded => !expanded), [])

    // Either there's an error in the trigger itself, or in any of the actions.
    const hasError = useMemo(
        () =>
            triggerEvent.status === EventStatus.ERROR ||
            triggerEvent.actions.nodes.some(action =>
                action.events.nodes.some(actionEvent => actionEvent.status === EventStatus.ERROR)
            ),
        [triggerEvent]
    )

    function getTriggerEventMessage(): string {
        if (triggerEvent.message) {
            return triggerEvent.message
        }

        switch (triggerEvent.status) {
            case EventStatus.ERROR:
                return 'Unknown error occurred when running the search'
            case EventStatus.PENDING:
                return 'Search is pending'
            default:
                return 'Search ran successfully'
        }
    }

    return (
        <>
            <Button onClick={toggleExpanded} className={classNames('btn-icon d-block', styles.expandButton)}>
                <Icon className="mr-2" as={expanded ? ChevronDownIcon : ChevronRightIcon} />

                {hasError ? <Icon className={classNames(styles.errorIcon, 'mr-2')} as={AlertCircleIcon} /> : <span />}

                <span>
                    Run <Timestamp date={triggerEvent.timestamp} noAbout={true} now={now} />
                    {triggerEvent.query && (
                        <Link
                            to={`/search?${buildSearchURLQuery(triggerEvent.query, SearchPatternType.literal, false)}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="font-weight-normal ml-2"
                        >
                            {triggerEvent.resultCount} new {pluralize('result', triggerEvent.resultCount)}{' '}
                            <Icon as={OpenInNewIcon} />
                        </Link>
                    )}
                </span>
            </Button>

            {expanded && (
                <>
                    <CollapsibleDetailsWithStatus
                        status={triggerEvent.status}
                        message={getTriggerEventMessage()}
                        title="Monitor trigger"
                        startOpen={startOpen}
                    />

                    {triggerEvent.actions.nodes.map(action => (
                        <>
                            {action.events.nodes.map(actionEvent => (
                                <CollapsibleDetailsWithStatus
                                    key={actionEvent.id}
                                    status={actionEvent.status}
                                    message={getActionEventMessage(actionEvent)}
                                    title={getActionEventTitle(action)}
                                    startOpen={startOpen}
                                />
                            ))}

                            {action.events.nodes.length === 0 && (
                                <CollapsibleDetailsWithStatus
                                    status="skipped"
                                    message="This action was not run because it was disabled or there were no new results."
                                    title={getActionEventTitle(action)}
                                    startOpen={startOpen}
                                />
                            )}
                        </>
                    ))}
                </>
            )}
        </>
    )
}

function getActionEventMessage(actionEvent: MonitorActionEvents['nodes'][number]): string {
    if (actionEvent.message) {
        return actionEvent.message
    }

    switch (actionEvent.status) {
        case EventStatus.ERROR:
            return 'Unknown error occurred when sending the notification'
        case EventStatus.PENDING:
            return 'Notification is pending'
        default:
            return 'Notification sent successfully'
    }
}

function getActionEventTitle(action: MonitorTriggerEventWithActions['actions']['nodes'][number]): string {
    switch (action.__typename) {
        case 'MonitorEmail':
            return 'Email'
        case 'MonitorSlackWebhook':
            return 'Slack'
        case 'MonitorWebhook':
            return 'Webhook'
    }
}
