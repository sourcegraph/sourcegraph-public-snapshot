import classNames from 'classnames'
import * as React from 'react'

import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

interface ConnectionNodesSummaryProps {
    summary: React.ReactFragment | undefined
    displayShowMoreButton?: boolean
    onShowMore?: () => void
    showMoreClassName?: string
}

export const ConnectionNodesSummaryShowMore: React.FunctionComponent<ConnectionNodesSummaryShowMoreProps> = ({
    onShowMore,
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
        >
            Show more
        </button>
    )

    if (isRedesignEnabled) {
        return (
            <div className="filtered-connection__summary-container">
                {summary}
                {showMoreButton}
            </div>
        )
    }

    return (
        <>
            {summary}
            {showMoreButton}
        </>
    )
}
