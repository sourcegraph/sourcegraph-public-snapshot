import React from 'react'
import { WrappedStatus } from '../../../../../shared/src/api/client/services/statusService'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'
import { StatusStateIcon } from './StatusStateIcon'

interface Props {
    status: WrappedStatus
    to?: string
    tag?: 'header'
    className?: string
}

/**
 * The title and icon of a status.
 */
export const StatusHeader: React.FunctionComponent<Props> = ({
    status: { status },
    to,
    tag: Tag = 'header',
    className = '',
}) => (
    <Tag className={`d-flex align-items-center ${className}`}>
        <StatusStateIcon state={status.state} className="mr-3" />
        <h3 className="mb-0 font-weight-normal font-size-base">
            <LinkOrSpan to={to} className="stretched-link">
                {status.title}
            </LinkOrSpan>
        </h3>
    </Tag>
)
