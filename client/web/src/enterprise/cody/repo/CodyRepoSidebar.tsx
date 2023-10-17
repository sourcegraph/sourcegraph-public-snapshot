import type { FC } from 'react'

import { SidebarGroupHeader, SidebarGroup, SidebarNavItem } from '../../../components/Sidebar'
import type { NavGroupDescriptor } from '../../../util/contributions'

export interface CodyRepoSideBarGroup extends NavGroupDescriptor {}

export type CodyRepoSidebarGroups = readonly CodyRepoSideBarGroup[]

interface Props {
    codyRepoSidebarGroups: CodyRepoSidebarGroups
    className?: string
    repo: { url: string }
}

export const CodyRepoSidebar: FC<Props> = ({ codyRepoSidebarGroups, className, repo }) => (
    <div className={className}>
        {codyRepoSidebarGroups.map(({ header, items }, index) => (
            <SidebarGroup key={index}>
                {header && <SidebarGroupHeader label={header.label} />}
                {items.map(({ label, to }) => (
                    <SidebarNavItem to={`${repo.url}/-/embeddings${to}`} key={label}>
                        {label}
                    </SidebarNavItem>
                ))}
            </SidebarGroup>
        ))}
    </div>
)
