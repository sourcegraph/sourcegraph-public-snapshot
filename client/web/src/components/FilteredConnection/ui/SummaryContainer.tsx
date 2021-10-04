import classNames from 'classnames'
import React from 'react'

import styles from './SummaryContainer.module.scss'

interface SummaryContainerProps {
    className?: string
    compact?: boolean
    centered?: boolean
}

/**
 * FilteredConnection styled summary container to support advanced styling.
 * Should wrap typically wrap <ConnectionSummary>.
 * May also be used to wrap <ShowMoreButton />.
 */
export const SummaryContainer: React.FunctionComponent<SummaryContainerProps> = ({
    children,
    className,
    centered,
    compact,
}) => (
    <div className={classNames(styles.normal, compact && styles.compact, centered && styles.centered, className)}>
        {children}
    </div>
)
