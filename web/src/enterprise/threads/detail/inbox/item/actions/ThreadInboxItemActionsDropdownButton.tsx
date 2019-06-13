import React, { useState, useCallback } from 'react'
import { CodeAction } from 'sourcegraph'
import { ButtonDropdown, DropdownToggle, DropdownMenu, DropdownItem } from 'reactstrap'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'

interface Props {
    codeActions: CodeAction[]
    className?: string
    buttonClassName?: string
}

export const ThreadInboxItemActionsDropdownButton: React.FunctionComponent<Props> = ({
    codeActions,
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
                {codeActions.map((codeAction, i) => (
                    <DropdownItem key={i} className="d-flex">
                        {codeAction.title}
                    </DropdownItem>
                ))}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
