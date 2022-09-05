import React from 'react'

import { mdiDotsVertical } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import {
    Button,
    Icon,
    Menu,
    MenuButton,
    MenuDivider,
    MenuItem,
    MenuList,
    Position,
    Tooltip,
} from '@sourcegraph/wildcard'

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

export const DashboardMenu: React.FunctionComponent<React.PropsWithChildren<DashboardMenuProps>> = props => {
    const { innerRef, dashboard, onSelect = () => {}, tooltipText, className } = props

    const { dashboard: dashboardPermission } = useUiFeatures()
    const menuPermissions = dashboardPermission.getContextActionsPermissions(dashboard)

    return (
        <Menu>
            <Tooltip content={tooltipText} placement="right">
                <MenuButton
                    ref={innerRef}
                    variant="icon"
                    outline={true}
                    className={classNames(className, styles.triggerButton)}
                    data-testid="dashboard-context-menu"
                >
                    <VisuallyHidden>Dashboard options</VisuallyHidden>
                    <Icon
                        svgPath={mdiDotsVertical}
                        aria-label="Dashboard options"
                        inline={false}
                        height={16}
                        width={16}
                    />
                </MenuButton>
            </Tooltip>

            <MenuList className={styles.menuList} position={Position.bottomEnd}>
                {menuPermissions.configure.display && (
                    <Tooltip content={menuPermissions.configure.tooltip} placement="right">
                        <MenuItem
                            as={Button}
                            outline={true}
                            disabled={menuPermissions.configure.disabled}
                            className={styles.menuItem}
                            onSelect={() => onSelect(DashboardMenuAction.Configure)}
                            data-testid="configure-dashboard"
                        >
                            Configure dashboard
                        </MenuItem>
                    </Tooltip>
                )}

                {menuPermissions.copy.display && (
                    <MenuItem
                        as={Button}
                        disabled={menuPermissions.copy.disabled}
                        className={styles.menuItem}
                        onSelect={() => onSelect(DashboardMenuAction.CopyLink)}
                        outline={true}
                        data-testid="copy-link"
                    >
                        Copy link
                    </MenuItem>
                )}

                {(menuPermissions.configure.display || menuPermissions.copy.display) &&
                    menuPermissions.delete.display && <MenuDivider />}

                {menuPermissions.delete.display && (
                    <Tooltip content={menuPermissions.delete.tooltip} placement="right">
                        <MenuItem
                            as={Button}
                            disabled={menuPermissions.delete.disabled}
                            className={classNames(styles.menuItem, styles.menuItemDanger)}
                            onSelect={() => onSelect(DashboardMenuAction.Delete)}
                            outline={true}
                            data-testid="delete"
                        >
                            Delete
                        </MenuItem>
                    </Tooltip>
                )}
            </MenuList>
        </Menu>
    )
}
