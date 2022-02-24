import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

import styles from './CaptureGroupSeriesInfoBadge.module.scss'

export const CaptureGroupSeriesInfoBadge: React.FunctionComponent = props => (
    <div className={classNames(styles.badge, 'text-muted')}>
        <Icon as={InformationOutlineIcon} inline={false} className={styles.icon} />
        <small>{props.children}</small>
    </div>
)
