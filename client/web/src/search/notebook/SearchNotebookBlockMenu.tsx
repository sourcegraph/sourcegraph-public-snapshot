import classNames from 'classnames'
import React from 'react'

import { Button, ButtonProps } from '@sourcegraph/wildcard'

import styles from './SearchNotebookBlockMenu.module.scss'

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

const BlockMenuActionComponent: React.FunctionComponent<BlockMenuActionComponentProps> = props => {
    const { className, label, type, id, isDisabled, icon, iconClassName, variant } = props

    const element = type === 'button' ? 'button' : 'a'
    const elementSpecificProps =
        props.type === 'button'
            ? { onClick: () => id && props.onClick(id) }
            : { href: props.url, target: '_blank', rel: 'noopener noreferrer' }

    return (
        <Button
            key={label}
            as={element as 'button'}
            className={classNames('d-flex align-items-center', className, styles.actionButton)}
            disabled={isDisabled}
            role="menuitem"
            data-testid={label}
            size="sm"
            variant={variant}
            {...elementSpecificProps}
        >
            <div className={iconClassName}>{icon}</div>
            <div className={classNames('ml-1', styles.hideOnSmallScreen)}>{label}</div>
            <div className={classNames('flex-grow-1', styles.hideOnSmallScreen)} />
            {type === 'button' && props.keyboardShortcutLabel && (
                <small className={classNames(props.keyboardShortcutLabelClassName, styles.hideOnSmallScreen)}>
                    {props.keyboardShortcutLabel}
                </small>
            )}
        </Button>
    )
}

interface SearchNotebookBlockMenuProps {
    id: string
    mainAction?: BlockMenuButtonAction
    actions: BlockMenuAction[]
}

export const SearchNotebookBlockMenu: React.FunctionComponent<SearchNotebookBlockMenuProps> = ({
    id,
    mainAction,
    actions,
}) => (
    <div className={styles.blockMenu} role="menu">
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
