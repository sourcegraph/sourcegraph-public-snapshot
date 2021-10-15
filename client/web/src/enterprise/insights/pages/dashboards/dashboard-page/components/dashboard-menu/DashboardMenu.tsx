import { Menu, MenuButton, MenuItem, MenuItems, MenuPopover } from '@reach/menu-button'
import classnames from 'classnames'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import React from 'react'

import { positionBottomRight } from '../../../../../components/context-menu/utils'
import { InsightDashboard } from '../../../../../core/types'
import { SupportedInsightSubject } from '../../../../../core/types/subjects';
import { getTooltipMessage, useDashboardPermissions } from '../../hooks/use-dashboard-permissions'

import styles from './DashboardMenu.module.scss'

export enum DashboardMenuAction {
    CopyLink,
    Delete,
    Configure,
    AddRemoveInsights,
}

export interface DashboardMenuProps {
    innerRef: React.Ref<HTMLButtonElement>
    subjects?: SupportedInsightSubject[]
    dashboard?: InsightDashboard
    onSelect?: (action: DashboardMenuAction) => void
    tooltipText?: string
}

export const DashboardMenu: React.FunctionComponent<DashboardMenuProps> = props => {
    const { innerRef, dashboard, subjects, onSelect = () => {}, tooltipText } = props

    const hasDashboard = dashboard !== undefined
    const permissions = useDashboardPermissions(dashboard, subjects)

    return (
        <Menu>
            <MenuButton
                ref={innerRef}
                data-tooltip={tooltipText}
                data-placement="right"
                className={classnames(styles.triggerButton, 'btn btn-icon')}
            >
                <DotsVerticalIcon size={16} />
            </MenuButton>

            <MenuPopover portal={true} position={positionBottomRight}>
                <MenuItems className={classnames(styles.menuList, 'dropdown-menu')}>
                    <MenuItem
                        as="button"
                        disabled={!permissions.isConfigurable}
                        data-tooltip={getTooltipMessage(dashboard, permissions)}
                        data-placement="right"
                        className={classnames(styles.menuItem, 'btn btn-outline')}
                        onSelect={() => onSelect(DashboardMenuAction.AddRemoveInsights)}
                    >
                        Add or remove insights
                    </MenuItem>

                    <MenuItem
                        as="button"
                        disabled={!permissions.isConfigurable}
                        data-tooltip={getTooltipMessage(dashboard, permissions)}
                        data-placement="right"
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
                        disabled={!permissions.isConfigurable}
                        data-tooltip={getTooltipMessage(dashboard, permissions)}
                        data-placement="right"
                        className={classnames(styles.menuItem, 'btn btn-outline', styles.menuItemDanger)}
                        onSelect={() => onSelect(DashboardMenuAction.Delete)}
                    >
                        Delete
                    </MenuItem>
                </MenuItems>
            </MenuPopover>
        </Menu>
    )
}
