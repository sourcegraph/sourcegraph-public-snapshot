import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { DropdownItem, DropdownItemProps } from 'reactstrap'

/**
 * A <DropdownItem /> wrapper that optionally shows a checkmark.
 */
export const CheckableDropdownItem: React.FunctionComponent<
    {
        checked: boolean
    } & DropdownItemProps
> = ({ checked, children, ...props }) => (
    <DropdownItem {...props}>
        <div className="d-flex align-items-start">
            <CheckIcon className={`icon-inline flex-0 small mr-3 ${checked ? '' : 'hidden invisible'}`} />
            <div>{children}</div>
        </div>
    </DropdownItem>
)
