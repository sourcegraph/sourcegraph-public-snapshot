import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import React from 'react'
import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../../..'

import styles from './SearchSidebar.module.scss'
import { SidebarResultTypeSection } from './SidebarResultTypeSection'

const SearchSidebarSection: React.FunctionComponent<{ header: string }> = ({ header, children }) => (
    <div>
        <h5>{header}</h5>
        <div>{children}</div>
    </div>
)

export interface SearchSidebarProps
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    query: string
}

export const SearchSidebar: React.FunctionComponent<SearchSidebarProps> = props => (
    <div className={styles.searchSidebar}>
        <SearchSidebarSection header="Result types">
            <SidebarResultTypeSection {...props} />
        </SearchSidebarSection>
        <SearchSidebarSection header="Dynamic filters" />
        <SearchSidebarSection header="Repositories" />
        <SearchSidebarSection header="Search snippets" />
        <SearchSidebarSection header="Quicklinks" />
    </div>
)
