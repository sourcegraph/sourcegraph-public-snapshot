import { Menu, MenuButton, MenuItem, MenuList } from '@reach/menu-button'
import classnames from 'classnames'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import React from 'react'

import { InsightDashboard, isVirtualDashboard } from '../../../../../core/types'
import { isSettingsBasedInsightsDashboard } from '../../../../../core/types/dashboard/real-dashboard'

import styles from './DashboardMenu.module.scss'

export enum DashboardMenuAction {
    CopyLink,
    Delete,
    Configure,
    AddRemoveInsights,
}

export interface DashboardMenuProps {
    dashboard?: InsightDashboard
    onSelect?: (action: DashboardMenuAction) => void
}

export const DashboardMenu: React.FunctionComponent<DashboardMenuProps> = props => {
    const { dashboard, onSelect = () => {} } = props

    const hasDashboard = dashboard !== undefined
    const isConfigurable = dashboard && !isVirtualDashboard(dashboard) && isSettingsBasedInsightsDashboard(dashboard)

    return (
        <Menu>
            <MenuButton className={classnames(styles.triggerButton, 'btn btn-icon')}>
                <DotsVerticalIcon size={16} />
            </MenuButton>

            <MenuList className={classnames(styles.menuList, 'dropdown-menu')}>
                <MenuItem
                    as="button"
                    disabled={!isConfigurable}
                    className={classnames(styles.menuItem, 'btn btn-outline')}
                    onSelect={() => onSelect(DashboardMenuAction.AddRemoveInsights)}
                >
                    Add insights
                </MenuItem>

                <MenuItem
                    as="button"
                    disabled={!isConfigurable}
                    className={classnames(styles.menuItem, 'btn btn-outline')}
                    onSelect={() => onSelect(DashboardMenuAction.Configure)}
                >
                    Configure dashboard
                </MenuItem>

                <MenuItem
                    as="button"
                    disabled={!hasDashboard}
                    className={classnames(styles.menuItem, 'btn btn-outline')}
                    onSelect={() => onSelect(DashboardMenuAction.CopyLink)}
                >
                    Copy link
                </MenuItem>

                <hr />

                <MenuItem
                    as="button"
                    className={classnames(styles.menuItem, 'btn btn-outline', styles.menuItemDanger)}
                    onSelect={() => onSelect(DashboardMenuAction.Delete)}
                >
                    Delete
                </MenuItem>
            </MenuList>
        </Menu>
    )
}
