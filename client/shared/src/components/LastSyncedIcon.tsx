import classNames from 'classnames'
import format from 'date-fns/format'
import CloudSyncOutlineIcon from 'mdi-react/CloudSyncOutlineIcon'
import React from 'react'

import styles from './LastSyncedIcon.module.scss'

export interface Props {
    lastSyncedTime: string
    className?: string
}

export const LastSyncedIcon: React.FunctionComponent<Props> = props => {
    const formattedTime = format(Date.parse(props.lastSyncedTime), "yyyy-MM-dd'T'HH:mm:ss")

    return (
        <CloudSyncOutlineIcon
            className={classNames(props.className, styles.lastSyncedIcon, 'icon-inline', 'text-muted')}
            data-tooltip={`Last synced: ${formattedTime}`}
        />
    )
}
