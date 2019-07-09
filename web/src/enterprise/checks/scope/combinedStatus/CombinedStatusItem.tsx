import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { CheckInformationWithID } from '../../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { CheckStateIcon } from '../../components/CheckStateIcon'
import { urlToCheck } from '../../url'
import { ChecksAreaContext } from '../ScopeChecksArea'

interface Props extends Pick<ChecksAreaContext, 'checksURL'>, ExtensionsControllerProps, PlatformContextProps {
    check: CheckInformationWithID

    tag: 'li'
    className?: string
    history: H.History
    location: H.Location
}

/**
 * A single check in a combined check ({@link CombinedCheck}).
 */
export const CombinedCheckItem: React.FunctionComponent<Props> = ({ check, tag: Tag, checksURL, className = '' }) => (
    <Tag className={`d-flex flex-wrap align-items-stretch position-relative ${className}`}>
        <CheckStateIcon check={check} className="mr-3" />
        <h3 className="mb-0 font-weight-normal font-size-base">
            <Link to={urlToCheck(checksURL, name)} className="stretched-link">
                {check.type} {check.id}
            </Link>
        </h3>
    </Tag>
)
