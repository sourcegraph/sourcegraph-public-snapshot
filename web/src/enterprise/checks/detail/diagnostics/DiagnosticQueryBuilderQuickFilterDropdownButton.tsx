import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

interface Props {
    className?: string
    buttonClassName?: string
}

/**
 * A dropdown menu with common diagnostic query options.
 *
 * TODO!(sqs): the contents are dummy
 */
export const DiagnosticQueryBuilderQuickFilterDropdownButton: React.FunctionComponent<Props> = ({
    className = '',
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])
    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle caret={true} className={buttonClassName} />
            <DropdownMenu>
                <DropdownItem>Open</DropdownItem>
                <DropdownItem>Assigned to you</DropdownItem>
                <DropdownItem>Acted on by you</DropdownItem>
                <DropdownItem>Closed</DropdownItem>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
