import classnames from 'classnames'
import React from 'react'

import { TruncatedText } from '../trancated-text/TrancatedText'

import styles from './Badge.module.scss'

interface BadgeProps {
    value: string
    className?: string
}

export const Badge: React.FunctionComponent<BadgeProps> = props => {
    const { value, className } = props

    return (
        <TruncatedText title={value} className={classnames(styles.badge, 'badge', 'badge-secondary', className)}>
            {value}
        </TruncatedText>
    )
}
