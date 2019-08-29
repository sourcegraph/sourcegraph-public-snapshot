import React, { useCallback, useState } from 'react'
import { Dropdown, DropdownToggle } from 'reactstrap'
import { LinkWithIconOnlyTooltip } from '../../../../components/LinkWithIconOnlyTooltip'
import { CampaignsNavItemDropdownMenu } from './CampaignsNavItemDropdownMenu'
import { CampaignsIcon } from '../../icons'

interface Props {
    className?: string
}

/**
 * An item in {@link GlobalNavbar} that links to the changesets area.
 */
export const CampaignsNavItem: React.FunctionComponent<Props> = ({ className = '' }) => {
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
        >
            <DropdownToggle tag="span" data-toggle="dropdown" aria-expanded={isOpen} onMouseEnter={setIsOpenTrue}>
                <LinkWithIconOnlyTooltip
                    to="/campaigns"
                    text="Campaigns"
                    icon={CampaignsIcon}
                    className={`nav-link btn btn-link px-3 text-decoration-none ${className}`}
                />
            </DropdownToggle>
            <CampaignsNavItemDropdownMenu className="mt-0" />
        </Dropdown>
    )
}
