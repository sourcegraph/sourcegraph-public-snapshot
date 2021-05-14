import classNames from 'classnames'
import * as React from 'react'

interface ConnectionNodesSummaryShowMoreProps {
    onShowMore?: () => void
    showMoreClassName?: string
}

export const ConnectionNodesSummaryShowMore: React.FunctionComponent<ConnectionNodesSummaryShowMoreProps> = ({
    onShowMore,
    showMoreClassName,
}) => (
    <button
        type="button"
        className={classNames('btn btn-sm filtered-connection__show-more btn-secondary', showMoreClassName)}
        onClick={onShowMore}
    >
        Show more
    </button>
)
