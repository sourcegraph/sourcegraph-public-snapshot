import classNames from 'classnames'
import React from 'react'

import { Icon } from '../Icon'

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
    <Icon className={classNames(styles.loadingSpinner, className)} inline={inline} as="div" />
)
