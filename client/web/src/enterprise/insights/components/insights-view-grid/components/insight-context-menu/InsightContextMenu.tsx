import { Menu, MenuButton, MenuItem, MenuItems, MenuLink, MenuPopover } from '@reach/menu-button'
import classnames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import React, { useContext } from 'react'
import { Link } from 'react-router-dom'

import { isSearchBasedInsightId } from '../../../../core/types/insight/search-insight'
import { DashboardInsightsContext } from '../../../../pages/dashboards/dashboard-page/components/dashboards-content/components/dashboard-inisghts/DashboardInsightsContext'

import styles from './InsightContextMenu.module.scss'

export interface InsightCardMenuProps {
    insightID: string
    zeroYAxisMin: boolean
    menuButtonClassName?: string
    onDelete: (insightID: string) => void
    onToggleZeroYAxisMin?: () => void
}

/**
 * Renders context menu (three dots menu) for particular insight card.
 */
export const InsightContextMenu: React.FunctionComponent<InsightCardMenuProps> = props => {
    const { insightID, zeroYAxisMin, menuButtonClassName, onDelete, onToggleZeroYAxisMin } = props

    // Get dashboard information in case if insight card component
    // is rendered on the dashboard page, otherwise get null value.
    const { dashboard } = useContext(DashboardInsightsContext)
    const editUrl = dashboard?.id
        ? `/insights/edit/${insightID}?dashboardId=${dashboard.id}`
        : `/insights/edit/${insightID}`

    const showYAxisToggleMenu = isSearchBasedInsightId(insightID) && onToggleZeroYAxisMin

    return (
        <Menu>
            {({ isOpen }) => (
                <>
                    <MenuButton
                        data-testid="InsightContextMenuButton"
                        className={classnames(menuButtonClassName, 'btn btn-outline p-1', styles.button)}
                    >
                        <DotsVerticalIcon
                            className={classnames(styles.buttonIcon, { [styles.buttonIconActive]: isOpen })}
                            size={16}
                        />
                    </MenuButton>
                    <MenuPopover portal={false}>
                        <MenuItems
                            data-testid={`context-menu.${insightID}`}
                            className={classnames(styles.panel, 'dropdown-menu')}
                        >
                            <MenuLink
                                as={Link}
                                data-testid="InsightContextMenuEditLink"
                                className={classnames('btn btn-outline', styles.item)}
                                to={editUrl}
                            >
                                Edit
                            </MenuLink>

                            {showYAxisToggleMenu && (
                                <MenuLink
                                    data-testid="InsightContextMenuEditLink"
                                    className={classnames('btn btn-outline border-bottom', styles.item)}
                                    onClick={onToggleZeroYAxisMin}
                                >
                                    <CheckIcon size={16} className={classnames('mr-2', { 'd-none': !zeroYAxisMin })} />{' '}
                                    Start Y Axis at 0
                                </MenuLink>
                            )}

                            <MenuItem
                                data-testid="insight-context-menu-delete-button"
                                onSelect={() => onDelete(insightID)}
                                className={classnames('btn btn-outline-', styles.item)}
                            >
                                Delete
                            </MenuItem>
                        </MenuItems>
                    </MenuPopover>
                </>
            )}
        </Menu>
    )
}
