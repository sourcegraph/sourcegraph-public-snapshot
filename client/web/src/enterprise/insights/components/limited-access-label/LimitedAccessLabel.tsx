import React from 'react'

import classNames from 'classnames'

import { Typography } from '@sourcegraph/wildcard'

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
        <Typography.Label className={styles.label} isUppercase={true}>
            <small>{label || 'Limited access'}</small>
        </Typography.Label>
        <span className={classNames(styles.message, 'small')}>{message}</span>
    </div>
)
