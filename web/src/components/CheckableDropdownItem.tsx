import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { DropdownItem } from 'reactstrap'

/**
 * A <DropdownItem /> wrapper that optionally shows a checkmark.
 */
export const CheckableDropdownItem: React.FunctionComponent<{ onClick: () => void; checked: boolean }> = ({
    onClick,
    checked,
    children,
}) => (
    <DropdownItem onClick={onClick}>
        <div className="d-flex align-items-start">
            <CheckIcon className={`icon-inline small mr-3 ${checked ? '' : 'hidden'}`} />
            <div>{children}</div>
        </div>
    </DropdownItem>
)
