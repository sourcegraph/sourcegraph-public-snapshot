import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { SidebarGroupHeader, SidebarGroup, SidebarNavItem } from '../../../components/Sidebar'
import { NavGroupDescriptor } from '../../../util/contributions'

export interface CodeIntelSideBarGroup extends NavGroupDescriptor<{ isSiteAdmin: boolean }> {}

export type CodeIntelSideBarGroups = readonly CodeIntelSideBarGroup[]

interface Props extends RouteComponentProps<{}> {
    codeIntelSidebarGroups: CodeIntelSideBarGroups
    className?: string
    repo: { url: string }
    isSiteAdmin: boolean
}

/** Sidebar for code intelligence pages. */
export const CodeIntelSidebar: React.FunctionComponent<Props> = ({
    codeIntelSidebarGroups,
    className,
    repo,
    isSiteAdmin,
}: Props) => (
    <div className={className}>
        {codeIntelSidebarGroups.map(
            ({ header, items, condition = () => true }, index) =>
                condition({ isSiteAdmin }) && (
                    <SidebarGroup key={index}>
                        {header && <SidebarGroupHeader label={header.label} />}
                        {items.map(
                            ({ label, to, exact, condition = () => true }) =>
                                condition({ isSiteAdmin }) && (
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
