import React, { useEffect, useState } from 'react'

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
    /**
     * If true, the spinner will not render until after 100ms.
     */
    delay?: boolean
}

/**
 * 300ms delay. TODO: Test different values. Allow configurable?
 */
const DELAY_TIME = 300

export const LoadingSpinner: React.FunctionComponent<React.PropsWithChildren<LoadingSpinnerProps>> = ({
    inline = true,
    delay = true,
    className,
    ...props
}) => {
    const [shouldShow, setShouldShow] = useState(false)

    useEffect(() => {
        const timer = setTimeout(() => {
            setShouldShow(true)
        }, DELAY_TIME)
        return () => clearTimeout(timer)
    }, [])

    if (!shouldShow) {
        return null
    }

    const finalClassName = classNames(styles.loadingSpinner, className)

    return (
        <Icon inline={inline} aria-label="Loading" aria-live="polite" className={finalClassName} as="div" {...props} />
    )
}
