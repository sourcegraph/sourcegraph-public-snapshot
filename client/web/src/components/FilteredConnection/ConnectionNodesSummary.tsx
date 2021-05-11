import classNames from 'classnames'
import * as React from 'react'

import styles from './ConnectionNodesSummary.module.scss'

interface ListSummaryProps {
    summary: React.ReactFragment | undefined
    displayShowMoreButton?: boolean
    onShowMore?: () => void
    showMoreClassName?: string
}

export const ConnectionNodesSummary: React.FunctionComponent<ListSummaryProps> = ({
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
                className={classNames('btn btn-sm btn-secondary', styles.summaryShowMore, showMoreClassName)}
                onClick={onShowMore}
            >
                Show more
            </button>
        )}
    </>
)
