import React, { useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Form } from '../../../../components/Form'

interface Props {
    /** The current value of the filter. */
    value: string

    /** Called when the value changes. */
    onChange: (value: string) => void

    className?: string
}

/**
 * The filter control (dropdown and input field) for a list of thread inbox items.
 */
// tslint:disable: jsx-no-lambda
export const ThreadInboxItemsListFilter: React.FunctionComponent<Props> = ({ value, onChange, className }) => {
    const [uncommittedValue, setUncommittedValue] = useState(value)
    const [isOpen, setIsOpen] = useState(false)

    return (
        <Form
            className={`form ${className}`}
            onSubmit={e => {
                e.preventDefault()
                onChange(uncommittedValue)
            }}
        >
            <div className="input-group">
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
                    aria-label="Filter threads"
                    autoCapitalize="off"
                    value={uncommittedValue}
                    onChange={e => setUncommittedValue(e.currentTarget.value)}
                />
            </div>
        </Form>
    )
}
