import React, { useCallback, useState } from 'react'
import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { LinkWithIconOnlyTooltip } from '../../../../components/LinkWithIconOnlyTooltip'
import { ChecksIcon } from '../../icons'
import { ChecksNavItemDropdownMenu } from './ChecksNavItemDropdownMenu'

interface Props {
    className?: string
}

/**
 * An item in {@link GlobalNavbar} that links to the checks area.
 */
export const ChecksNavItem: React.FunctionComponent<Props> = ({ className = '' }) => {
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
                    to="/checks"
                    text="Checks"
                    icon={ChecksIcon}
                    className={`nav-link btn btn-link text-decoration-none ${className}`}
                />
            </DropdownToggle>
            <ChecksNavItemDropdownMenu className="mt-0 threads-nav-item__dropdown-menu" />
        </Dropdown>
    )
}
