import H from 'history'
import React from 'react'
import { CheckInformationOrError } from '../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ChecksAreaContext } from '../scope/ScopeChecksArea'
import { ChecksListItem } from './ChecksListItem'

export interface CombinedStatusContext {
    itemClassName?: string
}

interface Props
    extends Pick<ChecksAreaContext, 'checksURL'>,
        CombinedStatusContext,
        ExtensionsControllerProps,
        PlatformContextProps {
    checksInfoOrError: CheckInformationOrError[]

    history: H.History
    location: H.Location
}

/**
 * A list of checks.
 */
export const ChecksList: React.FunctionComponent<Props> = ({ itemClassName, checksInfoOrError, ...props }) => (
    <div className="checks-list">
        <ul className="list-group mb-0">
            {checksInfoOrError.map(checkInfoOrError => (
                <ChecksListItem
                    {...props}
                    key={`${checkInfoOrError.type}/${checkInfoOrError.id}`}
                    tag="li"
                    checkInfoOrError={checkInfoOrError}
                    className={`list-group-item list-group-item-action ${itemClassName}`}
                />
            ))}
        </ul>
    </div>
)
