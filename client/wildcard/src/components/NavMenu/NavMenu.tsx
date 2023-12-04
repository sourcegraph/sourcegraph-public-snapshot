import React, { forwardRef, type ReactNode } from 'react'

import type * as H from 'history'
import type { MdiReactIconProps } from 'mdi-react'

import type { ForwardReferenceComponent, Position } from '../..'
import type { ButtonProps } from '../Button'
import { Icon } from '../Icon'
import { Link } from '../Link'
import {
    Menu,
    MenuDivider,
    MenuHeader,
    MenuButton,
    MenuList,
    type MenuListProps,
    type MenuItemProps,
    MenuText,
    MenuItem,
    MenuLink,
    type MenuHeadingType,
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
            {navItems.map(({ key, ...rest }) => (
                <NavMenuItem key={key} {...rest} />
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
    prefixIcon?: React.ComponentType<React.PropsWithChildren<MdiReactIconProps>>
    suffixIcon?: React.ComponentType<React.PropsWithChildren<MdiReactIconProps>>
    content: string | ReactNode
    itemAs?: any
    to?: string | H.LocationDescriptor<any>
    onSelect?: () => void
    key: number | string
}

const NavMenuItem = forwardRef(({ content, prefixIcon, suffixIcon, onSelect, to, itemAs, ...rest }, reference) => {
    const contentWithIcon = (
        <>
            {prefixIcon && <Icon aria-hidden={true} as={prefixIcon} className="mr-1" />}
            {content}
            {suffixIcon && <Icon aria-hidden={true} as={suffixIcon} className="ml-1" />}
        </>
    )

    if (onSelect) {
        return (
            <MenuItem as={itemAs} onSelect={onSelect} ref={reference} {...rest}>
                {contentWithIcon}
            </MenuItem>
        )
    }
    if (to) {
        return (
            <MenuLink as={itemAs || Link} to={to} {...rest}>
                {contentWithIcon}
            </MenuLink>
        )
    }

    return (
        <MenuText as={itemAs || MenuItem} {...rest}>
            {contentWithIcon}
        </MenuText>
    )
}) as ForwardReferenceComponent<'div', NavItemProps>

type TriggerContent = { text: string } | { node: ReactNode } | { trigger: (isOpen: boolean) => string | ReactNode }

interface NavMenuTriggerProps extends ButtonProps {
    triggerContent: TriggerContent
}
export interface NavMenuProps {
    navTrigger: NavMenuTriggerProps
    sections: NavMenuSectionProps[]
    position?: Position
}

const renderTiggerContent = (triggerContent: TriggerContent, isOpen: boolean): ReactNode => {
    if ('trigger' in triggerContent) {
        return triggerContent.trigger(isOpen)
    }
    if ('node' in triggerContent) {
        return triggerContent.node
    }
    return <>{triggerContent.text}</>
}

export const NavMenu = forwardRef(({ navTrigger, sections, position }, reference) => {
    const { triggerContent, ...otherTriggerProps } = navTrigger

    return (
        <Menu ref={reference}>
            {({ isOpen }) => (
                <>
                    <MenuButton {...otherTriggerProps}>{renderTiggerContent(triggerContent, isOpen)}</MenuButton>
                    <NavMenuContent sections={sections} position={position} />
                </>
            )}
        </Menu>
    )
}) as ForwardReferenceComponent<'div', NavMenuProps>
