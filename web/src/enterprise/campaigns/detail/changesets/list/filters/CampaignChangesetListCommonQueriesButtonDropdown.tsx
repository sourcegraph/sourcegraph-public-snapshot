import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle, DropdownItem } from 'reactstrap'
import { Link } from 'react-router-dom'
import { QueryParameterProps } from '../../../../../../util/useQueryParameter'

interface Props extends Pick<QueryParameterProps, 'locationWithQuery'> {
    buttonClassName?: string
}

/**
 * A dropdown menu with common changeset list filter options.
 */
export const CampaignChangesetListCommonQueriesButtonDropdown: React.FunctionComponent<Props> = ({
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
                <DropdownItem header={true}>Common changeset filters</DropdownItem>
                <DropdownItem divider={true} />
                <DropdownItem tag={Link} to={locationWithQuery('')}>
                    All
                </DropdownItem>
                <DropdownItem tag={Link} to={locationWithQuery('is:open review:approved checks:passed')}>
                    Approved &amp; passing checks
                </DropdownItem>
                <DropdownItem tag={Link} to={locationWithQuery('review:pending')}>
                    Awaiting review
                </DropdownItem>
                <DropdownItem tag={Link} to={locationWithQuery('checks:failed')}>
                    Failing checks
                </DropdownItem>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
