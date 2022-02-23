import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { useCallback, useMemo, useState } from 'react'

import { Button, Link } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import { CodeMonitorWithEvents, EventStatus } from '../../../../graphql-operations'

import styles from './MonitorLogNode.module.scss'
import { TriggerEvent } from './TriggerEvent'

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
                className="btn-icon text-left pl-0 d-flex align-items-center"
                aria-label="Expand code monitor"
            >
                {expanded ? <ChevronDownIcon className="mr-2" /> : <ChevronRightIcon className="mr-1" />}
                {hasError ? (
                    <AlertCircleIcon
                        className={classNames(styles.errorIcon, 'icon-inline mr-1')}
                        aria-label="A run of this code monitor has an error"
                    />
                ) : (
                    <span className={classNames(styles.errorIconSpacer, 'mr-1')} />
                )}
                {monitor.description}
            </Button>
            <span className="text-nowrap mr-2">
                {lastRun ? <Timestamp date={lastRun} now={now} noAbout={true} /> : <>Never</>}
            </span>

            {expanded && (
                <div className={styles.expandedRow}>
                    <Link to={`/code-monitoring/${monitor.id}`} className="d-block mb-3">
                        View code monitor details
                    </Link>

                    {monitor.trigger.events.nodes.map(triggerEvent => (
                        <TriggerEvent key={triggerEvent.id} triggerEvent={triggerEvent} startOpen={startOpen} />
                    ))}

                    {monitor.trigger.events.nodes.length === 0 && <div>This code monitor has not been run yet.</div>}
                </div>
            )}
        </>
    )
}
