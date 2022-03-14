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
    const {
        dashboards: { menu },
    } = useUiFeatures({ currentDashboard: dashboard })

    const hasDashboard = dashboard !== undefined

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
                    {menu.configure.display && (
                        <MenuItem
                            as={Button}
                            disabled={menu.configure.disabled}
                            data-tooltip={menu.configure.tooltip}
                            data-placement="right"
                            className={styles.menuItem}
                            onSelect={() => onSelect(DashboardMenuAction.Configure)}
                            outline={true}
                        >
                            Configure dashboard
                        </MenuItem>
                    )}

                    {menu.copy.display && (
                        <MenuItem
                            as={Button}
                            disabled={!hasDashboard}
                            className={styles.menuItem}
                            onSelect={() => onSelect(DashboardMenuAction.CopyLink)}
                            outline={true}
                        >
                            Copy link
                        </MenuItem>
                    )}

                    {menu.delete.display && (
                        <>
                            <hr />

                            <MenuItem
                                as={Button}
                                disabled={menu.delete.disabled}
                                data-tooltip={menu.delete.tooltip}
                                data-placement="right"
                                className={classNames(styles.menuItem, styles.menuItemDanger)}
                                onSelect={() => onSelect(DashboardMenuAction.Delete)}
                                outline={true}
                            >
                                Delete
                            </MenuItem>
                        </>
                    )}
                </MenuItems>
            </MenuPopover>
        </Menu>
    )
}
