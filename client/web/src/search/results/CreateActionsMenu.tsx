import React from 'react'

import { mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { Position, Menu, MenuButton, MenuList, MenuLink, Icon, Link, Tooltip } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'

import type { CreateAction } from './createActions'

import createActionsStyles from './CreateActions.module.scss'
import navStyles from './SearchResultsInfoBar.module.scss'

export interface CreateActionsMenuProps {
    createActions: CreateAction[]
    createCodeMonitorAction: CreateAction | null
    canCreateMonitor: boolean
    authenticatedUser: Pick<AuthenticatedUser, 'id'> | null
}

export const CreateActionsMenu: React.FunctionComponent<CreateActionsMenuProps> = ({
    createActions,
    createCodeMonitorAction,
    canCreateMonitor,
    authenticatedUser,
}) => (
    <Menu>
        {({ isExpanded }) => (
            <>
                <li className={classNames('mr-2', createActionsStyles.menu, navStyles.navItem)}>
                    <MenuButton
                        className={classNames('d-flex align-items-center text-decoration-none')}
                        aria-label={`${isExpanded ? 'Close' : 'Open'} create actions menu`}
                        variant="secondary"
                        outline={true}
                        size="sm"
                    >
                        <Icon aria-hidden={true} className="mr-1" svgPath={mdiPlus} />
                        Create …
                    </MenuButton>
                </li>
                <MenuList tabIndex={0} position={Position.bottomStart} aria-label="Create Actions. Open menu">
                    {createActions.map(createAction => (
                        <MenuLink key={createAction.label} as={Link} to={createAction.url}>
                            <Icon
                                aria-hidden="true"
                                className="mr-1"
                                {...(typeof createAction.icon === 'string'
                                    ? { svgPath: createAction.icon }
                                    : { as: createAction.icon })}
                            />
                            {createAction.label}
                        </MenuLink>
                    ))}
                    {createCodeMonitorAction && (
                        <Tooltip
                            content={
                                authenticatedUser && !canCreateMonitor
                                    ? 'Code monitors only support type:diff or type:commit searches.'
                                    : undefined
                            }
                        >
                            <MenuLink
                                as={Link}
                                disabled={!authenticatedUser || !canCreateMonitor}
                                to={createCodeMonitorAction.url}
                            >
                                <Icon
                                    aria-hidden={true}
                                    className="mr-1"
                                    {...(typeof createCodeMonitorAction.icon === 'string'
                                        ? { svgPath: createCodeMonitorAction.icon }
                                        : { as: createCodeMonitorAction.icon })}
                                />
                                Create Monitor
                            </MenuLink>
                        </Tooltip>
                    )}
                </MenuList>
            </>
        )}
    </Menu>
)
