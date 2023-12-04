import React from 'react'

import { mdiDotsVertical } from '@mdi/js'
import classNames from 'classnames'
import { noop } from 'lodash'

import { Menu, MenuButton, MenuItem, MenuList, Position, Checkbox, Icon } from '@sourcegraph/wildcard'

import type { Insight } from '../../../../../core'
import { useUiFeatures } from '../../../../../hooks/use-ui-features'

import styles from './StandaloneInsightContextMenu.module.scss'

export interface StandaloneInsightContextMenuProps {
    insight: Insight
    zeroYAxisMin: boolean
    onToggleZeroYAxisMin: (zeroYAxisMin: boolean) => void
}

export const StandaloneInsightContextMenu: React.FunctionComponent<StandaloneInsightContextMenuProps> = props => {
    const { insight, zeroYAxisMin, onToggleZeroYAxisMin = noop } = props

    const { insight: insightPermissions } = useUiFeatures()
    const menuPermissions = insightPermissions.getContextActionsPermissions(insight)

    if (!menuPermissions.showYAxis) {
        return null
    }

    return (
        <Menu>
            {({ isOpen }) => (
                <>
                    <MenuButton
                        data-testid="InsightContextMenuButton"
                        className={classNames('p-1', styles.button)}
                        aria-label="Insight options"
                        outline={true}
                    >
                        <Icon
                            className={classNames(styles.buttonIcon, { [styles.buttonIconActive]: isOpen })}
                            svgPath={mdiDotsVertical}
                            inline={false}
                            aria-hidden={true}
                            height={16}
                            width={16}
                        />
                    </MenuButton>
                    <MenuList position={Position.bottomEnd} data-testid={`context-menu.${insight.id}`}>
                        <MenuItem
                            role="menuitemcheckbox"
                            data-testid="InsightContextMenuEditLink"
                            className={classNames('d-flex align-items-center justify-content-end', styles.item)}
                            onSelect={() => onToggleZeroYAxisMin(!zeroYAxisMin)}
                            aria-checked={zeroYAxisMin}
                        >
                            <Checkbox
                                aria-hidden="true"
                                checked={zeroYAxisMin}
                                onChange={noop}
                                tabIndex={-1}
                                id="InsightContextMenuEditInput"
                                label="Start Y Axis at 0"
                            />
                        </MenuItem>
                    </MenuList>
                </>
            )}
        </Menu>
    )
}
