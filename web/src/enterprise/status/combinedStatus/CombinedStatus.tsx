import H from 'history'
import React from 'react'
import { WrappedStatus } from '../../../../../shared/src/api/client/services/statusService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { StatusesAreaContext } from '../statusesArea/StatusesArea'
import { CombinedStatusItem } from './CombinedStatusItem'

export interface CombinedStatusContext {
    itemClassName?: string
}

interface Props
    extends Pick<StatusesAreaContext, 'statusesURL'>,
        CombinedStatusContext,
        ExtensionsControllerProps,
        PlatformContextProps {
    statuses: WrappedStatus[]

    history: H.History
    location: H.Location
}

/**
 * The combined status, which summarizes and shows statuses from multiple status providers.
 */
export const CombinedStatus: React.FunctionComponent<Props> = ({ itemClassName, statuses, ...props }) => (
    <div className="combined-status">
        <ul className="list-group list-group-flush mb-0">
            {statuses.map((status, i) => (
                <CombinedStatusItem
                    {...props}
                    key={status.name}
                    tag="li"
                    status={status}
                    className={`list-group-item list-group-item-action ${itemClassName}`}
                />
            ))}
        </ul>
    </div>
)
