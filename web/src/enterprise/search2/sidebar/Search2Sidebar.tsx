import React from 'react'
import { Search2SidebarFacet } from './facets/Search2SidebarFacet'

interface Props {
    className?: string
}

export const Search2Sidebar: React.FunctionComponent<Props> = ({ className = '' }) => (
    <nav className={`Search2Sidebar ${className} card border-0`}>
        <Search2SidebarFacet title="Calls to" value="errors.WithMessage" />
        <Search2SidebarFacet title="Arguments" value="asdf" />
        <Search2SidebarFacet title="In repository" value="" />
        <Search2SidebarFacet title="By author" value="" />
        <Search2SidebarFacet title="Date range" value="" />
        <Search2SidebarFacet title="Nearby text" value="yaml" />
    </nav>
)
