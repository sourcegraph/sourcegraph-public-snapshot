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

const BlockMenuActionComponent: React.FunctionComponent<
    {
        id?: string
        className?: string
        iconClassName?: string
        keyboardShorcutLabelClassName?: string
        collapseMenu: boolean
    } & BlockMenuAction
> = props => {
    const Element = props.type === 'button' ? 'button' : 'a'
    const elementSpecificProps =
        props.type === 'button'
            ? { onClick: () => props.id && props.onClick(props.id), disabled: props.isDisabled ?? false }
            : { href: props.url, target: '_blank', rel: 'noopener noreferrer' }
    const commonClassNames = [styles.hideOnSmallScreen, props.collapseMenu && 'collapse-menu']
    return (
        <Element
            key={props.label}
            className={classNames('btn btn-sm d-flex align-items-center', props.className, styles.actionButton)}
            type="button"
            role="menuitem"
            data-testid={props.label}
            {...elementSpecificProps}
        >
            <div className={props.iconClassName}>{props.icon}</div>
            <div className={classNames('ml-1', ...commonClassNames)}>{props.label}</div>
            <div className={classNames('flex-grow-1', ...commonClassNames)} />
            {props.type === 'button' && (
                <small className={classNames(props.keyboardShorcutLabelClassName, ...commonClassNames)}>
                    {props.keyboardShortcutLabel}
                </small>
            )}
        </Element>
    )
}

interface SearchNotebookBlockMenuProps {
    id: string
    mainAction?: BlockMenuButtonAction
    actions: BlockMenuAction[]
    collapseMenu: boolean
}

export const SearchNotebookBlockMenu: React.FunctionComponent<SearchNotebookBlockMenuProps> = ({
    id,
    mainAction,
    actions,
    collapseMenu,
}) => (
    <div className={classNames(styles.blockMenu, collapseMenu && 'collapse-menu')} role="menu">
        {mainAction && (
            <div className={classNames(actions.length > 0 && styles.mainActionButtonWrapper)}>
                <BlockMenuActionComponent
                    className="btn-primary w-100"
                    id={id}
                    collapseMenu={collapseMenu}
                    {...mainAction}
                />
            </div>
        )}
        {actions.map(action => {
            if (action.type === 'button') {
                return (
                    <BlockMenuActionComponent
                        key={action.label}
                        id={id}
                        iconClassName="text-muted"
                        keyboardShorcutLabelClassName="text-muted"
                        collapseMenu={collapseMenu}
                        {...action}
                    />
                )
            }
            return <BlockMenuActionComponent key={action.label} collapseMenu={collapseMenu} {...action} />
        })}
    </div>
)
