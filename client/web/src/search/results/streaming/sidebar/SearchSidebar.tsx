import React, { useCallback } from 'react'
import { useHistory } from 'react-router'

import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../../..'
import { submitSearch, toggleSearchFilter } from '../../../helpers'
import { Filter } from '../../../stream'

import { getDynamicFilterLinks, getRepoFilterLinks, getSnippets } from './FilterLink'
import { getQuickLinks } from './QuickLink'
import styles from './SearchSidebar.module.scss'
import { SearchSidebarSection } from './SearchSidebarSection'
import { getSearchTypeLinks } from './SearchTypeLink'

export interface SearchSidebarProps
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        SettingsCascadeProps,
        TelemetryProps {
    query: string
    filters?: Filter[]
}

export const SearchSidebar: React.FunctionComponent<SearchSidebarProps> = props => {
    const history = useHistory()

    const onFilterClicked = useCallback(
        (value: string) => {
            props.telemetryService.log('DynamicFilterClicked', {
                search_filter: { value },
            })

            const newQuery = toggleSearchFilter(props.query, value)

            submitSearch({ ...props, query: newQuery, source: 'filter', history })
        },
        [history, props]
    )

    return (
        <div className={styles.searchSidebar}>
            <SearchSidebarSection header="Search types">{getSearchTypeLinks(props)}</SearchSidebarSection>
            <SearchSidebarSection header="Dynamic filters">
                {getDynamicFilterLinks(props.filters, onFilterClicked)}
            </SearchSidebarSection>
            <SearchSidebarSection header="Repositories">
                {getRepoFilterLinks(props.filters, onFilterClicked)}
            </SearchSidebarSection>
            <SearchSidebarSection header="Search snippets">
                {getSnippets(props.settingsCascade, onFilterClicked)}
            </SearchSidebarSection>
            <SearchSidebarSection header="Quicklinks">{getQuickLinks(props.settingsCascade)}</SearchSidebarSection>
        </div>
    )
}
