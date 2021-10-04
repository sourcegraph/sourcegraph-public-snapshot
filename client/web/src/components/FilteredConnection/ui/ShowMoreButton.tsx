import classNames from 'classnames'
import React from 'react'

import styles from './ShowMoreButton.module.scss'

interface ShowMoreProps {
    className?: string
    compact?: boolean
    centered?: boolean
    onClick: () => void
}

/**
 * FilteredConnection styled Button to support fetching more results
 */
export const ShowMoreButton: React.FunctionComponent<ShowMoreProps> = ({ className, compact, centered, onClick }) => (
    <button
        type="button"
        className={classNames(
            'btn btn-sm',
            styles.normal,
            !compact && styles.noncompact,
            centered && styles.centered,
            'btn-link',
            className
        )}
        onClick={onClick}
    >
        Show more
    </button>
)
