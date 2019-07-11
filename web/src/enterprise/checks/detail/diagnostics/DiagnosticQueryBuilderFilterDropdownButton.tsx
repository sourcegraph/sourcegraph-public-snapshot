import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

/**
 * A dropdown menu with common diagnostic query options.
 *
 * TODO!(sqs): the contents are dummy
 */
export const DiagnosticQueryBuilderFilterDropdownButton: React.FunctionComponent = () => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])
    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen}>
            <DropdownToggle caret={true}>Filter</DropdownToggle>
            <DropdownMenu>
                <DropdownItem>Open</DropdownItem>
                <DropdownItem>Assigned to you</DropdownItem>
                <DropdownItem>Acted on by you</DropdownItem>
                <DropdownItem>Closed</DropdownItem>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
