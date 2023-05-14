import React from 'react'

import classNames from 'classnames'

import { Text } from '@sourcegraph/wildcard'

import styles from './ContextScopeComponents.module.scss'

interface EmptyStateProps {
    icon: string
    message: string
}

export const EmptyState: React.FC<EmptyStateProps> = ({ icon, message }) => {
    return (
        <div className={classNames('d-flex align-items-center justify-content-center', styles.emptyState)}>
            <svg height={40} width={40} viewBox="0 0 24 24">
                <path d={icon} fill="currentColor" />
            </svg>

            <div className={classNames('d-flex', styles.emptyStateContainer)}>
                <Text size="small" className={styles.marginBottomZero}>
                    {message}
                </Text>
            </div>
        </div>
    )
}
