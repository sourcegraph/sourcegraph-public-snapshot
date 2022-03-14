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

export const LoadingSpinner: React.FunctionComponent<LoadingSpinnerProps> = ({ inline = true, className }) => {
    const finalClassName = classNames(styles.loadingSpinner, className)

    if (inline) {
        return <Icon className={finalClassName} as="div" />
    }

    return <div className={finalClassName} />
}
