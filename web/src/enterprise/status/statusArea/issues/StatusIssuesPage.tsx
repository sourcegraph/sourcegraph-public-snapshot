import H from 'history'
import React from 'react'
import { StatusAreaContext } from '../StatusArea'

interface Props extends Pick<StatusAreaContext, 'status'> {
    className?: string
    history: H.History
    location: H.Location
}

/**
 * The status issues page.
 */
export const StatusIssuesPage: React.FunctionComponent<Props> = ({ status, className = '', ...props }) => (
    <div className={`status-issues-page ${className}`}>ISSUES</div>
)
