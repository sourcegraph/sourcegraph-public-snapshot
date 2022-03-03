import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckBoldIcon from 'mdi-react/CheckBoldIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { useCallback, useMemo, useState } from 'react'

import { Button, Link } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import { CodeMonitorWithEvents, EventStatus } from '../../../../graphql-operations'

import styles from './MonitorLogNode.module.scss'
import { TriggerEvent } from './TriggerEvent'

const clickCatcher = (event: React.MouseEvent<HTMLAnchorElement>): void => {
    event.stopPropagation()
}

export const MonitorLogNode: React.FunctionComponent<{
    monitor: CodeMonitorWithEvents
    now?: () => Date
    startOpen?: boolean
}> = ({ monitor, now, startOpen = false }) => {
    const [expanded, setExpanded] = useState(startOpen)

    const toggleExpanded = useCallback(() => setExpanded(expanded => !expanded), [])

    // Either there's an error in the trigger itself, or in any of the actions.
    const hasError = useMemo(
        () =>
            monitor.trigger.events.nodes.some(
                triggerEvent =>
                    triggerEvent.status === EventStatus.ERROR ||
                    triggerEvent.actions.nodes.some(action =>
                        action.events.nodes.some(actionEvent => actionEvent.status === EventStatus.ERROR)
                    )
            ),
        [monitor]
    )

    // The most recent event is the first one in the list.
    const lastRun = useMemo(
        () => (monitor.trigger.events.nodes.length > 0 ? monitor.trigger.events.nodes[0].timestamp : null),
        [monitor.trigger.events.nodes]
    )

    return (
        <>
            <span className={styles.separator} />
            <Button
                onClick={toggleExpanded}
                className="btn-icon text-left pl-0 border-0 d-flex align-items-center"
                aria-label="Expand code monitor"
            >
                {expanded ? (
                    <ChevronDownIcon className="mr-2 flex-shrink-0" />
                ) : (
                    <ChevronRightIcon className="mr-2 flex-shrink-0" />
                )}
                {hasError ? (
                    <AlertCircleIcon
                        className={classNames(styles.errorIcon, 'icon-inline mr-1 flex-shrink-0')}
                        aria-label="One or more runs of this code monitor have an error"
                        data-tooltip="One or more runs of this code monitor have an error"
                        data-placement="top"
                    />
                ) : (
                    <CheckBoldIcon
                        className={classNames(styles.checkIcon, 'icon-inline mr-1 flex-shrink-0')}
                        aria-label="Monitor running as normal"
                        data-tooltip="Monitor running as normal"
                        data-placement="top"
                    />
                )}
                {monitor.description}
                {/* Use clickCatcher so clicking on link doesn't expand/collapse row */}
                <Link to={`/code-monitoring/${monitor.id}`} className="ml-2 font-weight-normal" onClick={clickCatcher}>
                    Monitor details
                </Link>
            </Button>
            <span className="text-nowrap mr-2">
                {lastRun ? <Timestamp date={lastRun} now={now} noAbout={true} /> : <>Never</>}
            </span>

            {expanded && (
                <div className={styles.expandedRow}>
                    {monitor.trigger.events.nodes.map(triggerEvent => (
                        <TriggerEvent
                            key={triggerEvent.id}
                            triggerEvent={triggerEvent}
                            startOpen={startOpen}
                            now={now}
                        />
                    ))}

                    {monitor.trigger.events.nodes.length === 0 && <div>This code monitor has not been run yet.</div>}
                </div>
            )}
        </>
    )
}
