import { Menu, MenuButton, MenuItem, MenuItems, MenuPopover } from '@reach/menu-button'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { positionBottomRight } from '../../../../../components/context-menu/utils'
import { InsightDashboard } from '../../../../../core/types'
import { SupportedInsightSubject } from '../../../../../core/types/subjects'
import { getTooltipMessage, getDashboardPermissions } from '../../utils/get-dashboard-permissions'

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
    className?: string
}

export const DashboardMenu: React.FunctionComponent<DashboardMenuProps> = props => {
    const { innerRef, dashboard, subjects, onSelect = () => {}, tooltipText, className } = props

    const hasDashboard = dashboard !== undefined
    const permissions = getDashboardPermissions(dashboard, subjects)

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
                    <MenuItem
                        as={Button}
                        disabled={!permissions.isConfigurable}
                        data-tooltip={getTooltipMessage(dashboard, permissions)}
                        data-placement="right"
                        className={styles.menuItem}
                        onSelect={() => onSelect(DashboardMenuAction.Configure)}
                        outline={true}
                    >
                        Configure dashboard
                    </MenuItem>

                    <MenuItem
                        as={Button}
                        disabled={!hasDashboard}
                        className={styles.menuItem}
                        onSelect={() => onSelect(DashboardMenuAction.CopyLink)}
                        outline={true}
                    >
                        Copy link
                    </MenuItem>

                    <hr />

                    <MenuItem
                        as={Button}
                        disabled={!permissions.isConfigurable}
                        data-tooltip={getTooltipMessage(dashboard, permissions)}
                        data-placement="right"
                        className={classNames(styles.menuItem, styles.menuItemDanger)}
                        onSelect={() => onSelect(DashboardMenuAction.Delete)}
                        outline={true}
                    >
                        Delete
                    </MenuItem>
                </MenuItems>
            </MenuPopover>
        </Menu>
    )
}
