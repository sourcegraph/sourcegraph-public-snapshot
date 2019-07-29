import React from 'react'

/**
 * The overview for a single thread's actions.
 */
export const ThreadActionsOverview: React.FunctionComponent<{ className?: string }> = ({ className = '' }) => (
    <div className={`thread-actions-overview text-muted ${className}`}>Select an action...</div>
)
