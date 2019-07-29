import React from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

interface Props {
    /** The button dropdown toggle label. */
    children: string

    /** The dropdown menu header text. */
    header?: string

    /** The dropdown items. */
    items: string[]
}

/**
 * A button dropdown with filter options shown in the list of threads.
 */
export const ThreadsListHeaderFilterButtonDropdown: React.FunctionComponent<Props> = ({ children, header, items }) => {
    const [isOpen, setIsOpen] = React.useState(false)

    return (
        /* tslint:disable-next-line:jsx-no-lambda */
        <ButtonDropdown isOpen={isOpen} toggle={() => setIsOpen(!isOpen)}>
            <DropdownToggle caret={true} className="bg-transparent border-0">
                {children}
            </DropdownToggle>
            <DropdownMenu>
                {header && (
                    <>
                        <DropdownItem header={true}>{header}</DropdownItem>
                        <DropdownItem divider={true} />
                    </>
                )}
                {items.map((item, i) => (
                    <DropdownItem key={i}>{item}</DropdownItem>
                ))}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
