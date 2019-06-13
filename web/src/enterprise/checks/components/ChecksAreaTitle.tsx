import React from 'react'
import { ChecksIcon } from '../icons'

interface Props {
    className?: string
    primaryActions?: JSX.Element | null
}

/**
 * The checks area title.
 */
export const ChecksAreaTitle: React.FunctionComponent<Props> = ({ className = '', primaryActions, children }) => (
    <div className="d-flex align-items-center mb-3">
        <h1 className={`h3 mb-0 d-flex align-items-center ${className}`}>
            <ChecksIcon className="icon-inline mr-1" /> Checks
        </h1>
        {children}
        {primaryActions && (
            <>
                <div className="flex-1" />
                {primaryActions}
            </>
        )}
    </div>
)
