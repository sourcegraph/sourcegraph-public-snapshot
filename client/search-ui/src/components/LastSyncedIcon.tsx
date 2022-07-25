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
    const formattedTime = format(Date.parse(props.lastSyncedTime), "yyyy-MM-dd'T'HH:mm:ss")

    return (
        <Tooltip content={`Last synced: ${formattedTime}`}>
            <Icon
                tabIndex={0}
                className={classNames(props.className, styles.lastSyncedIcon, 'text-muted')}
                aria-label={`Last synced: ${formattedTime}`}
                svgPath={mdiWeatherCloudyClock}
            />
        </Tooltip>
    )
}
