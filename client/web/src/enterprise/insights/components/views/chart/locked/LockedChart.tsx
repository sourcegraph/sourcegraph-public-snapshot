import React from 'react'

import { mdiLock } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '@sourcegraph/wildcard'

import styles from './LockedChart.module.scss'

export const LockedChart: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <section className={classNames(styles.wrapper, className)}>
        <Icon svgPath={mdiLock} height={40} width={40} inline={false} aria-hidden={true} />
        <div className={classNames(styles.banner)}>
            <span>Limited access</span>
            <small>Insight locked</small>
        </div>
    </section>
)
