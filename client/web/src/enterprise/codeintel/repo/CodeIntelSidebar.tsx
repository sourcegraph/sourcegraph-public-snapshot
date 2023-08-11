import type { FC } from 'react'

import { SidebarGroupHeader, SidebarGroup, SidebarNavItem } from '../../../components/Sidebar'
import type { NavGroupDescriptor } from '../../../util/contributions'

export interface CodeIntelSideBarGroup extends NavGroupDescriptor {}

export type CodeIntelSideBarGroups = readonly CodeIntelSideBarGroup[]

interface Props {
    codeIntelSidebarGroups: CodeIntelSideBarGroups
    className?: string
    repo: { url: string }
}

/** Sidebar for code navigation pages. */
export const CodeIntelSidebar: FC<Props> = ({ codeIntelSidebarGroups, className, repo }) => (
    <div className={className}>
        {codeIntelSidebarGroups.map(({ header, items }, index) => (
            <SidebarGroup key={index}>
                {header && <SidebarGroupHeader label={header.label} />}
                {items.map(({ label, to }) => (
                    <SidebarNavItem to={`${repo.url}/-/code-graph${to}`} key={label}>
                        {label}
                    </SidebarNavItem>
                ))}
            </SidebarGroup>
        ))}
    </div>
)
