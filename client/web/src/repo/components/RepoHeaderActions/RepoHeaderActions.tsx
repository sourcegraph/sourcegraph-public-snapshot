import classNames from 'classnames'
import React from 'react'

import { ButtonLink, ButtonLinkProps, MenuButton } from '@sourcegraph/wildcard'
import { MenuButtonProps } from '@sourcegraph/wildcard/src/components/Menu'

import styles from './RepoHeaderActions.module.scss'

type RepoHeaderButtonLinkProps = ButtonLinkProps & {
    /**
     * to determine if this button is for file or not
     */
    file?: boolean
}

export const RepoHeaderActionButtonLink: React.FunctionComponent<RepoHeaderButtonLinkProps> = ({
    children,
    className,
    file,
    ...rest
}) => (
    <ButtonLink className={classNames(file ? styles.fileAction : styles.action, className)} {...rest}>
        {children}
    </ButtonLink>
)

export const RepoHeaderActionDropdownToggle: React.FunctionComponent<MenuButtonProps> = ({
    children,
    className,
    ...rest
}) => (
    <MenuButton className={classNames('btn-icon', styles.action, className)} {...rest}>
        {children}
    </MenuButton>
)

export type RepoHeaderActionAnchorProps = ButtonLinkProps & {
    /**
     * to determine if this anchor is for file or not
     */
    file?: boolean
}

export const RepoHeaderActionAnchor: React.FunctionComponent<RepoHeaderActionAnchorProps> = ({
    children,
    className,
    file,
    ...rest
}) => (
    <ButtonLink className={classNames(file ? styles.fileAction : styles.action, className)} {...rest}>
        {children}
    </ButtonLink>
)
