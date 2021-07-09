import { Menu, MenuButton, MenuList, MenuItem, MenuLink } from '@reach/menu-button'
import classnames from 'classnames'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import React, { MouseEvent } from 'react'
import { useHistory } from 'react-router'

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
    const history = useHistory()

    // According to our naming convention of insight
    // <type>.<name>.<render view = insight page | directory | home page>
    // You can see insight id generation at extension codebase like here
    // https://github.com/sourcegraph/sourcegraph-search-insights/blob/master/src/search-insights.ts#L86
    const normalizedInsightID = insightID.split('.').slice(0, -1).join('.')

    const handleEditClick = (event: MouseEvent): void => {
        event.preventDefault()
        history.push(`/insights/edit/${normalizedInsightID}`)
    }

    return (
        <Menu>
            <MenuButton data-testid="InsightContextMenuButton" className={classnames(className, 'btn btn-outline p-1')}>
                <DotsVerticalIcon size={16} />
            </MenuButton>
            <MenuList
                data-testid={`context-menu.${normalizedInsightID}`}
                className={classnames(styles.panel, 'dropdown-menu')}
            >
                <MenuLink
                    data-testid="InsightContextMenuEditLink"
                    className={classnames('btn btn-outline', styles.item)}
                    onClick={handleEditClick}
                >
                    Edit
                </MenuLink>

                <MenuItem
                    data-testid="insight-context-menu-delete-button"
                    onSelect={() => onDelete(insightID)}
                    className={classnames('btn btn-outline-', styles.item)}
                >
                    Delete
                </MenuItem>
            </MenuList>
        </Menu>
    )
}
