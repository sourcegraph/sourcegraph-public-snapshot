import classNames from 'classnames'
import format from 'date-fns/format'
import WeatherCloudyClockIcon from 'mdi-react/WeatherCloudyClockIcon'
import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

import styles from './LastSyncedIcon.module.scss'

export interface Props {
    lastSyncedTime: string
    className?: string
}

export const LastSyncedIcon: React.FunctionComponent<Props> = props => {
    const formattedTime = format(Date.parse(props.lastSyncedTime), "yyyy-MM-dd'T'HH:mm:ss")

    return (
        <Icon
            className={classNames(props.className, styles.lastSyncedIcon, 'text-muted')}
            data-tooltip={`Last synced: ${formattedTime}`}
            as={WeatherCloudyClockIcon}
        />
    )
}
