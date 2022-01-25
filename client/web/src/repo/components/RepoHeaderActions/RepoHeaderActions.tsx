import classNames from 'classnames'
import React from 'react'
import { DropdownToggle, DropdownToggleProps } from 'reactstrap'

import { ButtonLink, ButtonLinkProps } from '@sourcegraph/shared/src/components/LinkOrButton'
import { Link } from '@sourcegraph/wildcard'

import styles from './RepoHeaderActions.module.scss'

type RepoHeaderButtonLinkProps = ButtonLinkProps & {
    /**
     * to determine if this button is for file or not
     */
    file?: boolean
}

type RepoHeaderActionAnchorProps = React.AnchorHTMLAttributes<HTMLAnchorElement> & {
    /**
     * to determine if this anchor is for file or not
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

export const RepoHeaderActionDropdownToggle: React.FunctionComponent<DropdownToggleProps> = ({
    children,
    className,
    ...rest
}) => (
    <DropdownToggle className={classNames(styles.action, className)} {...rest}>
        {children}
    </DropdownToggle>
)

export const RepoHeaderActionAnchor: React.FunctionComponent<RepoHeaderActionAnchorProps> = ({
    children,
    className,
    file,
    href,
    ...rest
}) => (
    <Link
        to={href !== undefined ? href : ''}
        className={classNames(file ? styles.fileAction : styles.action, className)}
        {...rest}
    >
        {children}
    </Link>
)
