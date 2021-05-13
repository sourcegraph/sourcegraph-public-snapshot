import classNames from 'classnames'
import * as React from 'react'

import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

interface ConnectionNodesSummaryProps {
    summary: React.ReactFragment | undefined
    displayShowMoreButton?: boolean
    onShowMore?: () => void
    showMoreClassName?: string
    hasNextPage?: boolean
    className?: string
}

export const ConnectionNodesSummary: React.FunctionComponent<ConnectionNodesSummaryProps> = ({
    summary,
    displayShowMoreButton,
    showMoreClassName,
    onShowMore,
    hasNextPage,
    className,
}) => {
    const [isRedesignEnabled] = useRedesignToggle()

    const showMoreButton = displayShowMoreButton && (
        <button
            type="button"
            className={classNames(
                'btn btn-sm filtered-connection__show-more',
                isRedesignEnabled ? 'btn-link' : 'btn-secondary',
                showMoreClassName
            )}
            onClick={onShowMore}
            disabled={!hasNextPage}
        >
            Show more
        </button>
    )

    return (
        <div className={classNames('filtered-connection__summary-container', className)}>
            {summary}
            {showMoreButton}
        </div>
    )
}
