import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { Status } from '../status'
import { CombinedStatusItem } from './CombinedStatusItem'

export interface CombinedStatusContext {
    itemClassName?: string
}

interface Props extends CombinedStatusContext, ExtensionsControllerProps, PlatformContextProps {
    statuses: Status[]

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
                <li key={i} className="list-group-item px-0">
                    <CombinedStatusItem
                        {...props}
                        key={JSON.stringify(status)}
                        status={status}
                        className={itemClassName}
                        headerClassName="pl-5"
                    />
                </li>
            ))}
        </ul>
    </div>
)
