import React, { useState, useCallback } from 'react'
import { ButtonDropdown, DropdownToggle, DropdownMenu, DropdownItem } from 'reactstrap'
import PlusIcon from 'mdi-react/PlusIcon'
import { RULE_ACTION_TYPES } from '.'

interface Props {
    items: typeof RULE_ACTION_TYPES
    onSelect: (ruleActionInitialValue: { type: string } & object) => void
    className?: string
}

/**
 * A dropdown button with a menu to add a new action to a rule.
 */
export const AddRuleActionDropdownButton: React.FunctionComponent<Props> = ({ items, onSelect, className = '' }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen}>
            <DropdownToggle color="" className={className}>
                <PlusIcon className="icon-inline" /> Action
            </DropdownToggle>
            <DropdownMenu>
                {items.map(({ id, title, icon: Icon, initialValue }) => (
                    // eslint-disable-next-line react/jsx-no-bind
                    <DropdownItem key={id} onClick={() => onSelect(initialValue)} className="d-flex align-items-center">
                        {Icon && <Icon className="icon-inline mr-1" />}
                        {title}
                    </DropdownItem>
                ))}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
