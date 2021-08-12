import { Menu, MenuButton, MenuItem, MenuItems, MenuLink, MenuPopover } from '@reach/menu-button'
import classnames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import React, { MouseEvent, useContext } from 'react'
import { useHistory } from 'react-router'

import { isSearchBasedInsightId } from '../../../../../../core/types'
import { positionRight } from '../../../../../context-menu/utils'
import { LineChartSettingsContext } from '../../../../../insight-view-content/chart-view-content/charts/line/line-chart-settings-provider'

import styles from './InsightCardMenu.module.scss'

export interface InsightCardMenuProps {
    menuButtonClassName?: string
    onDelete: (insightID: string) => void
    insightID: string
    onToggleZeroYAxisMin?: () => void
}

/**
 * Renders context menu (three dots menu) for particular insight card.
 */
export const InsightCardMenu: React.FunctionComponent<InsightCardMenuProps> = props => {
    const { insightID, menuButtonClassName, onDelete, onToggleZeroYAxisMin } = props
    const history = useHistory()
    const { zeroYAxisMin } = useContext(LineChartSettingsContext)

    const handleEditClick = (event: MouseEvent): void => {
        event.preventDefault()
        history.push(`/insights/edit/${insightID}`)
    }

    const showYAxisToggleMenu = isSearchBasedInsightId(insightID) && onToggleZeroYAxisMin

    return (
        <Menu>
            <MenuButton
                data-testid="InsightContextMenuButton"
                className={classnames(menuButtonClassName, 'btn btn-outline p-1')}
            >
                <DotsVerticalIcon size={16} />
            </MenuButton>
            <MenuPopover portal={true} position={positionRight}>
                <MenuItems
                    data-testid={`context-menu.${insightID}`}
                    className={classnames(styles.panel, 'dropdown-menu')}
                >
                    <MenuLink
                        data-testid="InsightContextMenuEditLink"
                        className={classnames('btn btn-outline', styles.item)}
                        onClick={handleEditClick}
                    >
                        Edit
                    </MenuLink>

                    {showYAxisToggleMenu && (
                        <MenuLink
                            data-testid="InsightContextMenuEditLink"
                            className={classnames('btn btn-outline border-bottom', styles.item)}
                            onClick={onToggleZeroYAxisMin}
                        >
                            <CheckIcon size={16} className={classnames('mr-2', { 'd-none': !zeroYAxisMin })} /> Start Y
                            Axis at Zero
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
        </Menu>
    )
}
