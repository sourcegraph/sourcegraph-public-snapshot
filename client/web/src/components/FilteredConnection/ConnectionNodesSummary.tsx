import classNames from 'classnames'
import * as React from 'react'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'
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
}) => {
    const [isRedesignEnabled] = useRedesignToggle()

    const showMoreButton = displayShowMoreButton && (
        <button
            type="button"
            className={classNames(
                'btn btn-sm',
                isRedesignEnabled ? 'btn-link' : 'btn-secondary',
                styles.summaryShowMore,
                showMoreClassName
            )}
            onClick={onShowMore}
        >
            Show more
        </button>
    )

    if (isRedesignEnabled) {
        return (
            <div className={styles.summary}>
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
