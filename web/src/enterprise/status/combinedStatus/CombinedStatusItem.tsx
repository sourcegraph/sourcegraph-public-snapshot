import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { WrappedStatus } from '../../../../../shared/src/api/client/services/statusService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { StatusStateIcon } from '../components/StatusStateIcon'
import { StatusesAreaContext } from '../statusesArea/StatusesArea'
import { urlToStatus } from '../url'

interface Props extends Pick<StatusesAreaContext, 'statusesURL'>, ExtensionsControllerProps, PlatformContextProps {
    status: WrappedStatus

    tag: 'li'
    className?: string
    history: H.History
    location: H.Location
}

/**
 * A single status in a combined status ({@link CombinedStatus}).
 */
export const CombinedStatusItem: React.FunctionComponent<Props> = ({
    status: { name, status },
    tag: Tag,
    statusesURL,
    className = '',
}) => (
    <Tag className={`d-flex flex-wrap align-items-stretch position-relative ${className}`}>
        <StatusStateIcon status={status} className="mr-3" />
        <h3 className="mb-0 font-weight-normal font-size-base">
            <Link to={urlToStatus(statusesURL, name)} className="stretched-link">
                {status.title}
            </Link>
        </h3>
    </Tag>
)
