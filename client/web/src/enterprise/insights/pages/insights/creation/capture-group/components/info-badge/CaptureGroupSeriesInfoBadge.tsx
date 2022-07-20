import React from 'react'

import { mdiInformationOutline } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '@sourcegraph/wildcard'

import styles from './CaptureGroupSeriesInfoBadge.module.scss'

export const CaptureGroupSeriesInfoBadge: React.FunctionComponent<React.PropsWithChildren<unknown>> = props => (
    <div className={classNames(styles.badge, 'text-muted')}>
        <Icon className={styles.icon} svgPath={mdiInformationOutline} inline={false} aria-hidden={true} />
        <small>{props.children}</small>
    </div>
)
