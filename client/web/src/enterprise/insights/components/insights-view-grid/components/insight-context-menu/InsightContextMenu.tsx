import React from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'

import { Link, Menu, MenuButton, MenuDivider, MenuItem, MenuLink, MenuList, Position } from '@sourcegraph/wildcard'

import { Insight, InsightDashboard, isSearchBasedInsight, isVirtualDashboard } from '../../../../core/types'

import styles from './InsightContextMenu.module.scss'

export interface InsightCardMenuProps {
    insight: Insight
    dashboard: InsightDashboard | null
    zeroYAxisMin: boolean
    menuButtonClassName?: string
    onDelete: (insightID: string) => void
    onRemoveFromDashboard: (dashboard: InsightDashboard) => unknown
    onToggleZeroYAxisMin?: () => void
}

/**
 * Renders context menu (three dots menu) for particular insight card.
 */
export const InsightContextMenu: React.FunctionComponent<InsightCardMenuProps> = props => {
    const {
        insight,
        dashboard,
        zeroYAxisMin,
        menuButtonClassName,
        onDelete,
        onRemoveFromDashboard,
        onToggleZeroYAxisMin = noop,
    } = props

    const insightID = insight.id
    const editUrl = dashboard?.id
        ? `/insights/edit/${insightID}?dashboardId=${dashboard.id}`
        : `/insights/edit/${insightID}`

    const withinVirtualDashboard = !!dashboard && isVirtualDashboard(dashboard)

    return (
        <Menu>
            {({ isOpen }) => (
                <>
                    <MenuButton
                        data-testid="InsightContextMenuButton"
                        className={classNames(menuButtonClassName, 'p-1', styles.button)}
                        aria-label="Insight options"
                        outline={true}
                    >
                        <DotsVerticalIcon
                            className={classNames(styles.buttonIcon, { [styles.buttonIconActive]: isOpen })}
                            size={16}
                        />
                    </MenuButton>
                    <MenuList position={Position.bottomEnd} data-testid={`context-menu.${insightID}`}>
                        <MenuLink
                            as={Link}
                            data-testid="InsightContextMenuEditLink"
                            className={styles.item}
                            to={editUrl}
                        >
                            Edit
                        </MenuLink>

                        {isSearchBasedInsight(insight) && (
                            <MenuItem
                                role="menuitemcheckbox"
                                data-testid="InsightContextMenuEditLink"
                                className={classNames('d-flex align-items-center justify-content-between', styles.item)}
                                onSelect={onToggleZeroYAxisMin}
                                aria-checked={zeroYAxisMin}
                            >
                                <input
                                    type="checkbox"
                                    aria-hidden="true"
                                    checked={zeroYAxisMin}
                                    onChange={noop}
                                    tabIndex={-1}
                                />
                                <span>Start Y Axis at 0</span>
                            </MenuItem>
                        )}

                        {dashboard && (
                            <MenuItem
                                data-testid="insight-context-remove-from-dashboard-button"
                                onSelect={() => onRemoveFromDashboard(dashboard)}
                                disabled={withinVirtualDashboard}
                                data-tooltip={
                                    withinVirtualDashboard
                                        ? "Removing insight isn't available for the All insights dashboard"
                                        : undefined
                                }
                                data-placement="left"
                                className={styles.item}
                            >
                                Remove from this dashboard
                            </MenuItem>
                        )}

                        <MenuDivider />

                        <MenuItem
                            data-testid="insight-context-menu-delete-button"
                            onSelect={() => onDelete(insightID)}
                            className={styles.item}
                        >
                            Delete
                        </MenuItem>
                    </MenuList>
                </>
            )}
        </Menu>
    )
}
