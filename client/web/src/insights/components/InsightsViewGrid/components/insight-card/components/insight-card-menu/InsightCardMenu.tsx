import { Menu, MenuButton, MenuPopover } from '@reach/menu-button'
import classnames from 'classnames'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import React from 'react'

import styles from './InsightCardMenu.module.scss'

export interface InsightCardMenuProps {
    className?: string
    onDelete: () => void
}

/**
 * Renders context menu (three dots menu) for particular insight card.
 */
export const InsightCardMenu: React.FunctionComponent<InsightCardMenuProps> = props => {
    const { className, onDelete } = props

    return (
        <Menu>
            <MenuButton className={classnames(className, 'btn btn-outline p-1')}>
                <DotsVerticalIcon size={16} />
            </MenuButton>
            <MenuPopover portal={true}>
                <ul className={classnames('dropdown-menu', styles.menuPanel)}>
                    <li>
                        <button
                            onClick={onDelete}
                            className={classnames('btn btn-outline', styles.menuItemButton)}
                            type="button"
                        >
                            Delete
                        </button>
                    </li>
                </ul>
            </MenuPopover>
        </Menu>
    )
}
