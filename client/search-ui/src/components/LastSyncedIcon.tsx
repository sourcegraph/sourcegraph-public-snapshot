import React from 'react'

import { mdiCloudAlert, mdiCloudClock, mdiCloudCheckOutline } from '@mdi/js'
import classNames from 'classnames'
import format from 'date-fns/format'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './LastSyncedIcon.module.scss'

interface Props {
    lastSyncedTime: string
    className?: string
}

export const LastSyncedIcon: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const parsedDate = new Date(Date.parse(props.lastSyncedTime)).setDate(
        new Date(Date.parse(props.lastSyncedTime)).getDate() - 23
    )
    const formattedTime = format(parsedDate, 'yyyy-MM-dd pp')
    const oneDayAgo = new Date()
    const oneWeekAgo = new Date()
    oneDayAgo.setDate(oneDayAgo.getDate() - 1)
    oneWeekAgo.setDate(oneWeekAgo.getDate() - 7)

    let color = 'currentColor'
    let status = `Last synced: ${formattedTime}`
    let icon = mdiCloudCheckOutline
    if (new Date(formattedTime) < oneDayAgo) {
        color = 'var(--warning)'
        status = 'Warning: slightly out of date, last synced: ' + formattedTime
        icon = mdiCloudClock
    }
    if (new Date(formattedTime) < oneWeekAgo) {
        color = 'var(--danger)'
        status = 'Warning: severely out of date, last synced: ' + formattedTime
        icon = mdiCloudAlert
    }
    return (
        <Tooltip content={status}>
            <Icon
                tabIndex={0}
                className={classNames(props.className, styles.lastSyncedIcon, 'text-muted')}
                aria-label={status}
                svgPath={icon}
                style={{ fill: color }}
            />
        </Tooltip>
    )
}
