import { MenuItem, MenuItems, MenuLink } from '@reach/menu-button'
import React from 'react'
import { Link } from 'react-router-dom'

interface Props {
    className?: string
}

export const GlobalMenu: React.FunctionComponent<Props> = ({ className = '' }) => (
    <menu className={`GlobalMenu ${className}`}>
        <div className="GlobalMenu__inner">
            Signed in as <strong>sqs</strong>
            <Link to="/users/sqs">Profile</Link>
            <Link to="/users/sqs">Settings</Link>
            <Link to="/users/sqs">Campaigns</Link>
            <Link to="/users/sqs">Extensions</Link>
            <Link to="/users/sqs">Profile</Link>
            <Link to="/users/sqs">Sign out</Link>
        </div>
    </menu>
)

export const GlobalMenuOLD: React.FunctionComponent<Props> = () => (
    <menu className="GlobalMenu">
        <MenuItems>
            <MenuLink as={Link} to="/users/sqs">
                Signed in as <strong>sqs</strong>
            </MenuLink>
        </MenuItems>
        <MenuItems>
            <MenuLink as={Link} to="/users/sqs">
                Profile
            </MenuLink>
            <MenuLink as={Link} to="/users/sqs">
                Settings
            </MenuLink>
            <MenuLink as={Link} to="/users/sqs">
                Campaigns
            </MenuLink>
            <MenuLink as={Link} to="/users/sqs">
                Extensions
            </MenuLink>
            <MenuLink as={Link} to="/users/sqs">
                Profile
            </MenuLink>
            <MenuItem onSelect={() => {}} disabled={true}>
                -
            </MenuItem>
            <MenuLink as={Link} to="/users/sqs">
                Sign out
            </MenuLink>
        </MenuItems>
    </menu>
)
