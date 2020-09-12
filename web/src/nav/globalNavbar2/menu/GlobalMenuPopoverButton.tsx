import React, { useCallback, useMemo, useRef, useState } from 'react'
import Popover, { positionDefault } from '@reach/popover'
import { BrandLogo } from '../../../components/branding/BrandLogo'
import { ThemeProps } from '../../../../../shared/src/theme'
import { UserNavItem } from '../../UserNavItem'
import { Menu, MenuButton, MenuItem, MenuItems, MenuLink, MenuPopover } from '@reach/menu-button'
import { Link } from 'react-router-dom'

interface Props extends ThemeProps {
    logoClassName?: string
}

export const GlobalMenuPopoverButton: React.FunctionComponent<Props> = ({ logoClassName = '', isLightTheme }) => (
    <Menu>
        <MenuButton className="btn btn-link">
            <BrandLogo
                branding={window.context.branding}
                isLightTheme={isLightTheme}
                variant="symbol"
                className={logoClassName}
            />{' '}
            <strong>sqs</strong>
            <span aria-hidden={true}>▾</span>
        </MenuButton>
        <MenuPopover className="popover bg-white ml-2">
            <MenuItems className="list-group list-group-flush">
                <MenuLink as={Link} to="/users/sqs" className="list-group-item list-group-item-action">
                    Profile
                </MenuLink>
                <MenuLink as={Link} to="/users/sqs" className="list-group-item list-group-item-action">
                    Settings
                </MenuLink>
                <MenuLink as={Link} to="/users/sqs" className="list-group-item list-group-item-action">
                    Campaigns
                </MenuLink>
                <MenuLink as={Link} to="/users/sqs" className="list-group-item list-group-item-action">
                    Extensions
                </MenuLink>
                <MenuLink as={Link} to="/users/sqs" className="list-group-item list-group-item-action">
                    Profile
                </MenuLink>
                <MenuItem disabled={true}>-</MenuItem>
                <MenuLink as={Link} to="/users/sqs" className="list-group-item list-group-item-action">
                    Sign out
                </MenuLink>
            </MenuItems>
        </MenuPopover>
    </Menu>
)
