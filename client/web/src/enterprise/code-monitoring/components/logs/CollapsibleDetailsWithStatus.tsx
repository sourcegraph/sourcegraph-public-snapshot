import React, { useCallback, useMemo, useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'

import { Badge, Button, Icon } from '@sourcegraph/wildcard'

import { EventStatus } from '../../../../graphql-operations'

import styles from './CollapsibleDetailsWithStatus.module.scss'

export const CollapsibleDetailsWithStatus: React.FunctionComponent<
    React.PropsWithChildren<{
        title: string
        status: EventStatus | 'skipped'
        message: string
        startOpen?: boolean
    }>
> = ({ title, status, message, startOpen = false }) => {
    const [expanded, setExpanded] = useState(startOpen)

    const toggleExpanded = useCallback(() => setExpanded(expanded => !expanded), [])

    const statusBadge = useMemo(() => {
        switch (status) {
            case EventStatus.ERROR: {
                return 'danger'
            }
            case EventStatus.PENDING: {
                return 'warning'
            }
            case EventStatus.SUCCESS: {
                return 'primary'
            }
            case 'skipped': {
                return 'warning'
            }
        }
    }, [status])

    const statusText = useMemo(() => {
        switch (status) {
            case EventStatus.ERROR: {
                return 'Error'
            }
            case EventStatus.PENDING: {
                return 'Pending'
            }
            case EventStatus.SUCCESS: {
                return 'Success'
            }
            case 'skipped': {
                return 'Skipped'
            }
        }
    }, [status])

    return (
        <li className={styles.wrapper}>
            <Button onClick={toggleExpanded} className={styles.expandButton}>
                {expanded ? (
                    <Icon svgPath={mdiChevronUp} className="mr-2" aria-label="Collapse details." />
                ) : (
                    <Icon svgPath={mdiChevronDown} className="mr-2" aria-label="Expand details." />
                )}
                <span>{title}</span>
                <Badge
                    variant={statusBadge}
                    className="ml-2 text-uppercase"
                    aria-label={`, Monitor state: ${statusText}`}
                >
                    {statusText}
                </Badge>
            </Button>

            {expanded && <pre className={styles.message}>{message}</pre>}
        </li>
    )
}
