import React from 'react'

import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../../..'

import { getQuickLinks } from './QuickLink'
import styles from './SearchSidebar.module.scss'
import { SearchSidebarSection } from './SearchSidebarSection'
import { getSearchTypeLinks } from './SearchTypeLink'

export interface SearchSidebarProps
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        SettingsCascadeProps {
    query: string
}

export const SearchSidebar: React.FunctionComponent<SearchSidebarProps> = props => (
    <div className={styles.searchSidebar}>
        <SearchSidebarSection header="Search types">{getSearchTypeLinks(props)}</SearchSidebarSection>
        <SearchSidebarSection header="Dynamic filters" />
        <SearchSidebarSection header="Repositories" />
        <SearchSidebarSection header="Search snippets" />
        <SearchSidebarSection header="Quicklinks">{getQuickLinks(props.settingsCascade)}</SearchSidebarSection>
    </div>
)
