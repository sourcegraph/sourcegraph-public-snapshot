import classNames from 'classnames'
import React from 'react'

import { ButtonLink, ButtonLinkProps, Button, MenuButton } from '@sourcegraph/wildcard'

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

export const RepoHeaderActionDropdownToggle: React.FunctionComponent = ({ children }) => (
    <Button as={MenuButton} className={classNames('btn-icon', styles.action)}>
        {children}
    </Button>
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
