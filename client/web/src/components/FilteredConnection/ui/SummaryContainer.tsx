import classNames from 'classnames'
import React from 'react'

interface SummaryContainerProps {
    className?: string
}

/**
 * FilteredConnection styled summary container to support advanced styling.
 * Should wrap typically wrap <ConnectionSummary>.
 * May also be used to wrap <ShowMoreButton />.
 */
export const SummaryContainer: React.FunctionComponent<SummaryContainerProps> = ({ children, className }) => (
    <div className={classNames('filtered-connection__summary-container', className)}>{children}</div>
)
