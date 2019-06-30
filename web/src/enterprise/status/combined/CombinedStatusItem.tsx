import H from 'history'
import React from 'react'
import { WrappedStatus } from '../../../../../shared/src/api/client/services/statusService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { StatusHeader } from '../components/StatusHeader'
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
    status,
    tag: Tag,
    areaURL,
    className = '',
    headerClassName = '',
}) => (
    <Tag className={`d-flex flex-wrap align-items-stretch position-relative ${className}`}>
        <StatusHeader status={status} to={urlToStatus(areaURL, status.name)} tag="header" className={headerClassName} />
    </Tag>
)
