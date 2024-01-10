import React from 'react'

import classNames from 'classnames'

import {
    ButtonLink,
    type ButtonLinkProps,
    type ForwardReferenceComponent,
    MenuButton,
    MenuItem,
    MenuLink,
    type MenuItemProps,
    type MenuLinkProps,
} from '@sourcegraph/wildcard'

import styles from './RepoHeaderActions.module.scss'

type RepoHeaderButtonLinkProps = ButtonLinkProps & {
    /**
     * to determine if this button is for file or not
     */
    file?: boolean
}

export const RepoHeaderActionButtonLink = React.forwardRef(
    ({ children, className, file, ...rest }: React.PropsWithChildren<RepoHeaderButtonLinkProps>, reference) => (
        <ButtonLink
            ref={reference}
            className={classNames(file ? styles.fileAction : styles.action, className)}
            {...rest}
        >
            {children}
        </ButtonLink>
    )
) as ForwardReferenceComponent<typeof ButtonLink, React.PropsWithChildren<RepoHeaderButtonLinkProps>>
RepoHeaderActionButtonLink.displayName = 'RepoHeaderActionButtonLink'

export const RepoHeaderActionDropdownToggle: React.FunctionComponent<
    React.PropsWithChildren<Pick<React.AriaAttributes, 'aria-label'>>
> = ({ children, ...ariaAttributes }) => (
    <MenuButton className={styles.action} {...ariaAttributes}>
        {children}
    </MenuButton>
)

export type RepoHeaderActionAnchorProps = Omit<ButtonLinkProps, 'as' | 'href'> & {
    /**
     * to determine if this anchor is for file or not
     */
    file?: boolean
}

export const RepoHeaderActionAnchor = React.forwardRef((props: RepoHeaderActionAnchorProps, reference) => {
    const { children, className, file, ...rest } = props

    return (
        <ButtonLink
            className={classNames(file ? styles.fileAction : styles.action, className)}
            ref={reference}
            {...rest}
        >
            {children}
        </ButtonLink>
    )
}) as ForwardReferenceComponent<typeof ButtonLink, RepoHeaderActionAnchorProps>
RepoHeaderActionAnchor.displayName = 'RepoHeaderActionAnchor'

type RepoHeaderActionMenuLinkProps = MenuLinkProps & {
    /**
     * to determine if this button is for file or not
     */
    file?: boolean
    className?: string
}

export const RepoHeaderActionMenuLink = React.forwardRef((props: RepoHeaderActionMenuLinkProps, reference) => {
    const { children, className, file, ...rest } = props

    return (
        <MenuLink className={classNames(file ? styles.fileAction : styles.action, className)} ref={reference} {...rest}>
            {children}
        </MenuLink>
    )
}) as ForwardReferenceComponent<'a', RepoHeaderActionMenuLinkProps>
RepoHeaderActionMenuLink.displayName = 'RepoHeaderActionMenuLink'

export type RepoHeaderActionMenuItemProps = MenuItemProps & {
    /**
     * to determine if this anchor is for file or not
     */
    file?: boolean
    className?: string
}

export const RepoHeaderActionMenuItem = React.forwardRef((props: RepoHeaderActionMenuItemProps, reference) => {
    const { children, className, file, onSelect, ...rest } = props

    return (
        <MenuItem
            className={classNames(file ? styles.fileAction : styles.action, className)}
            ref={reference}
            onSelect={onSelect}
            {...rest}
        >
            {children}
        </MenuItem>
    )
}) as ForwardReferenceComponent<'div', RepoHeaderActionMenuItemProps>
RepoHeaderActionMenuItem.displayName = 'RepoHeaderActionMenuItem'
