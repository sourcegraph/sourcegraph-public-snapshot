import classNames from 'classnames'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

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
    <Button
        className={classNames(styles.normal, !compact && styles.noncompact, centered && styles.centered, className)}
        onClick={onClick}
        size="sm"
        variant="link"
    >
        Show more
    </Button>
)
