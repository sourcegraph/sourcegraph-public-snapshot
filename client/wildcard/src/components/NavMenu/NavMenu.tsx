import React, { forwardRef, ReactNode } from 'react'

import * as H from 'history'
import { MdiReactIconProps } from 'mdi-react'

import { ForwardReferenceComponent, Position } from '../..'
import { ButtonProps } from '../Button'
import { Icon } from '../Icon'
import { Link } from '../Link'
import {
    Menu,
    MenuDivider,
    MenuHeader,
    MenuButton,
    MenuList,
    MenuListProps,
    MenuItemProps,
    MenuText,
    MenuItem,
    MenuLink,
    MenuHeadingType,
} from '../Menu'

export interface NavMenuSectionProps {
    headerContent?: string | ReactNode
    headerAs?: MenuHeadingType
    navItems?: NavItemProps[]
    hideDivider?: boolean
}

const NavMenuSection = forwardRef(
    ({ headerContent, headerAs, navItems = [], children, hideDivider, ...otherAttributes }, reference) => (
        <div {...otherAttributes} ref={reference}>
            {hideDivider && <MenuDivider />}
            {headerContent && <MenuHeader as={headerAs}>{headerContent}</MenuHeader>}
            {navItems.map((navItem, key) => (
                <NavMenuItem key={key} {...navItem} />
            ))}
        </div>
    )
) as ForwardReferenceComponent<'div', NavMenuSectionProps>

interface NavMenuContentProps extends MenuListProps {
    sections: NavMenuSectionProps[]
    position?: Position
}

const NavMenuContent = forwardRef(({ sections = [], position, children, ...rest }, reference) => (
    <MenuList ref={reference} {...rest} position={position}>
        {sections.map((sections, index) => (
            <NavMenuSection {...sections} hideDivider={index !== 0} key={index} />
        ))}
    </MenuList>
)) as ForwardReferenceComponent<'div', NavMenuContentProps>

interface NavItemProps extends Omit<MenuItemProps, 'children' | 'onSelect'> {
    prefixIcon?: React.ComponentType<MdiReactIconProps>
    suffixIcon?: React.ComponentType<MdiReactIconProps>
    itemContent: string | ReactNode
    itemAs?: any
    to?: string | H.LocationDescriptor<any>
    onSelect?: () => void
    key?: number | string
}

const NavMenuItem = forwardRef(({ itemContent, prefixIcon, suffixIcon, onSelect, to, itemAs, ...rest }, reference) => {
    const content = (
        <>
            {prefixIcon && <Icon as={prefixIcon} className="mr-1" />}
            {itemContent}
            {suffixIcon && <Icon as={suffixIcon} className="ml-1" />}
        </>
    )

    if (onSelect) {
        return (
            <MenuItem as={itemAs} onSelect={onSelect} ref={reference} {...rest}>
                {content}
            </MenuItem>
        )
    }
    if (to) {
        return (
            <MenuLink as={itemAs || Link} to={to} {...rest}>
                {content}
            </MenuLink>
        )
    }

    return (
        <MenuText as={itemAs || MenuItem} {...rest}>
            {content}
        </MenuText>
    )
}) as ForwardReferenceComponent<'div', NavItemProps>

type TriggerContent = string | ReactNode | ((isOpen: boolean) => string | ReactNode)
interface NavMenuTriggerProps extends ButtonProps {
    content?: TriggerContent
}
export interface NavMenuProps {
    navTrigger: NavMenuTriggerProps
    sections: NavMenuSectionProps[]
    position?: Position
}

export const NavMenu = forwardRef(({ navTrigger, sections, position }, reference) => {
    const { content, ...otherTriggerProps } = navTrigger

    const renderTiggerContent = (triggerContent: TriggerContent, isOpen: boolean): ReactNode => {
        if (typeof triggerContent === 'function') {
            return triggerContent(isOpen)
        }
        return <>{content}</>
    }

    return (
        <Menu ref={reference}>
            {({ isOpen }) => (
                <>
                    <MenuButton {...otherTriggerProps}>{renderTiggerContent(content, isOpen)}</MenuButton>
                    <NavMenuContent sections={sections} position={position} />
                </>
            )}
        </Menu>
    )
}) as ForwardReferenceComponent<'div', NavMenuProps>
