import React, { useCallback, useState } from 'react'
import { Dropdown, DropdownToggle } from 'reactstrap'
import { LinkWithIconOnlyTooltip } from '../../../../components/LinkWithIconOnlyTooltip'
import { GitPullRequestIcon } from '../../../../util/octicons'
import { ChecksNavItemDropdownMenu } from './ChangesetsNavItemDropdownMenu'

interface Props {
    className?: string
}

/**
 * An item in {@link GlobalNavbar} that links to the changesets area.
 */
export const ChangesetsNavItem: React.FunctionComponent<Props> = ({ className = '' }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])
    const setIsOpenTrue = useCallback(() => setIsOpen(true), [])
    const setIsOpenFalse = useCallback(() => setIsOpen(false), [])
    return (
        <Dropdown
            isOpen={isOpen}
            toggle={toggleIsOpen}
            onMouseLeave={setIsOpenFalse}
            onClick={setIsOpenFalse}
            inNavbar={true}
            direction="down"
        >
            <DropdownToggle tag="span" data-toggle="dropdown" aria-expanded={isOpen} onMouseEnter={setIsOpenTrue}>
                <LinkWithIconOnlyTooltip
                    to="/changesets"
                    text="Changesets"
                    icon={GitPullRequestIcon}
                    className={`nav-link btn btn-link text-decoration-none ${className}`}
                />
            </DropdownToggle>
            <ChecksNavItemDropdownMenu className="mt-0 threads-nav-item__dropdown-menu" />
        </Dropdown>
    )
}
