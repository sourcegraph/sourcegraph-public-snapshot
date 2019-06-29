import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { WrappedStatus } from '../../../../../shared/src/api/client/services/statusService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { StatusStateIcon } from '../components/StatusStateIcon'
import { urlToStatus } from '../url'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    status: WrappedStatus

    tag: 'li'
    className?: string
    headerClassName?: string
    areaURL: string
    history: H.History
    location: H.Location
}

/**
 * A single status in a combined status ({@link CombinedStatus}).
 */
export const CombinedStatusItem: React.FunctionComponent<Props> = ({
    status: { name, status },
    tag: Tag,
    areaURL,
    className = '',
    headerClassName = '',
}) => (
    <Tag className={`d-flex flex-wrap align-items-stretch position-relative ${className}`}>
        <header className={`d-flex align-items-center ${headerClassName}`}>
            <StatusStateIcon state={status.state} className="icon-inline mr-3" />
            <h3 className="mb-0 font-weight-normal font-size-base">
                <Link to={urlToStatus(areaURL, name)} className="text-body stretched-link">
                    {status.title}
                </Link>
            </h3>
        </header>
    </Tag>
)
