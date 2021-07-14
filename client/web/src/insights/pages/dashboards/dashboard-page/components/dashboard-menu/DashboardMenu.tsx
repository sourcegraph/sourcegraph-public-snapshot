import { Menu, MenuButton, MenuItem, MenuList } from '@reach/menu-button'
import classnames from 'classnames'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import React from 'react'

import { SettingsCascadeOrError, SettingsCascadeProps } from '@sourcegraph/shared/out/src/settings/settings'

import { Settings } from '../../../../../../schema/settings.schema'
import { InsightDashboard, isRealDashboard, isVirtualDashboard } from '../../../../../core/types'
import { isSettingsBasedInsightsDashboard } from '../../../../../core/types/dashboard/real-dashboard'
import { isGlobalSubject } from '../../../../../core/types/subjects'
import { useInsightSubjects } from '../../../../../hooks/use-insight-subjects/use-insight-subjects'

import styles from './DashboardMenu.module.scss'

export enum DashboardMenuAction {
    CopyLink,
    Delete,
    Configure,
    AddRemoveInsights,
}

export interface DashboardMenuProps extends SettingsCascadeProps<Settings> {
    dashboard?: InsightDashboard
    onSelect?: (action: DashboardMenuAction) => void
}

export const DashboardMenu: React.FunctionComponent<DashboardMenuProps> = props => {
    const { dashboard, settingsCascade, onSelect = () => {} } = props

    const hasDashboard = dashboard !== undefined
    const permissions = useDashboardPermissions(dashboard, settingsCascade)

    return (
        <Menu>
            <MenuButton className={classnames(styles.triggerButton, 'btn btn-icon')}>
                <DotsVerticalIcon size={16} />
            </MenuButton>

            <MenuList className={classnames(styles.menuList, 'dropdown-menu')}>
                <MenuItem
                    as="button"
                    disabled={!permissions.isConfigurable}
                    data-tooltip={getTooltipMessage(permissions)}
                    data-placement="right"
                    className={classnames(styles.menuItem, 'btn btn-outline')}
                    onSelect={() => onSelect(DashboardMenuAction.AddRemoveInsights)}
                >
                    Add insights
                </MenuItem>

                <MenuItem
                    as="button"
                    disabled={!permissions.isConfigurable}
                    data-tooltip={getTooltipMessage(permissions)}
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
                    data-tooltip={getTooltipMessage(permissions)}
                    data-placement="right"
                    className={classnames(styles.menuItem, 'btn btn-outline', styles.menuItemDanger)}
                    onSelect={() => onSelect(DashboardMenuAction.Delete)}
                >
                    Delete
                </MenuItem>
            </MenuList>
        </Menu>
    )
}

enum DashboardReasonDenied {
    BuiltInCantBeEdited,
    PermissionDenied,
    UnknownDashboard,
}

type DashboardPermissions =
    | {
          isConfigurable: false
          reason: DashboardReasonDenied
      }
    | {
          isConfigurable: true
      }

const DEFAULT_DASHBOARD_PERMISSIONS: DashboardPermissions = {
    isConfigurable: false,
    reason: DashboardReasonDenied.UnknownDashboard,
}

function useDashboardPermissions(
    dashboard: InsightDashboard | undefined,
    settingsCascade: SettingsCascadeOrError<Settings>
): DashboardPermissions {
    const supportedSubject = useInsightSubjects({ settingsCascade })

    if (isVirtualDashboard(dashboard)) {
        return {
            isConfigurable: false,
            reason: DashboardReasonDenied.BuiltInCantBeEdited,
        }
    }

    const dashboardOwner = supportedSubject.find(subject => subject.id === dashboard?.owner?.id)

    // No dashboard can't be modified
    if (!dashboard || !dashboardOwner) {
        return DEFAULT_DASHBOARD_PERMISSIONS
    }

    if (isRealDashboard(dashboard)) {
        // Settings based insights dashboards (custom dashboards created by users)
        if (isSettingsBasedInsightsDashboard(dashboard)) {
            // Global scope permission handling
            if (isGlobalSubject(dashboardOwner)) {
                const canBeEdited = dashboardOwner.viewerCanAdminister && dashboardOwner.allowSiteSettingsEdits

                if (!canBeEdited) {
                    return {
                        isConfigurable: false,
                        reason: DashboardReasonDenied.PermissionDenied,
                    }
                }
            }

            return {
                isConfigurable: true,
            }
        }

        // Not settings based dashboard (built-in-dashboard case)
        return {
            isConfigurable: false,
            reason: DashboardReasonDenied.BuiltInCantBeEdited,
        }
    }

    return DEFAULT_DASHBOARD_PERMISSIONS
}

function getTooltipMessage(permissions: DashboardPermissions): string | undefined {
    if (permissions.isConfigurable) {
        return
    }

    switch (permissions.reason) {
        case DashboardReasonDenied.UnknownDashboard:
            return "We didn't find a dashboard. You can't edit unknown dashboard"
        case DashboardReasonDenied.PermissionDenied:
            return "You don't a permission to edit this dashboard"
        case DashboardReasonDenied.BuiltInCantBeEdited:
            return "This dashboard is built-in. You can't edit built-in dashboards"
    }
}
