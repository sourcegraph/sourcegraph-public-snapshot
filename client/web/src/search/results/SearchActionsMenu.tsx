import React from 'react'

import { mdiArrowCollapseUp, mdiArrowExpandDown, mdiBookmarkOutline, mdiChevronDown, mdiDownload } from '@mdi/js'
import classNames from 'classnames'

import type { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import {
    Icon,
    Link,
    Menu,
    MenuButton,
    MenuHeader,
    MenuItem,
    MenuLink,
    MenuList,
    Position,
    Tooltip,
} from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'

import type { CreateAction } from './createActions'

import navStyles from './SearchResultsInfoBar.module.scss'

interface SearchActionsMenuProps {
    results?: AggregateStreamingSearchResults
    authenticatedUser: Pick<AuthenticatedUser, 'id'> | null
    createActions: CreateAction[]
    createCodeMonitorAction: CreateAction | null
    canCreateMonitor: boolean
    allExpanded: boolean
    onExpandAllResultsToggle: () => void
    onSaveQueryClick: () => void
    onExportCsvClick: () => void
}

export const SearchActionsMenu: React.FunctionComponent<SearchActionsMenuProps> = ({
    results,
    authenticatedUser,
    createActions,
    createCodeMonitorAction,
    canCreateMonitor,
    allExpanded,
    onExpandAllResultsToggle,
    onSaveQueryClick,
    onExportCsvClick,
}) => {
    const resultsFound = results ? results.results.length > 0 : false

    return (
        <Menu>
            {({ isExpanded }) => (
                <li className={classNames('mr-2', navStyles.navItem)}>
                    <MenuButton
                        className={classNames('d-flex align-items-center text-decoration-none')}
                        aria-label={`${isExpanded ? 'Close' : 'Open'} search actions menu`}
                        variant="secondary"
                        outline={true}
                        size="sm"
                    >
                        Actions
                        <Icon aria-hidden={true} data-caret={true} className="ml-1" svgPath={mdiChevronDown} />
                    </MenuButton>
                    <MenuList tabIndex={0} position={Position.bottomEnd} aria-label="Search Actions. Open menu">
                        {resultsFound && (
                            <>
                                <MenuHeader>Search results</MenuHeader>
                                <MenuItem onSelect={onExpandAllResultsToggle}>
                                    <Icon
                                        aria-hidden={true}
                                        className="mr-1"
                                        svgPath={allExpanded ? mdiArrowCollapseUp : mdiArrowExpandDown}
                                    />
                                    {allExpanded ? 'Collapse all' : 'Expand all'}
                                </MenuItem>
                                <MenuItem onSelect={onExportCsvClick}>
                                    <Icon aria-hidden={true} className="mr-1" svgPath={mdiDownload} />
                                    Export results
                                </MenuItem>
                            </>
                        )}
                        <MenuHeader>Search query</MenuHeader>
                        {createActions.map(createAction => (
                            <Tooltip content={createAction.tooltip} key={createAction.label} placement="left">
                                <MenuLink as={Link} to={createAction.url} disabled={createAction.disabled}>
                                    <Icon
                                        aria-hidden="true"
                                        className="mr-1"
                                        {...(typeof createAction.icon === 'string'
                                            ? { svgPath: createAction.icon }
                                            : { as: createAction.icon })}
                                    />
                                    {createAction.label}
                                </MenuLink>
                            </Tooltip>
                        ))}
                        {createCodeMonitorAction && (
                            <Tooltip
                                content={
                                    authenticatedUser && !canCreateMonitor
                                        ? 'Code monitors only support type:diff or type:commit searches.'
                                        : undefined
                                }
                                placement="left"
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
                                    Create monitor
                                </MenuLink>
                            </Tooltip>
                        )}
                        <MenuItem onSelect={onSaveQueryClick} disabled={!authenticatedUser}>
                            <Icon aria-hidden={true} className="mr-1" svgPath={mdiBookmarkOutline} />
                            Save search
                        </MenuItem>
                    </MenuList>
                </li>
            )}
        </Menu>
    )
}
