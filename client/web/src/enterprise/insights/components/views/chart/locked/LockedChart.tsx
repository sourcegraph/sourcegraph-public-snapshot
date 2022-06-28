import React from 'react'

import classNames from 'classnames'
import LockIcon from 'mdi-react/LockOutlineIcon'

import styles from './LockedChart.module.scss'

export const LockedChart: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <section className={classNames(styles.wrapper, className)}>
        <LockIcon size={40} />
        <div className={classNames(styles.banner)}>
            <span>Limited access</span>
            <small>Insight locked</small>
        </div>
    </section>
)
