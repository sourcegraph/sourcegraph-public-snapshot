import React from 'react'

import classNames from 'classnames'

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

export const LoadingSpinner: React.FunctionComponent<React.PropsWithChildren<LoadingSpinnerProps>> = ({
    inline = true,
    className,
    ...props
}) => {
    const finalClassName = classNames(styles.loadingSpinner, className)

    return (
        <Icon inline={inline} aria-label="Loading" aria-live="polite" className={finalClassName} as="div" {...props} />
    )
}
