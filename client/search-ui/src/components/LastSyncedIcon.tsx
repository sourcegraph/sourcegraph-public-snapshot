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
    const oneDayAgo = new Date();
    const oneWeekAgo = new Date();
    oneDayAgo.setDate(oneDayAgo.getDate() - 1);
    oneWeekAgo.setDate(oneWeekAgo.getDate()-7);

    let color ='green';
    if (new Date(parsedDate) < oneDayAgo) {
        color = 'yellow';
    }
    if (new Date(parsedDate) < oneWeekAgo){
        color = 'red';
    }
    return (
        <Tooltip content={`Last synced: ${formattedTime}`}>
            <Icon
                tabIndex={0}
                className={classNames(props.className, styles.lastSyncedIcon, 'text-muted')}
                aria-label={`Last synced: ${formattedTime}`}
                svgPath={mdiWeatherCloudyClock}
                style={{ fill: color }}
            />
        </Tooltip>
    )
}
