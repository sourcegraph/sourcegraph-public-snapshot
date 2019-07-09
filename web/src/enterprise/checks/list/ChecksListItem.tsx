import H from 'history'
import QuestionMarkCircleOutlineIcon from 'mdi-react/QuestionMarkCircleOutlineIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { CheckInformationOrError } from '../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { CheckStateIcon } from '../components/CheckStateIcon'
import { ChecksAreaContext } from '../scope/ScopeChecksArea'
import { urlToCheck } from '../url'

interface Props extends Pick<ChecksAreaContext, 'checksURL'>, ExtensionsControllerProps, PlatformContextProps {
    checkInfoOrError: CheckInformationOrError

    tag: 'li'
    className?: string
    history: H.History
    location: H.Location
}

/**
 * A single check in a list.
 */
export const ChecksListItem: React.FunctionComponent<Props> = ({
    checkInfoOrError,
    tag: Tag,
    checksURL,
    className = '',
}) => (
    <Tag className={`d-flex flex-wrap align-items-stretch position-relative ${className}`}>
        {checkInfoOrError.information ? (
            <CheckStateIcon checkInfo={checkInfoOrError.information} className="mr-3" />
        ) : (
            <QuestionMarkCircleOutlineIcon className="mr-3 text-danger" />
        )}
        <h3 className="mb-0 font-weight-normal font-size-base d-flex align-items-center">
            <Link to={urlToCheck(checksURL, checkInfoOrError)} className="stretched-link">
                {checkInfoOrError.type} {checkInfoOrError.id}
            </Link>
        </h3>
    </Tag>
)
