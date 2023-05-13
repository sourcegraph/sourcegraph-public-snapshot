import React from 'react'

import { Text } from '@sourcegraph/wildcard'

import styles from './ContextScopeComponents.module.scss'

interface EmptyStateProps {
    icon: string
    message: string
}

export const EmptyState: React.FC<EmptyStateProps> = ({ icon, message }) => {
    return (
        <div className={styles.emptyState}>
            <svg height={40} width={40} viewBox="0 0 24 24">
                <path d={icon} fill="currentColor" />
            </svg>

            <div className={styles.emptyStateContainer}>
                <Text size="small" style={{ marginBottom: 0 }}>
                    {message}
                </Text>
            </div>
        </div>
    )
}
