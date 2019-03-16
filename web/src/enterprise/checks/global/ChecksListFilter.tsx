import React, { useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

interface Props {
    /** The current value of the filter. */
    value: string

    className: string
}

/**
 * The filter control (dropdown and input field) for a list of checks.
 */
export const ChecksListFilter: React.FunctionComponent<Props> = ({ value, className }) => {
    const [isOpen, setIsOpen] = useState(false)

    return (
        <div className={`input-group ${className}`}>
            <div className="input-group-prepend">
                {/* tslint:disable-next-line:jsx-no-lambda */}
                <ButtonDropdown isOpen={isOpen} toggle={() => setIsOpen(!isOpen)}>
                    <DropdownToggle caret={true}>Filter</DropdownToggle>
                    <DropdownMenu>
                        <DropdownItem>Open</DropdownItem>
                        <DropdownItem>Assigned to you</DropdownItem>
                        <DropdownItem>Acted on by you</DropdownItem>
                        <DropdownItem>Closed</DropdownItem>
                    </DropdownMenu>
                </ButtonDropdown>
            </div>
            <input
                type="text"
                className="form-control"
                aria-label="Filter checks"
                autoCapitalize="off"
                value={value}
                // tslint:disable-next-line:jsx-no-lambda
                onChange={() => void 0}
            />
        </div>
    )
}
