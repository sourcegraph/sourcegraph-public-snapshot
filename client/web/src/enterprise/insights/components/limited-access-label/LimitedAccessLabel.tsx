import React from 'react'

import classNames from 'classnames'

import styles from './LimitedAccessLabel.module.scss'

interface LimitedAccessLabelProps {
    className?: string
    label?: string
    message: string
}

export const LimitedAccessLabel: React.FunctionComponent<React.PropsWithChildren<LimitedAccessLabelProps>> = ({
    message,
    label,
    className,
}) => (
    <div className={classNames(styles.wrapper, className)}>
        <label className={classNames(styles.label, 'text-uppercase')}>
            <small>{label || 'Limited access'}</small>
        </label>
        <span className={classNames(styles.message, 'small')}>{message}</span>
    </div>
)
