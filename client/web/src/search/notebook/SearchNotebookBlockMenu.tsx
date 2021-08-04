import classNames from 'classnames'
import React from 'react'

import styles from './SearchNotebookBlockMenu.module.scss'

interface BaseBlockMenuAction {
    type: 'button' | 'link'
    icon: JSX.Element
    label: string
}

interface BlockMenuButtonAction extends BaseBlockMenuAction {
    type: 'button'
    onClick: (id: string) => void
    keyboardShortcutLabel: string
    isDisabled?: boolean
}

interface BlockMenuLinkAction extends BaseBlockMenuAction {
    type: 'link'
    url: string
}

export type BlockMenuAction = BlockMenuButtonAction | BlockMenuLinkAction

interface SearchNotebookBlockMenuProps {
    id: string
    mainAction?: BlockMenuButtonAction
    actions: BlockMenuAction[]
}

const BlockMenuButtonActionComponent: React.FunctionComponent<
    {
        id: string
        className?: string
        iconClassName?: string
        keyboardShorcutLabelClassName?: string
    } & BlockMenuButtonAction
> = ({
    id,
    className,
    iconClassName,
    keyboardShorcutLabelClassName,
    label,
    icon,
    isDisabled,
    onClick,
    keyboardShortcutLabel,
}) => (
    <button
        key={label}
        className={classNames('btn btn-sm d-flex align-items-center', className, styles.actionButton)}
        type="button"
        role="menuitem"
        data-testid={label}
        disabled={isDisabled ?? false}
        onClick={() => onClick(id)}
    >
        <div className={iconClassName}>{icon}</div>
        <div className={classNames('ml-1', styles.hideOnSmallScreen)}>{label}</div>
        <div className={classNames('flex-grow-1', styles.hideOnSmallScreen)} />
        <small className={classNames(keyboardShorcutLabelClassName, styles.hideOnSmallScreen)}>
            {keyboardShortcutLabel}
        </small>
    </button>
)

const BlockMenuButtonLinkComponent: React.FunctionComponent<{ className?: string } & BlockMenuLinkAction> = ({
    className,
    label,
    icon,
    url,
}) => (
    <a
        key={label}
        href={url}
        className={classNames('btn btn-sm d-flex align-items-center', className, styles.actionButton)}
        type="button"
        role="menuitem"
        target="_blank"
        rel="noopener noreferrer"
        data-testid={label}
    >
        <div className="text-muted">{icon}</div>
        <div className={classNames('ml-1', styles.hideOnSmallScreen)}>{label}</div>
        <div className={classNames('flex-grow-1', styles.hideOnSmallScreen)} />
    </a>
)

export const SearchNotebookBlockMenu: React.FunctionComponent<SearchNotebookBlockMenuProps> = ({
    id,
    mainAction,
    actions,
}) => (
    <div className={styles.blockMenu} role="menu">
        {mainAction && (
            <div className={classNames(actions.length > 0 && styles.mainActionButtonWrapper)}>
                <BlockMenuButtonActionComponent className="btn-primary w-100" id={id} {...mainAction} />
            </div>
        )}
        {actions.map(action => {
            if (action.type === 'button') {
                return (
                    <BlockMenuButtonActionComponent
                        key={action.label}
                        id={id}
                        iconClassName="text-muted"
                        keyboardShorcutLabelClassName="text-muted"
                        {...action}
                    />
                )
            }
            return <BlockMenuButtonLinkComponent key={action.label} {...action} />
        })}
    </div>
)
