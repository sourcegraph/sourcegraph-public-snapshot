import React, { useCallback, useState } from 'react'

import { mdiMenu } from '@mdi/js'
import classNames from 'classnames'

import { Link, Icon, Button } from '@sourcegraph/wildcard'

import type { BatchChangesProps } from '../batches'
import { SidebarGroup, SidebarCollapseItems, SidebarNavItem } from '../components/Sidebar'
import type { NavGroupDescriptor } from '../util/contributions'

import styles from './SiteAdminSidebar.module.scss'

export interface SiteAdminSideBarGroupContext extends BatchChangesProps {
    isSourcegraphDotCom: boolean
    isCodyApp: boolean
    codeInsightsEnabled: boolean
    endUserOnboardingEnabled: boolean
}

export interface SiteAdminSideBarGroup extends NavGroupDescriptor<SiteAdminSideBarGroupContext> {}

export type SiteAdminSideBarGroups = readonly SiteAdminSideBarGroup[]

export interface SiteAdminSidebarProps extends SiteAdminSideBarGroupContext {
    /** The items for the side bar, by group */
    groups: SiteAdminSideBarGroups
    className?: string
}

/**
 * Sidebar for the site admin area.
 */
export const SiteAdminSidebar: React.FunctionComponent<React.PropsWithChildren<SiteAdminSidebarProps>> = ({
    className,
    groups,
    ...props
}) => {
    const [isMobileExpanded, setIsMobileExpanded] = useState(false)
    const collapseMobileSidebar = useCallback((): void => setIsMobileExpanded(false), [])

    return (
        <>
            <Button className="d-sm-none align-self-start mb-3" onClick={() => setIsMobileExpanded(!isMobileExpanded)}>
                <Icon aria-hidden={true} svgPath={mdiMenu} className="mr-2" />
                {isMobileExpanded ? 'Hide' : 'Show'} menu
            </Button>
            <SidebarGroup className={classNames(className, 'd-sm-block', !isMobileExpanded && 'd-none')}>
                <ul className="list-group">
                    {groups.map(
                        ({ header, items, condition = () => true }, index) =>
                            condition(props) &&
                            (items.length > 1 ? (
                                <li className="p-0 list-group-item" key={index}>
                                    <SidebarCollapseItems
                                        icon={header?.icon}
                                        label={header?.label}
                                        openByDefault={true}
                                    >
                                        {items.map(
                                            ({ label, to, source = 'client', exact, condition = () => true }) =>
                                                condition(props) && (
                                                    <SidebarNavItem
                                                        to={to}
                                                        key={label}
                                                        source={source}
                                                        className={styles.navItem}
                                                        onClick={collapseMobileSidebar}
                                                        exact={exact}
                                                    >
                                                        {label}
                                                    </SidebarNavItem>
                                                )
                                        )}
                                    </SidebarCollapseItems>
                                </li>
                            ) : (
                                <li className="p-0 list-group-item" key={items[0].label}>
                                    <Link
                                        to={items[0].to}
                                        className="bg-2 border-0 d-flex list-group-item-action p-2 w-100"
                                        onClick={collapseMobileSidebar}
                                    >
                                        <span>
                                            {header?.icon && (
                                                <>
                                                    <Icon
                                                        className="sidebar__icon mr-1"
                                                        as={header.icon}
                                                        aria-hidden={true}
                                                    />{' '}
                                                </>
                                            )}
                                            {items[0].label}
                                        </span>
                                    </Link>
                                </li>
                            ))
                    )}
                </ul>
            </SidebarGroup>
        </>
    )
}
