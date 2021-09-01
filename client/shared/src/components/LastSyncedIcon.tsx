import classNames from 'classnames'
import format from 'date-fns/format'
import CloudSyncOutlineIcon from 'mdi-react/CloudSyncOutlineIcon'
import React from 'react'
import styles from './LastSyncedIcon.module.scss'

export const LastSyncedIcon: React.FunctionComponent<{ lastSyncedTime: string }> = ({ lastSyncedTime }) => {
    const formattedTime = format(Date.parse(lastSyncedTime), "yyyy-MM-dd'T'HH:mm:ss")

    return (<CloudSyncOutlineIcon
		className={classNames(styles.lastSyncedIcon, "icon-inline", "text-muted")}
		data-tooltip={`Last synced: ${formattedTime}`}
	/>)
}
