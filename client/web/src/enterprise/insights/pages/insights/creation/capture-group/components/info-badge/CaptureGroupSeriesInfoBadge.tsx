import React from 'react'

import classNames from 'classnames'

import styles from './CaptureGroupSeriesInfoBadge.module.scss'
import { mdiInformationOutline } from "@mdi/js";
import { Icon } from "@sourcegraph/wildcard";

export const CaptureGroupSeriesInfoBadge: React.FunctionComponent<React.PropsWithChildren<unknown>> = props => (
    <div className={classNames(styles.badge, 'text-muted')}>
        <Icon className={styles.icon} svgPath={mdiInformationOutline} inline={false} aria-hidden={true} />
        <small>{props.children}</small>
    </div>
)
