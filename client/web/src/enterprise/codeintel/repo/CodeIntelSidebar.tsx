import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { AuthenticatedUser } from '../../../auth'
import { SidebarGroupHeader, SidebarGroup, SidebarNavItem } from '../../../components/Sidebar'
import { NavGroupDescriptor } from '../../../util/contributions'

export interface CodeIntelSideBarGroup extends NavGroupDescriptor<{ authenticatedUser: AuthenticatedUser }> {}

export type CodeIntelSideBarGroups = readonly CodeIntelSideBarGroup[]

interface Props extends RouteComponentProps<{}> {
    codeIntelSidebarGroups: CodeIntelSideBarGroups
    className?: string
    repo: { url: string }
    authenticatedUser: AuthenticatedUser
}

/** Sidebar for code intelligence pages. */
export const CodeIntelSidebar: React.FunctionComponent<Props> = ({
    codeIntelSidebarGroups,
    className,
    repo,
    authenticatedUser,
}: Props) => (
    <div className={className}>
        {codeIntelSidebarGroups.map(
            ({ header, items, condition = () => true }, index) =>
                condition({ authenticatedUser }) && (
                    <SidebarGroup key={index}>
                        {header && <SidebarGroupHeader label={header.label} />}
                        {items.map(
                            ({ label, to, exact, condition = () => true }) =>
                                condition({ authenticatedUser }) && (
                                    <SidebarNavItem
                                        to={`${repo.url}/-/code-intelligence${to}`}
                                        exact={exact}
                                        key={label}
                                    >
                                        {label}
                                    </SidebarNavItem>
                                )
                        )}
                    </SidebarGroup>
                )
        )}
    </div>
)
