import classNames from 'classnames'
import React from 'react'

import styles from './LoadingSpinner.module.scss'

export interface LoadingSpinnerProps {
    className?: string
    /**
     * Whether to show loading spinner with icon-inline
     *
     * @default true
     */
    inline?: boolean
}

export const LoadingSpinner: React.FunctionComponent<LoadingSpinnerProps> = ({ inline = true, className }) => (
    <div className={classNames(styles.loadingSpinner, inline && 'icon-inline', className)} />
)
