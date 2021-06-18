import classNames from 'classnames'
import React, { useCallback } from 'react'
import { useHistory } from 'react-router'
import StickyBox from 'react-sticky-box'

import { Filter } from '@sourcegraph/shared/src/search/stream'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../..'
import { AuthenticatedUser } from '../../../auth'
import { FeatureFlagProps } from '../../../featureFlags/featureFlags'
import { submitSearch, toggleSearchFilter } from '../../helpers'

import { getDynamicFilterLinks, getRepoFilterLinks, getSearchSnippetLinks } from './FilterLink'
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
        TelemetryProps,
        FeatureFlagProps {
    authenticatedUser: AuthenticatedUser | null
    query: string
    filters?: Filter[]
    className?: string
}

export const SearchSidebar: React.FunctionComponent<SearchSidebarProps> = props => {
    const history = useHistory()

    const onFilterClicked = useCallback(
        (value: string) => {
            const newQuery = toggleSearchFilter(props.query, value)
            submitSearch({ ...props, query: newQuery, source: 'filter', history })
        },
        [history, props]
    )

    const onDynamicFilterClicked = useCallback(
        (value: string) => {
            props.telemetryService.log('DynamicFilterClicked', {
                search_filter: { value },
            })

            onFilterClicked(value)
        },
        [onFilterClicked, props.telemetryService]
    )

    const onSnippetClicked = useCallback(
        (value: string) => {
            props.telemetryService.log('SearchSnippetClicked')
            onFilterClicked(value)
        },
        [onFilterClicked, props.telemetryService]
    )

    const onSearchSnippetsCtaLinkClick = useCallback(() => {
        props.telemetryService.log('SignUpPLGSnippet_1_Search')
    }, [props.telemetryService])

    const showSnippetsCtaLink = !props.authenticatedUser && props.featureFlags.get('w0-signup-optimisation')

    return (
        <div className={classNames(styles.searchSidebar, props.className)}>
            <StickyBox className={styles.searchSidebarStickyBox}>
                <SearchSidebarSection className={styles.searchSidebarItem} header="Search types">
                    {getSearchTypeLinks(props)}
                </SearchSidebarSection>
                <SearchSidebarSection className={styles.searchSidebarItem} header="Dynamic filters">
                    {getDynamicFilterLinks(props.filters, onDynamicFilterClicked)}
                </SearchSidebarSection>
                <SearchSidebarSection className={styles.searchSidebarItem} header="Repositories" showSearch={true}>
                    {getRepoFilterLinks(props.filters, onDynamicFilterClicked)}
                </SearchSidebarSection>
                <SearchSidebarSection
                    className={styles.searchSidebarItem}
                    header="Search snippets"
                    ctaLinkText={showSnippetsCtaLink ? 'Sign up to create code snippets' : undefined}
                    ctaLinkTo={
                        showSnippetsCtaLink
                            ? `/sign-up?src=Snippet&returnTo=${encodeURIComponent('/user/settings')}`
                            : undefined
                    }
                    onCtaLinkClick={onSearchSnippetsCtaLinkClick}
                >
                    {getSearchSnippetLinks(props.settingsCascade, onSnippetClicked)}
                </SearchSidebarSection>
                <SearchSidebarSection className={styles.searchSidebarItem} header="Quicklinks">
                    {getQuickLinks(props.settingsCascade)}
                </SearchSidebarSection>
            </StickyBox>
        </div>
    )
}
