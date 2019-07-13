import CheckAllIcon from 'mdi-react/CheckAllIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

export interface CreateOrPreviewChangesetButtonProps {
    disabled?: boolean
    className?: string
    buttonClassName?: string
}

/**
 * A button dropdown for performing batch actions ondiagnostics.
 */
export const DiagnosticsBatchActionsButtonDropdown: React.FunctionComponent<CreateOrPreviewChangesetButtonProps> = ({
    disabled,
    className = '',
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    // TODO!(sqs): actually compute what the actions are instead of hardcoding

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle color="" className={buttonClassName} caret={true} disabled={disabled}>
                <CheckAllIcon className="icon-inline mr-2" /> Batch actions
            </DropdownToggle>
            <DropdownMenu>
                <DropdownItem header={true} className="py-1">
                    Apply to all diagnostics with no selected action
                </DropdownItem>
                <DropdownItem divider={true} className="py-1" />
                <DropdownItem>
                    Fix <small className="text-muted">31 diagnostics</small>
                </DropdownItem>
                <DropdownItem>
                    Disable rule <small className="text-muted">39 diagnostics</small>
                </DropdownItem>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
