import H from 'history'
import React from 'react'
import { CheckInformationWithID } from '../../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { ChecksAreaContext } from '../ScopeChecksArea'
import { CombinedCheckItem } from './CombinedStatusItem'

export interface CombinedStatusContext {
    itemClassName?: string
}

interface Props
    extends Pick<ChecksAreaContext, 'checksURL'>,
        CombinedStatusContext,
        ExtensionsControllerProps,
        PlatformContextProps {
    statuses: CheckInformationWithID[]

    history: H.History
    location: H.Location
}

/**
 * The combined check, which summarizes and shows information from multiple check providers.
 */
export const CombinedStatus: React.FunctionComponent<Props> = ({ itemClassName, statuses, ...props }) => (
    <div className="combined-check">
        <ul className="list-group mb-0">
            {statuses.map((check, i) => (
                <CombinedCheckItem
                    {...props}
                    key={`${check.type}/${check.id}`}
                    tag="li"
                    check={check}
                    className={`list-group-item list-group-item-action ${itemClassName}`}
                />
            ))}
        </ul>
    </div>
)
