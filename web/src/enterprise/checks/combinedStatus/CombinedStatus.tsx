import H from 'history'
import React from 'react'
import { CheckWithType } from '../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ChecksAreaContext } from '../list/ChecksArea'
import { CombinedStatusItem } from './CombinedStatusItem'

export interface CombinedStatusContext {
    itemClassName?: string
}

interface Props
    extends Pick<ChecksAreaContext, 'checksURL'>,
        CombinedStatusContext,
        ExtensionsControllerProps,
        PlatformContextProps {
    statuses: CheckWithType[]

    history: H.History
    location: H.Location
}

/**
 * The combined status, which summarizes and shows statuses from multiple status providers.
 */
export const CombinedStatus: React.FunctionComponent<Props> = ({ itemClassName, statuses, ...props }) => (
    <div className="combined-status">
        <ul className="list-group mb-0">
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
