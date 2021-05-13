import classNames from 'classnames'
import * as React from 'react'

import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

interface ConnectionNodesSummaryShowMoreProps {
    onShowMore?: () => void
    showMoreClassName?: string
}

export const ConnectionNodesSummaryShowMore: React.FunctionComponent<ConnectionNodesSummaryShowMoreProps> = ({
    onShowMore,
    showMoreClassName,
}) => {
    const [isRedesignEnabled] = useRedesignToggle()

    return (
        <button
            type="button"
            className={classNames(
                'btn btn-sm filtered-connection__show-more',
                isRedesignEnabled ? 'btn-link' : 'btn-secondary',
                showMoreClassName
            )}
            onClick={onShowMore}
        >
            Show more
        </button>
    )
}

interface ConnectionNodesSummaryProps {
    summary: React.ReactFragment | undefined
    className?: string
}

export const ConnectionNodesSummary: React.FunctionComponent<ConnectionNodesSummaryProps> = ({
    summary,
    className,
    children,
}) => {
    if (!summary && !children) {
        return null
    }

    return (
        <div className={classNames('filtered-connection__summary-container', className)}>
            {summary}
            {children}
        </div>
    )
}
