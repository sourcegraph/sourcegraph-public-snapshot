import React from 'react'

import classNames from 'classnames'

import {
    ButtonLink,
    ButtonLinkProps,
    Button,
    ForwardReferenceComponent,
    MenuButton,
    MenuItem,
    MenuLink,
    MenuItemProps,
    MenuLinkProps,
} from '@sourcegraph/wildcard'

import styles from './RepoHeaderActions.module.scss'

type RepoHeaderButtonLinkProps = ButtonLinkProps & {
    /**
     * to determine if this button is for file or not
     */
    file?: boolean
}

export const RepoHeaderActionButtonLink: React.FunctionComponent<
    React.PropsWithChildren<RepoHeaderButtonLinkProps>
> = ({ children, className, file, ...rest }) => (
    <ButtonLink className={classNames(file ? styles.fileAction : styles.action, className)} {...rest}>
        {children}
    </ButtonLink>
)

export const RepoHeaderActionDropdownToggle: React.FunctionComponent<
    React.PropsWithChildren<Pick<React.AriaAttributes, 'aria-label'>>
> = ({ children, ...ariaAttributes }) => (
    <Button as={MenuButton} className={classNames('btn-icon', styles.action)} {...ariaAttributes}>
        {children}
    </Button>
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
        <MenuLink
            className={classNames(file ? styles.fileAction : styles.action, className)}
            ref={reference}
            // keep Menu open on select (mouse only)
            onMouseUp={event => event.preventDefault()}
            {...rest}
        >
            {children}
        </MenuLink>
    )
}) as ForwardReferenceComponent<'a', RepoHeaderActionMenuLinkProps>

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
            // keep Menu open on select (mouse only)
            onMouseUp={event => event.preventDefault()}
            // we still need `onSelect` binding for keyboard users
            // but we don't know how to keep Menu open for keyboard usages
            onSelect={onSelect}
            // `onClick` will be triggered when using mouse
            // instead of `onSelect` because we already catch `onMouseUp`
            onClick={onSelect}
            {...rest}
        >
            {children}
        </MenuItem>
    )
}) as ForwardReferenceComponent<'div', RepoHeaderActionMenuItemProps>
