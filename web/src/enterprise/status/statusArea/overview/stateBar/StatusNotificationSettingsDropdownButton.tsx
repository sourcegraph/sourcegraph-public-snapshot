import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { MuteIcon } from '../../../../../util/octicons'

interface Props {
    className?: string
    buttonClassName?: string
}

export const StatusNotificationSettingsDropdownButton: React.FunctionComponent<Props> = ({
    className = '',
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle className={`btn ${buttonClassName}`} color="none">
                <MuteIcon className="icon-inline mr-2" /> Unsubscribe
            </DropdownToggle>
            <DropdownMenu>
                <DropdownItem>All changes</DropdownItem>
                <DropdownItem>Failures only</DropdownItem>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
