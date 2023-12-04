import React from 'react'

import classNames from 'classnames'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import styles from './ConnectionLoading.module.scss'

interface ConnectionLoadingProps {
    className?: string
    compact?: boolean
}

/**
 * FilteredConnection styled loading spinner
 */
export const ConnectionLoading: React.FunctionComponent<React.PropsWithChildren<ConnectionLoadingProps>> = ({
    className,
    compact,
}) => (
    <span
        data-testid="filtered-connection-loader"
        className={classNames(compact && styles.compact, styles.normal, className)}
    >
        <LoadingSpinner />
    </span>
)
