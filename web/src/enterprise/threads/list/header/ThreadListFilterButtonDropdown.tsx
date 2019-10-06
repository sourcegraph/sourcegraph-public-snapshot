import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle, DropdownItem } from 'reactstrap'
import { QueryParameterProps } from '../../../../util/useQueryParameter'
import { Link } from 'react-router-dom'

interface Props extends Pick<QueryParameterProps, 'locationWithQuery'> {
    buttonClassName?: string
}

/**
 * A dropdown menu with common thread list filter options.
 */
export const ThreadListButtonDropdownFilter: React.FunctionComponent<Props> = ({
    locationWithQuery,
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])
    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen}>
            <DropdownToggle caret={true} className={buttonClassName}>
                Filter
            </DropdownToggle>
            <DropdownMenu>
                <DropdownItem tag={Link} to={locationWithQuery('is:open')}>
                    Open
                </DropdownItem>
                <DropdownItem tag={Link} to={locationWithQuery('is:closed')}>
                    Closed
                </DropdownItem>
                {/* TODO!(sqs): add more */}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
