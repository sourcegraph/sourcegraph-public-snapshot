import React, { useState } from 'react'

import { mdiDotsVertical } from '@mdi/js'
import classNames from 'classnames'
import { noop } from 'lodash'

import {
    Link,
    Menu,
    MenuButton,
    MenuDivider,
    MenuItem,
    MenuLink,
    MenuList,
    Position,
    Checkbox,
    Tooltip,
    Icon,
} from '@sourcegraph/wildcard'

import { useExperimentalFeatures } from '../../../../../../stores'
import { Insight, InsightDashboard, InsightType, isVirtualDashboard } from '../../../../core'
import { useUiFeatures } from '../../../../hooks'
import { ConfirmDeleteModal } from '../../../modals/ConfirmDeleteModal'
import { ShareLinkModal } from '../../../modals/ShareLinkModal/ShareLinkModal'

import { ConfirmRemoveModal } from './ConfirmRemoveModal'

import styles from './InsightContextMenu.module.scss'

export interface InsightCardMenuProps {
    insight: Insight
    currentDashboard: InsightDashboard | null
    dashboards: InsightDashboard[]
    zeroYAxisMin: boolean
    onToggleZeroYAxisMin?: () => void
}

/**
 * Renders context menu (three dots menu) for particular insight card.
 */
export const InsightContextMenu: React.FunctionComponent<React.PropsWithChildren<InsightCardMenuProps>> = props => {
    const { insight, currentDashboard, dashboards, zeroYAxisMin, onToggleZeroYAxisMin = noop } = props

    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
    const [showRemoveConfirm, setShowRemoveConfirm] = useState(false)
    const [showShareModal, setShowShareModal] = useState(false)

    const { insight: insightPermissions } = useUiFeatures()
    const menuPermissions = insightPermissions.getContextActionsPermissions(insight)

    const insightID = insight.id
    const editUrl = currentDashboard?.id
        ? `/insights/edit/${insightID}?dashboardId=${currentDashboard.id}`
        : `/insights/edit/${insightID}`

    const features = useExperimentalFeatures()
    const showQuickFix = insight.title.includes('[quickfix]') && features?.goCodeCheckerTemplates

    const quickFixUrl =
        insight.type === InsightType.SearchBased
            ? `/batch-changes/create?kind=goChecker${insight.series[0]?.name}&title=${insight.title}`
            : undefined

    const withinVirtualDashboard = !!currentDashboard && isVirtualDashboard(currentDashboard)

    return (
        <>
            <Menu>
                {({ isOpen }) => (
                    <>
                        <MenuButton
                            data-testid="InsightContextMenuButton"
                            className={classNames('p-1 ml-1 d-inline-flex', styles.button)}
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
                        <MenuList
                            position={Position.bottomStart}
                            data-testid={`context-menu.${insightID}`}
                            onKeyDown={event => event.stopPropagation()}
                        >
                            <MenuLink
                                as={Link}
                                data-testid="InsightContextMenuEditLink"
                                className={styles.item}
                                to={editUrl}
                            >
                                Edit
                            </MenuLink>

                            <MenuLink
                                data-testid="InsightContextMenuShareLink"
                                className={styles.item}
                                onSelect={() => setShowShareModal(true)}
                            >
                                Get shareable link
                            </MenuLink>

                            {menuPermissions.showYAxis && (
                                <MenuItem
                                    role="menuitemcheckbox"
                                    aria-checked={zeroYAxisMin}
                                    data-testid="InsightContextMenuEditLink"
                                    className={styles.item}
                                    onSelect={onToggleZeroYAxisMin}
                                >
                                    <Checkbox
                                        id="InsightContextMenuEditInput"
                                        aria-hidden="true"
                                        checked={zeroYAxisMin}
                                        tabIndex={-1}
                                        label={<span className="font-weight-normal">Start Y Axis at 0</span>}
                                    />
                                </MenuItem>
                            )}

                            {quickFixUrl && showQuickFix && (
                                <MenuLink as={Link} className={styles.item} to={quickFixUrl}>
                                    Golang quick fixes
                                </MenuLink>
                            )}

                            {currentDashboard && (
                                <Tooltip
                                    content={
                                        withinVirtualDashboard
                                            ? "Removing insight isn't available for the All insights dashboard"
                                            : undefined
                                    }
                                    placement="left"
                                >
                                    <MenuItem
                                        data-testid="insight-context-remove-from-dashboard-button"
                                        onSelect={() => setShowRemoveConfirm(true)}
                                        disabled={withinVirtualDashboard}
                                        className={styles.item}
                                    >
                                        Remove from this dashboard
                                    </MenuItem>
                                </Tooltip>
                            )}

                            <MenuDivider />

                            <MenuItem
                                data-testid="insight-context-menu-delete-button"
                                onSelect={() => setShowDeleteConfirm(true)}
                                className={styles.item}
                            >
                                Delete
                            </MenuItem>
                        </MenuList>
                    </>
                )}
            </Menu>
            <ConfirmDeleteModal
                insight={insight}
                showModal={showDeleteConfirm}
                onCancel={() => setShowDeleteConfirm(false)}
            />
            <ConfirmRemoveModal
                insight={insight}
                dashboard={currentDashboard}
                showModal={showRemoveConfirm}
                onCancel={() => setShowRemoveConfirm(false)}
            />
            <ShareLinkModal
                aria-label="Share insight"
                insight={insight}
                dashboards={dashboards}
                isOpen={showShareModal}
                onDismiss={() => setShowShareModal(false)}
            />
        </>
    )
}
