import React, { useCallback } from 'react'
import { DropdownItem } from 'reactstrap'

interface Props {
    value: string
    onChange: (value: string) => void
    placeholder?: string

    header: React.ReactFragment
}

export const DropdownMenuFilter: React.FunctionComponent<Props> = ({
    value,
    onChange: parentOnChange,
    placeholder,
    header,
}) => {
    const onChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            parentOnChange(e.currentTarget.value)
        },
        [parentOnChange]
    )
    return (
        <>
            {header && <DropdownItem header={true}>{header}</DropdownItem>}
            <DropdownItem
                tag="div"
                role="search"
                header={true}
                /* TODO!(sqs): make Esc key when focused close the dropdown */
            >
                <input
                    type="search"
                    className="form-control small py-2"
                    value={value}
                    placeholder={placeholder}
                    onChange={onChange}
                    autoFocus={true}
                />
            </DropdownItem>
            <DropdownItem divider={true} />
        </>
    )
}
