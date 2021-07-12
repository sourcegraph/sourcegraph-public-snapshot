import React from 'react'

import styles from './SearchNotebookBlockMenu.module.scss'

interface BlockMenuAction {
    onClick: (id: string) => void
    // icon: string
    label: string
    // keyboardShortcutLabel: string
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
            <div className={styles.mainActionButton}>
                <button
                    className="btn btn-sm btn-primary w-100"
                    type="button"
                    role="menuitem"
                    disabled={mainAction.isDisabled ?? false}
                    onClick={() => mainAction.onClick(id)}
                >
                    <div>{mainAction.label}</div>
                </button>
            </div>
        )}
        {actions.map(action => (
            <button
                key={action.label}
                className="btn btn-sm btn-secondary mb-2"
                type="button"
                role="menuitem"
                disabled={action.isDisabled ?? false}
                onClick={() => action.onClick(id)}
            >
                <div>{action.label}</div>
            </button>
        ))}
    </div>
)
