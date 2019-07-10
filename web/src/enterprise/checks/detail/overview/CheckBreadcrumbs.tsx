import React from 'react'
import { Link } from 'react-router-dom'
import { CheckStateIcon } from '../../components/CheckStateIcon'
import { CheckAreaContext } from '../CheckArea'

interface Props extends Pick<CheckAreaContext, 'checkID' | 'checkInfo' | 'checkURL' | 'checksURL'> {
    className?: string
}

/**
 * The breadcrumbs for a check.
 */
export const CheckBreadcrumbs: React.FunctionComponent<Props> = ({
    checkID,
    checkInfo,
    checkURL,
    checksURL,
    className = '',
}) => (
    <nav className={`d-flex align-items-center ${className}`} aria-label="breadcrumb">
        <ol className="breadcrumb">
            <li className="breadcrumb-item">
                <Link to={checksURL}>Checks</Link>
            </li>
            <li className="breadcrumb-item active font-weight-bold">
                <Link to={checkURL} className="d-inline-flex align-items-center">
                    {checkID.type} <CheckStateIcon checkInfo={checkInfo} className="icon-inline ml-2" />
                </Link>
            </li>
        </ol>
    </nav>
)
