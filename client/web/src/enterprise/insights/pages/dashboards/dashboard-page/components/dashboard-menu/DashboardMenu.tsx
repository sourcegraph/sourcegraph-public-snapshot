import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { Button, Menu, MenuButton, MenuDivider, MenuItem, MenuList, Position, Icon } from '@sourcegraph/wildcard'

import { InsightDashboard } from '../../../../../core/types'
import { useUiFeatures } from '../../../../../hooks/use-ui-features'

import styles from './DashboardMenu.module.scss'
import { mdiDotsVertical } from "@mdi/js";

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

export const DashboardMenu: React.FunctionComponent<React.PropsWithChildren<DashboardMenuProps>> = props => {
    const { innerRef, dashboard, onSelect = () => {}, tooltipText, className } = props

    const { dashboard: dashboardPermission } = useUiFeatures()
    const menuPermissions = dashboardPermission.getContextActionsPermissions(dashboard)

    return (
        <Menu>
            <MenuButton
                ref={innerRef}
                data-tooltip={tooltipText}
                data-placement="right"
                variant="icon"
                outline={true}
                className={classNames(className, styles.triggerButton)}
                aria-label="dashboard context menu"
            >
                <VisuallyHidden>Dashboard options</VisuallyHidden>
                <Icon svgPath={mdiDotsVertical} inline={false} aria-hidden={true} height={16} width={16} />
            </MenuButton>

            <MenuList className={styles.menuList} position={Position.bottomEnd}>
                {menuPermissions.configure.display && (
                    <MenuItem
                        as={Button}
                        outline={true}
                        disabled={menuPermissions.configure.disabled}
                        data-tooltip={menuPermissions.configure.tooltip}
                        data-placement="right"
                        className={styles.menuItem}
                        aria-label="configure dashboard"
                        onSelect={() => onSelect(DashboardMenuAction.Configure)}
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
                    menuPermissions.delete.display && <MenuDivider />}

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
            </MenuList>
        </Menu>
    )
}
