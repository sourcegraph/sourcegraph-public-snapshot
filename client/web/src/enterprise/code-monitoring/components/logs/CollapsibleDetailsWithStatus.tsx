import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

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
            case EventStatus.ERROR:
                return 'danger'
            case EventStatus.PENDING:
                return 'warning'
            case EventStatus.SUCCESS:
                return 'primary'
            case 'skipped':
                return 'warning'
        }
    }, [status])

    const statusText = useMemo(() => {
        switch (status) {
            case EventStatus.ERROR:
                return 'Error'
            case EventStatus.PENDING:
                return 'Pending'
            case EventStatus.SUCCESS:
                return 'Success'
            case 'skipped':
                return 'Skipped'
        }
    }, [status])

    return (
        <div className={styles.wrapper}>
            <Button onClick={toggleExpanded} className={classNames('btn-icon d-block', styles.expandButton)}>
                <Icon className="mr-2" as={expanded ? ChevronDownIcon : ChevronRightIcon} />
                <span>{title}</span>
                <Badge variant={statusBadge} className="ml-2 text-uppercase">
                    {statusText}
                </Badge>
            </Button>

            {expanded && <pre className={styles.message}>{message}</pre>}
        </div>
    )
}
