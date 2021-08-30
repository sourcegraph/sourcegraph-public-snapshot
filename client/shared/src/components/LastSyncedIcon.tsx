import CloudSyncOutlineIcon from 'mdi-react/CloudSyncOutlineIcon'
import React from 'react'
import format from 'date-fns/format'

export const LastSyncedIcon = (lastSyncedTime: string) => {
    const formattedTime = format(Date.parse(lastSyncedTime), "yyyy-MM-dd'T'HH:mm:ss")

    return (
        <CloudSyncOutlineIcon
            className="last-synced-icon icon-inline text-muted"
            data-tooltip={`Last synced: ${formattedTime}`}
        />
    )
}
