import React, { useCallback, useMemo, useRef, useState } from 'react'
import Popover, { positionDefault } from '@reach/popover'
import { BrandLogo } from '../../../components/branding/BrandLogo'
import { ThemeProps } from '../../../../../shared/src/theme'
import { UserNavItem } from '../../UserNavItem'
import { Menu, MenuButton, MenuItem, MenuItems, MenuLink, MenuPopover } from '@reach/menu-button'
import { Link } from 'react-router-dom'
import { GlobalMenu } from './GlobalMenu'

interface Props extends ThemeProps {
    logoClassName?: string

    setShowGlobalMenu: (value: boolean) => void
}

export const GlobalMenuPopoverButton: React.FunctionComponent<Props> = ({
    logoClassName = '',
    setShowGlobalMenu,
    isLightTheme,
}) => (
    <Menu>
        <MenuButton
            className="btn btn-link text-muted d-flex align-items-center"
            onMouseEnter={() => setShowGlobalMenu(true)}
            // onMouseLeave={() => setShowGlobalMenu(false)}
        >
            <BrandLogo
                branding={window.context.branding}
                isLightTheme={isLightTheme}
                variant="symbol"
                className={logoClassName}
            />{' '}
            {/* <strong>sqs</strong> */}
            <span aria-hidden={true}>▾</span>
        </MenuButton>
        <MenuPopover className="popover bg-white ml-2" portal={false}>
            <GlobalMenu />
        </MenuPopover>
    </Menu>
)

// TODO(sqs): open the menu on page load
setInterval(() => {
    for (const e of document.querySelectorAll('[data-reach-menu-popover]')) {
        // e.removeAttribute('hidden')
    }
}, 200)
