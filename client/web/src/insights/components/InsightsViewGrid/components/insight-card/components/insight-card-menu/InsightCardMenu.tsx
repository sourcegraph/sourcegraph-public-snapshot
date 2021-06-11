import { Menu, MenuButton, MenuList, MenuItem } from '@reach/menu-button'
import classnames from 'classnames'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import styles from './InsightCardMenu.module.scss'

export interface InsightCardMenuProps {
    className?: string
    onDelete: (insightID: string) => void
    insightID: string
}

/**
 * Renders context menu (three dots menu) for particular insight card.
 */
export const InsightCardMenu: React.FunctionComponent<InsightCardMenuProps> = props => {
    const { insightID, className, onDelete } = props

    // According to our naming convention of insight
    // <type>.<name>.<render view = insight page | directory | home page>
    // You can see insight id generation at extension codebase like here
    // https://github.com/sourcegraph/sourcegraph-search-insights/blob/master/src/search-insights.ts#L86
    const normalizedInsightID = insightID.split('.').slice(0, -1).join('.')

    return (
        <Menu>
            <MenuButton data-testid="InsightContextMenuButton" className={classnames(className, 'btn btn-outline p-1')}>
                <DotsVerticalIcon size={16} />
            </MenuButton>
            <MenuList
                data-testid={`context-menu.${normalizedInsightID}`}
                className={classnames(styles.menuPanel, 'dropdown-menu dropdown-menu-sw')}
            >
                <Link
                    data-testid="InsightContextMenuEditLink"
                    className={classnames('btn btn-outline-secondary', styles.menuItemButton)}
                    to={`/insights/edit/${normalizedInsightID}`}
                >
                    Edit
                </Link>
                <MenuItem
                    data-testid="insight-context-menu-delete-button"
                    onSelect={() => onDelete(insightID)}
                    className={classnames('btn btn-outline-secondary', styles.menuItemButton)}
                >
                    Delete
                </MenuItem>
            </MenuList>
        </Menu>
    )
}
