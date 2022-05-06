import React from 'react'

import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'

import styles from './CaptureGroupSeriesInfoBadge.module.scss'

export const CaptureGroupSeriesInfoBadge: React.FunctionComponent<React.PropsWithChildren<unknown>> = props => (
    <div className={classNames(styles.badge, 'text-muted')}>
        <InformationOutlineIcon className={styles.icon} />
        <small>{props.children}</small>
    </div>
)
