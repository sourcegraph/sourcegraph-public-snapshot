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

// eslint-disable-next-line react/display-name
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

// eslint-disable-next-line react/display-name
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
