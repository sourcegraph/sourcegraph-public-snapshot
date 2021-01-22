import MenuIcon from 'mdi-react/MenuIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'

interface MenuNavItemProps {
    children: React.ReactNode
}

export const MenuNavItem: React.FunctionComponent<MenuNavItemProps> = props => {
    const { children } = props
    const [isOpen, setIsOpen] = useState(() => false)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])

    return (
        <ButtonDropdown className="menu-nav-items" direction="down" isOpen={isOpen} toggle={toggleIsOpen}>
            <DropdownToggle caret={false} className="bg-transparent" nav={true}>
                <MenuIcon className="icon-inline" />
                {isOpen ? <MenuUpIcon className="icon-inline" /> : <MenuDownIcon className="icon-inline" />}
            </DropdownToggle>
            <DropdownMenu className="menu-nav-items__dropdown-menu">
                {React.Children.map(children, child => child && <DropdownItem>{child}</DropdownItem>)}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
