import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { useCallback, useState } from 'react'

import { Badge, Button } from '@sourcegraph/wildcard'

import { EventStatus } from '../../../../graphql-operations'

import styles from './CollapsibleDetailsWithStatus.module.scss'

export const CollapsibleDetailsWithStatus: React.FunctionComponent<{
    title: string
    status: EventStatus
    message: string
    startOpen?: boolean
}> = ({ title, status, message, startOpen = false }) => {
    const [expanded, setExpanded] = useState(startOpen)

    const toggleExpanded = useCallback(() => setExpanded(expanded => !expanded), [])

    const statusBadge = status === EventStatus.ERROR ? 'danger' : status === EventStatus.PENDING ? 'warning' : 'primary'
    const statusText = status === EventStatus.ERROR ? 'Error' : status === EventStatus.PENDING ? 'Pending' : 'Success'

    return (
        <div className={styles.wrapper}>
            <Button onClick={toggleExpanded} className={classNames('btn-icon d-block', styles.expandButton)}>
                {expanded ? (
                    <ChevronDownIcon className="icon-inline mr-2" />
                ) : (
                    <ChevronRightIcon className="icon-inline mr-2" />
                )}
                <span>{title}</span>
                <Badge variant={statusBadge} className="ml-2 text-uppercase">
                    {statusText}
                </Badge>
            </Button>

            {expanded && <pre className={styles.message}>{message}</pre>}
        </div>
    )
}
