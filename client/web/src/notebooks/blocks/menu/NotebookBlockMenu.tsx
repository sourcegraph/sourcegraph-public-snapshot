import React from 'react'

import classNames from 'classnames'

import { Button, ButtonLink, type ButtonProps } from '@sourcegraph/wildcard'

import styles from './NotebookBlockMenu.module.scss'

interface BaseBlockMenuAction {
    type: 'button' | 'link'
    icon: JSX.Element
    label: string
    isDisabled?: boolean
}

interface BlockMenuButtonAction extends BaseBlockMenuAction {
    type: 'button'
    onClick: (id: string) => void
    keyboardShortcutLabel?: string
}

interface BlockMenuLinkAction extends BaseBlockMenuAction {
    type: 'link'
    url: string
}

export type BlockMenuAction = BlockMenuButtonAction | BlockMenuLinkAction

export type BlockMenuActionComponentProps = {
    id?: string
    className?: string
    iconClassName?: string
    keyboardShortcutLabelClassName?: string
} & BlockMenuAction &
    Pick<ButtonProps, 'variant'>

const BlockMenuActionComponent: React.FunctionComponent<
    React.PropsWithChildren<BlockMenuActionComponentProps>
> = props => {
    const { className, label, type, id, isDisabled, icon, iconClassName, variant } = props

    const commonProps = {
        key: label,
        className: classNames('d-flex align-items-center', className, styles.actionButton),
        disabled: isDisabled,
        role: 'menuitem',
        'data-testid': label,
        'aria-label': label,
        size: 'sm',
        variant,
    } as const

    const commonContent = (
        <>
            <div className={iconClassName}>{icon}</div>
            <div className={classNames('ml-1', styles.hideOnSmallScreen)}>{label}</div>
            <div className={classNames('flex-grow-1', styles.hideOnSmallScreen)} />
        </>
    )

    if (type === 'button') {
        return (
            <Button {...commonProps} onClick={() => id && props.onClick(id)}>
                {commonContent}
                {props.keyboardShortcutLabel && (
                    <small className={classNames(props.keyboardShortcutLabelClassName, styles.hideOnSmallScreen)}>
                        {props.keyboardShortcutLabel}
                    </small>
                )}
            </Button>
        )
    }

    return (
        <ButtonLink {...commonProps} to={props.url} target="_blank" rel="noopener noreferrer">
            {commonContent}
        </ButtonLink>
    )
}

export interface NotebookBlockMenuProps {
    id: string
    mainAction?: BlockMenuButtonAction
    actions: BlockMenuAction[]
}

export const NotebookBlockMenu: React.FunctionComponent<React.PropsWithChildren<NotebookBlockMenuProps>> = ({
    id,
    mainAction,
    actions,
}) => (
    <div
        className={classNames('block-menu', styles.blockMenu)}
        // To fix Rule: "aria-required-children"
        // Fails accessibility rule when div has no children with role="menu"
        role={!!mainAction || actions.length > 0 ? 'menu' : undefined}
    >
        {mainAction && (
            <div className={classNames(actions.length > 0 && styles.mainActionButtonWrapper)}>
                <BlockMenuActionComponent variant="primary" className="w-100" id={id} {...mainAction} />
            </div>
        )}
        {actions.map(action => {
            if (action.type === 'button') {
                return (
                    <BlockMenuActionComponent
                        key={action.label}
                        id={id}
                        iconClassName="text-muted"
                        keyboardShortcutLabelClassName="text-muted"
                        {...action}
                    />
                )
            }
            return <BlockMenuActionComponent key={action.label} {...action} />
        })}
    </div>
)
