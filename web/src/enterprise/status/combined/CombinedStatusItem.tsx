import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { Status } from '../status'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    status: Status

    className?: string
    headerClassName?: string
    history: H.History
    location: H.Location
}

/**
 * A single status in a combined status ({@link CombinedStatus}).
 */
export const CombinedStatusItem: React.FunctionComponent<Props> = ({
    status,
    className = '',
    headerClassName = '',
}) => (
    <div className={`d-flex flex-wrap align-items-stretch ${className}`}>
        <header className={headerClassName}>
            <h3 className="mb-0 font-weight-normal font-size-base">{status.title}</h3>
        </header>
    </div>
)
