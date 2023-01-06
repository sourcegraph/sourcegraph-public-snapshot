import React from 'react'

import { mdiWeatherCloudyClock } from '@mdi/js'
import classNames from 'classnames'
import format from 'date-fns/format'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './LastSyncedIcon.module.scss'

interface Props {
    lastSyncedTime: string
    className?: string
}

export const LastSyncedIcon: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const parsedDate = Date.parse(props.lastSyncedTime)
    const formattedTime = format(parsedDate, 'yyyy-MM-dd pp')
    const oneDayAgo = new Date()
    const oneWeekAgo = new Date()
    oneDayAgo.setDate(oneDayAgo.getDate() - 1)
    oneWeekAgo.setDate(oneWeekAgo.getDate() - 7)

    let color = '#17ab52' // Green 400
    let status = `Last synced: ${formattedTime}`
    if (new Date(formattedTime) < oneDayAgo) {
        color = '#ff9933' // Lemon 400
        status = 'Warning: slightly out of date, last synced: ' + formattedTime
    }
    if (new Date(formattedTime) < oneWeekAgo) {
        color = '#ed2e20' // Vermillion 400
        status = 'Warning: severely out of date, last synced: ' + formattedTime
    }
    return (
        <Tooltip content={status}>
            <Icon
                tabIndex={0}
                className={classNames(props.className, styles.lastSyncedIcon, 'text-muted')}
                aria-label={status}
                svgPath={mdiWeatherCloudyClock}
                style={{ fill: color }}
            />
        </Tooltip>
    )
}
