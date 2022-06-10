import React from 'react'

import classNames from 'classnames'

import { ButtonLink, ButtonLinkProps, Button, ForwardReferenceComponent, MenuButton } from '@sourcegraph/wildcard'

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

export const RepoHeaderActionDropdownToggle: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({
    children,
    ...rest
}) => (
    <Button as={MenuButton} className={classNames('btn-icon', styles.action)} {...rest}>
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
