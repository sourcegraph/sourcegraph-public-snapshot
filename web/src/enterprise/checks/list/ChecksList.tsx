import H from 'history'
import React from 'react'
import { CheckInformationOrError } from '../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ChecksAreaContext } from '../scope/ScopeChecksArea'
import { CombinedCheckItem } from './ChecksListItem'

export interface CombinedStatusContext {
    itemClassName?: string
}

interface Props
    extends Pick<ChecksAreaContext, 'checksURL'>,
        CombinedStatusContext,
        ExtensionsControllerProps,
        PlatformContextProps {
    checks: CheckInformationOrError[]

    history: H.History
    location: H.Location
}

/**
 * A list of checks.
 */
export const ChecksList: React.FunctionComponent<Props> = ({ itemClassName, checks, ...props }) => (
    <div className="checks-list">
        <ul className="list-group mb-0">
            {checks.map(check => (
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
