import classNames from 'classnames'
import React, { PropsWithChildren } from 'react'
import { DropdownToggle, DropdownToggleProps } from 'reactstrap'

import { ButtonLink, ButtonLinkProps } from '@sourcegraph/shared/src/components/LinkOrButton'
import { Button } from '@sourcegraph/wildcard'

import styles from './RepoHeaderActions.module.scss'

type RepoHeaderButtonLinkProps = ButtonLinkProps & {
    /**
     * to determine if this button is for file or not
     */
    file?: boolean
}

export type RepoHeaderActionAnchorProps = React.AnchorHTMLAttributes<HTMLAnchorElement> & {
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
    <Button<typeof ButtonLink, PropsWithChildren<ButtonLinkProps>>
        as={ButtonLink}
        className={classNames(file ? styles.fileAction : styles.action, className)}
        {...rest}
    >
        {children}
    </Button>
)

export const RepoHeaderActionDropdownToggle: React.FunctionComponent<DropdownToggleProps> = ({
    children,
    className,
    ...rest
}) => (
    <Button<typeof DropdownToggle, DropdownToggleProps>
        as={DropdownToggle}
        className={classNames('btn-icon', styles.action, className)}
        {...rest}
    >
        {children}
    </Button>
)

export const RepoHeaderActionAnchor: React.FunctionComponent<RepoHeaderActionAnchorProps> = ({
    children,
    className,
    file,
    ...rest
}) => (
    <Button<'a', RepoHeaderActionAnchorProps>
        as="a"
        className={classNames(file ? styles.fileAction : styles.action, className)}
        {...rest}
    >
        {children}
    </Button>
)
