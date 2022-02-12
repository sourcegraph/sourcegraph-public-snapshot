import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { useCallback, useMemo, useState } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import { CodeMonitorWithEvents, EventStatus } from '../../../../graphql-operations'

import styles from './MonitorLogNode.module.scss'

export const MonitorLogNode: React.FunctionComponent<{ monitor: CodeMonitorWithEvents }> = ({ monitor }) => {
    const [expanded, setExpanded] = useState(false)

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
        [monitor.trigger.events.nodes]
    )

    // The most recent event is the first one in the list.
    const lastRun = useMemo(
        () => (monitor.trigger.events.nodes.length > 0 ? monitor.trigger.events.nodes[0].timestamp : 'Never'),
        [monitor.trigger.events.nodes]
    )

    return (
        <>
            <span className={styles.separator} />
            <Button onClick={toggleExpanded} className="btn-icon mr-2">
                {expanded ? <ChevronDownIcon /> : <ChevronRightIcon />}
            </Button>
            {hasError ? <AlertCircleIcon className={classNames(styles.errorIcon, 'icon-inline')} /> : <span />}
            <span>{monitor.description}</span>
            <span className="text-nowrap">
                <Timestamp date={lastRun} />
            </span>

            {expanded && (
                <div className={styles.expandedRow}>
                    <pre>{monitor.trigger.query}</pre>
                </div>
            )}
        </>
    )
}
