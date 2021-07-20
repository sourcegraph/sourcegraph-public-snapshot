import classNames from 'classnames'
import React from 'react'

import styles from './SearchNotebookBlockMenu.module.scss'

export interface BlockMenuAction {
    onClick: (id: string) => void
    icon: JSX.Element
    label: string
    keyboardShortcutLabel: string
    isDisabled?: boolean
}

interface SearchNotebookBlockMenuProps {
    id: string
    mainAction?: BlockMenuAction
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
                <button
                    className="btn btn-sm btn-primary d-flex align-items-center w-100"
                    type="button"
                    role="menuitem"
                    data-testid={mainAction.label}
                    disabled={mainAction.isDisabled ?? false}
                    onClick={() => mainAction.onClick(id)}
                >
                    <div>{mainAction.icon}</div>
                    <div className={classNames('ml-1', styles.hideOnSmallScreen)}>{mainAction.label}</div>
                    <div className={classNames('flex-grow-1', styles.hideOnSmallScreen)} />
                    <small className={styles.hideOnSmallScreen}>{mainAction.keyboardShortcutLabel}</small>
                </button>
            </div>
        )}
        {actions.map(action => (
            <button
                key={action.label}
                className={classNames('btn btn-sm d-flex align-items-center', styles.actionButton)}
                type="button"
                role="menuitem"
                data-testid={action.label}
                disabled={action.isDisabled ?? false}
                onClick={() => action.onClick(id)}
            >
                <div className="text-muted">{action.icon}</div>
                <div className={classNames('ml-1', styles.hideOnSmallScreen)}>{action.label}</div>
                <div className={classNames('flex-grow-1', styles.hideOnSmallScreen)} />
                <small className={classNames('text-muted', styles.hideOnSmallScreen)}>
                    {action.keyboardShortcutLabel}
                </small>
            </button>
        ))}
    </div>
)
