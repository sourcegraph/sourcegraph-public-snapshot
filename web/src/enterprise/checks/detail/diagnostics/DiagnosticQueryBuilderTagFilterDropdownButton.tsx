import H from 'history'
import React, { useCallback, useState } from 'react'
import { Link } from 'react-router-dom'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { DropdownMenuFilter } from '../../../../components/dropdownMenuFilter/DropdownMenuFilter'

interface Item {
    text: string
    url: H.LocationDescriptor
    count?: number
}

interface Props {
    items: Item[]
    pluralNoun: string
    buttonText: string
    headerText: string
    queryPlaceholderText: string

    className?: string
    buttonClassName?: string
}

/**
 * A dropdown menu for filtering diagnostics.
 */
export const DiagnosticQueryBuilderFilterDropdownButton: React.FunctionComponent<Props> = ({
    items,
    pluralNoun,
    buttonText,
    headerText,
    queryPlaceholderText,
    className = '',
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const [filter, setFilter] = useState('')
    const itemsFiltered = items.filter(({ text }) => text.toLowerCase().includes(filter.toLowerCase()))

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle color="" caret={true} className={buttonClassName}>
                {buttonText}
            </DropdownToggle>
            <DropdownMenu /* TODO!(sqs) set width to avoid changing width when filter changes */>
                <DropdownMenuFilter
                    value={filter}
                    onChange={setFilter}
                    placeholder={queryPlaceholderText}
                    header={headerText}
                />
                {itemsFiltered.length > 0 ? (
                    itemsFiltered.map(({ text, url, count }) => (
                        <Link
                            key={text}
                            to={url}
                            className="dropdown-item d-flex align-items-center justify-content-between"
                            onClick={toggleIsOpen}
                        >
                            <span className="text-truncate">{text}</span>{' '}
                            {typeof count === 'number' && <span className="ml-3 badge badge-secondary">{count}</span>}
                        </Link>
                    ))
                ) : (
                    <DropdownItem header={true}>No {pluralNoun} found</DropdownItem>
                )}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
