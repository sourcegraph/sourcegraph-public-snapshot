import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { UnmuteIcon } from '../../../../../util/octicons'

interface Props {
    className?: string
    buttonClassName?: string
}

// TODO!(sqs) use this
export const CheckNotificationSettingsDropdownButton: React.FunctionComponent<Props> = ({
    className = '',
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle className={`btn ${buttonClassName}`} color="none">
                <UnmuteIcon className="icon-inline mr-2" /> Subscribe
            </DropdownToggle>
            <DropdownMenu>
                <DropdownItem>All changes</DropdownItem>
                <DropdownItem>Failures only</DropdownItem>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
