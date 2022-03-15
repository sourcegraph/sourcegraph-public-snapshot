import classNames from 'classnames'
import LockIcon from 'mdi-react/LockOutlineIcon'
import React from 'react'

import styles from './LockedBanner.module.scss'

export const LockedBanner: React.FunctionComponent = () => (
    <div className={classNames(styles.wrapper)}>
        <LockIcon size={40} className={classNames(styles.icon)} />
        <div>
            <span className={classNames(styles.label)}>Limited access</span>
            <small className={classNames(styles.message)}>Insight locked</small>
        </div>
    </div>
)
