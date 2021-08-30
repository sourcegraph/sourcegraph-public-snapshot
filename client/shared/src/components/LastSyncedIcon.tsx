import CloudSyncOutlineIcon from 'mdi-react/CloudSyncOutlineIcon'
import React from 'react'

export const LastSyncedIcon = (lastSyncedTime: string) => {
    return (
        <CloudSyncOutlineIcon
            className="last-synced-icon icon-inline text-muted"
            data-tooltip={`Last synced: ${lastSyncedTime}`}
        />
    )
}
