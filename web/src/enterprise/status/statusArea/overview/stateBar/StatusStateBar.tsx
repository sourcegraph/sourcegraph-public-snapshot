import React from 'react'
import { StatusStateIcon } from '../../../components/StatusStateIcon'
import { themeColorForStatus } from '../../../util'
import { StatusAreaContext } from '../../StatusArea'

interface Props extends Pick<StatusAreaContext, 'status'> {
    className?: string
}

/**
 * A bar that displays the state of a status.
 */
export const StatusStateBar: React.FunctionComponent<Props> = ({ status, className = '' }) => (
    <div className={`d-flex align-items-center border border-${themeColorForStatus(status.status)} ${className}`}>
        <StatusStateIcon status={status.status} className="icon-inline mr-3" />
        {status.status.state.message && (
            <span className={`text-${themeColorForStatus(status.status)}`}>{status.status.state.message}</span>
        )}
        <div className="flex-1" />
    </div>
)
