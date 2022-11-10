import React, { useCallback, useState } from 'react'

import { mdiMenu } from '@mdi/js'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router-dom'

import { Button, Icon } from '@sourcegraph/wildcard'

import { SidebarGroupHeader, SidebarGroup, SidebarNavItem } from '../../components/Sidebar'
import { SettingsAreaRepositoryFields } from '../../graphql-operations'
import { NavGroupDescriptor } from '../../util/contributions'

export interface RepoSettingsSideBarGroup extends NavGroupDescriptor {}

export type RepoSettingsSideBarGroups = readonly RepoSettingsSideBarGroup[]

interface Props extends RouteComponentProps<{}> {
    repoSettingsSidebarGroups: RepoSettingsSideBarGroups
    className?: string
    repo: SettingsAreaRepositoryFields
}

/**
 * Sidebar for repository settings pages.
 */
export const RepoSettingsSidebar: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repoSettingsSidebarGroups,
    className,
    repo,
}: Props) => {
    const [isMobileExpanded, setIsMobileExpanded] = useState(false)
    const collapseMobileSidebar = useCallback((): void => setIsMobileExpanded(false), [])

    return (
        <>
            <Button className="d-sm-none align-self-start mb-3" onClick={() => setIsMobileExpanded(!isMobileExpanded)}>
                <Icon aria-hidden={true} svgPath={mdiMenu} className="mr-2" />
                {isMobileExpanded ? 'Hide' : 'Show'} menu
            </Button>
            <div className={classNames(className, 'd-sm-block', !isMobileExpanded && 'd-none')}>
                {repoSettingsSidebarGroups.map(
                    ({ header, items, condition = () => true }, index) =>
                        condition({}) && (
                            <SidebarGroup key={index}>
                                {header && <SidebarGroupHeader label={header.label} />}
                                {items.map(
                                    ({ label, to, exact, condition = () => true }) =>
                                        condition({}) && (
                                            <SidebarNavItem
                                                to={`${repo.url}/-/settings${to}`}
                                                exact={exact}
                                                key={label}
                                                onClick={collapseMobileSidebar}
                                            >
                                                {label}
                                            </SidebarNavItem>
                                        )
                                )}
                            </SidebarGroup>
                        )
                )}
            </div>
        </>
    )
}
