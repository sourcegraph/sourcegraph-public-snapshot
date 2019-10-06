import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { DropdownItem } from 'reactstrap'

/**
 * A <DropdownItem /> wrapper that optionally shows a checkmark.
 */
export const CheckableDropdownItem: React.FunctionComponent<{
    onClick: () => void
    checked: boolean
    className?: string
}> = ({ onClick, checked, className, children }) => (
    <DropdownItem onClick={onClick} className={className}>
        <div className="d-flex align-items-start">
            <CheckIcon className={`icon-inline flex-0 small mr-3 ${checked ? '' : 'hidden invisible'}`} />
            <div>{children}</div>
        </div>
    </DropdownItem>
)
