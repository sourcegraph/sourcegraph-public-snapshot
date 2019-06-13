import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Action } from '../../../../../shared/src/api/types/action'

interface Props {
    actions: Action[]
    className?: string
    buttonClassName?: string
}

/**
 * A button that toggles a dropdown with actions.
 */
export const ActionsDropdownButton: React.FunctionComponent<Props> = ({
    actions,
    className = '',
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle className={`btn ${buttonClassName}`} color="none">
                <DotsHorizontalIcon className="icon-inline" />
            </DropdownToggle>
            <DropdownMenu>
                {actions.map((action, i) => (
                    // TODO!(sqs): add onClick
                    <DropdownItem key={i} className="d-flex">
                        {action.title}
                    </DropdownItem>
                ))}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
