import React from 'react'

import { Menu, MenuButton, MenuItem, MenuItems, MenuPopover } from '@reach/menu-button'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'

import { Button } from '@sourcegraph/wildcard'

import { positionBottomRight } from '../../../../../components/context-menu/utils'
import { InsightDashboard } from '../../../../../core/types'
import { useUiFeatures } from '../../../../../hooks/use-ui-features'

import styles from './DashboardMenu.module.scss'

export enum DashboardMenuAction {
    CopyLink,
    Delete,
    Configure,
    AddRemoveInsights,
}

export interface DashboardMenuProps {
    innerRef: React.Ref<HTMLButtonElement>
    dashboard?: InsightDashboard
    onSelect?: (action: DashboardMenuAction) => void
    tooltipText?: string
    className?: string
}

export const DashboardMenu: React.FunctionComponent<DashboardMenuProps> = props => {
    const { innerRef, dashboard, onSelect = () => {}, tooltipText, className } = props

    const { dashboard: dashboardPermission } = useUiFeatures()
    const menuPermissions = dashboardPermission.getContextActionsPermissions(dashboard)

    return (
        <Menu>
            <MenuButton
                as={Button}
                ref={innerRef}
                data-tooltip={tooltipText}
                data-placement="right"
                className={classNames(className, styles.triggerButton, 'btn-icon')}
            >
                <VisuallyHidden>Dashboard options</VisuallyHidden>
                <DotsVerticalIcon size={16} />
            </MenuButton>

            <MenuPopover portal={true} position={positionBottomRight}>
                <MenuItems className={classNames(styles.menuList, 'dropdown-menu')}>
                    {menuPermissions.configure.display && (
                        <MenuItem
                            as={Button}
                            disabled={menuPermissions.configure.disabled}
                            data-tooltip={menuPermissions.configure.tooltip}
                            data-placement="right"
                            className={styles.menuItem}
                            onSelect={() => onSelect(DashboardMenuAction.Configure)}
                            outline={true}
                        >
                            Configure dashboard
                        </MenuItem>
                    )}

                    {menuPermissions.copy.display && (
                        <MenuItem
                            as={Button}
                            disabled={menuPermissions.copy.disabled}
                            className={styles.menuItem}
                            onSelect={() => onSelect(DashboardMenuAction.CopyLink)}
                            outline={true}
                        >
                            Copy link
                        </MenuItem>
                    )}

                    {(menuPermissions.configure.display || menuPermissions.copy.display) &&
                        menuPermissions.delete.display && <hr />}

                    {menuPermissions.delete.display && (
                        <MenuItem
                            as={Button}
                            disabled={menuPermissions.delete.disabled}
                            data-tooltip={menuPermissions.delete.tooltip}
                            data-placement="right"
                            className={classNames(styles.menuItem, styles.menuItemDanger)}
                            onSelect={() => onSelect(DashboardMenuAction.Delete)}
                            outline={true}
                        >
                            Delete
                        </MenuItem>
                    )}
                </MenuItems>
            </MenuPopover>
        </Menu>
    )
}
