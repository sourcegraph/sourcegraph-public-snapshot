import classNames from 'classnames'
import * as React from 'react'

interface ConnectionNodesSummaryProps {
    summary: React.ReactFragment | undefined
    displayShowMoreButton?: boolean
    onShowMore?: () => void
    showMoreClassName?: string
}

export const ConnectionNodesSummary: React.FunctionComponent<ConnectionNodesSummaryProps> = ({
    summary,
    displayShowMoreButton,
    showMoreClassName,
    onShowMore,
}) => (
    <>
        {summary}
        {displayShowMoreButton && (
            <button
                type="button"
                className={classNames('btn btn-sm btn-secondary filtered-connection__show-more', showMoreClassName)}
                onClick={onShowMore}
            >
                Show more
            </button>
        )}
    </>
)
