import React from 'react'

import { mdiArrowCollapseUp, mdiArrowExpandDown, mdiBookmarkOutline, mdiDotsHorizontal, mdiDownload } from '@mdi/js'
import classNames from 'classnames'

import { SearchPatternTypeProps } from '@sourcegraph/search'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import {
    Position,
    Menu,
    MenuButton,
    MenuList,
    MenuLink,
    Icon,
    Link,
    MenuHeader,
    MenuItem,
    Tooltip,
} from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'

import { CreateAction } from './createActions'
import { useExportSearchResultsQuery } from './useExportSearchResultsQuery'

import navStyles from './SearchResultsInfoBar.module.scss'

interface SearchActionsMenuProps extends SearchPatternTypeProps, Pick<PlatformContext, 'sourcegraphURL'> {
    query?: string
    authenticatedUser: Pick<AuthenticatedUser, 'id'> | null
    createActions: CreateAction[]
    createCodeMonitorAction: CreateAction | null
    canCreateMonitor: boolean
    resultsFound: boolean
    allExpanded: boolean
    onExpandAllResultsToggle: () => void
    onSaveQueryClick: () => void
}

export const SearchActionsMenu: React.FunctionComponent<SearchActionsMenuProps> = ({
    query = '',
    patternType,
    sourcegraphURL,
    authenticatedUser,
    createActions,
    createCodeMonitorAction,
    canCreateMonitor,
    resultsFound,
    allExpanded,
    onExpandAllResultsToggle,
    onSaveQueryClick,
}) => {
    const [requestSearchResultsExport] = useExportSearchResultsQuery({ query, patternType, sourcegraphURL })

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
                        <Icon aria-hidden={true} svgPath={mdiDotsHorizontal} />
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
                                <MenuItem onSelect={requestSearchResultsExport}>
                                    <Icon aria-hidden={true} className="mr-1" svgPath={mdiDownload} />
                                    Export Results
                                </MenuItem>
                            </>
                        )}
                        <MenuHeader>Search query</MenuHeader>
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
