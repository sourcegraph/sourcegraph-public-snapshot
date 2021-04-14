import * as H from 'history'
import React, { useEffect, useMemo } from 'react'
import { of } from 'rxjs'
import { startWith } from 'rxjs/operators'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { AuthenticatedUser } from '../auth'
import { CodeMonitoringProps } from '../code-monitoring'
import { BrandLogo } from '../components/branding/BrandLogo'
import { KeyboardShortcutsProps } from '../keyboardShortcuts/keyboardShortcuts'
import { LayoutRouteProps } from '../routes'
import { VersionContext } from '../schema/site.schema'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    CopyQueryButtonProps,
    OnboardingTourProps,
    ParsedSearchQueryProps,
    SearchContextProps,
    isSearchContextSpecAvailable,
    getGlobalSearchContextFilter,
} from '../search'
import { QueryState } from '../search/helpers'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { ThemePreferenceProps } from '../theme'
import { showDotComMarketing } from '../util/features'

import { NavLinks } from './NavLinks'
import { ExtensionAlertAnimationProps } from './UserNavItem'
import { VersionContextDropdown } from './VersionContextDropdown'

interface Props
    extends SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        ThemeProps,
        ThemePreferenceProps,
        ExtensionAlertAnimationProps,
        ActivationProps,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        PatternTypeProps,
        CaseSensitivityProps,
        CopyQueryButtonProps,
        VersionContextProps,
        SearchContextProps,
        CodeMonitoringProps,
        OnboardingTourProps {
    history: H.History
    location: H.Location<{ query: string }>
    authenticatedUser: AuthenticatedUser | null
    authRequired: boolean
    navbarSearchQueryState: QueryState
    onNavbarQueryChange: (queryState: QueryState) => void
    isSourcegraphDotCom: boolean
    isSearchRelatedPage: boolean
    showBatchChanges: boolean
    routes: readonly LayoutRouteProps<{}>[]

    // Whether globbing is enabled for filters.
    globbing: boolean

    // Whether to additionally highlight or provide hovers for tokens, e.g., regexp character sets.
    enableSmartQuery: boolean

    /**
     * Which variation of the global navbar to render.
     *
     * 'low-profile' renders the the navbar with no border or background. Used on the search
     * homepage.
     *
     * 'low-profile-with-logo' renders the low-profile navbar but with the homepage logo. Used on repogroup pages.
     */
    variant: 'default' | 'low-profile' | 'low-profile-with-logo' | 'no-search-input'

    setVersionContext: (versionContext: string | undefined) => Promise<void>
    availableVersionContexts: VersionContext[] | undefined

    minimalNavLinks?: boolean
    isSearchAutoFocusRequired?: boolean
    branding?: typeof window.context.branding

    /** For testing only. Used because reactstrap's Popover is incompatible with react-test-renderer. */
    hideNavLinks: boolean
}

export const GlobalNavbar: React.FunctionComponent<Props> = ({
    authRequired,
    isSearchRelatedPage,
    navbarSearchQueryState,
    versionContext,
    setVersionContext,
    availableVersionContexts,
    caseSensitive,
    patternType,
    onNavbarQueryChange,
    hideNavLinks,
    variant,
    isLightTheme,
    branding,
    location,
    history,
    minimalNavLinks,
    ...props
}) => {
    // Workaround: can't put this in optional parameter value because of https://github.com/babel/babel/issues/11166
    branding = branding ?? window.context?.branding

    const query = props.parsedSearchQuery

    const globalSearchContextSpec = useMemo(() => getGlobalSearchContextFilter(query), [query])
    const isSearchContextAvailable = useObservable(
        useMemo(
            () =>
                globalSearchContextSpec
                    ? // While we wait for the result of the `isSearchContextSpecAvailable` call, we assume the context is available
                      // to prevent flashing and moving content in the query bar. This optimizes for the most common use case where
                      // user selects a search context from the dropdown.
                      // See https://github.com/sourcegraph/sourcegraph/issues/19918 for more info.
                      isSearchContextSpecAvailable(globalSearchContextSpec.spec).pipe(startWith(true))
                    : of(false),
            [globalSearchContextSpec]
        )
    )

    useEffect(() => {
        // On a non-search related page or non-repo page, we clear the query in
        // the main query input to avoid misleading users
        // that the query is relevant in any way on those pages.
        if (!isSearchRelatedPage) {
            onNavbarQueryChange({ query: '' })
            return
        }
        // Do nothing if there is no query in the URL
        if (!query) {
            return
        }

        // If a global search context spec is available to the user, we omit it from the
        // query and move it to the search contexts dropdown
        const finalQuery =
            globalSearchContextSpec && isSearchContextAvailable && props.showSearchContext
                ? omitFilter(query, globalSearchContextSpec.filter)
                : query

        onNavbarQueryChange({ query: finalQuery })
    }, [
        isSearchRelatedPage,
        onNavbarQueryChange,
        query,
        globalSearchContextSpec,
        isSearchContextAvailable,
        props.showSearchContext,
    ])

    const logo = (
        <LinkOrSpan to={authRequired ? undefined : '/search'} className="global-navbar__logo-link">
            <BrandLogo
                branding={branding}
                isLightTheme={isLightTheme}
                variant="symbol"
                className="global-navbar__logo"
            />
        </LinkOrSpan>
    )
    const navLinks = !authRequired && !hideNavLinks && (
        <NavLinks
            showDotComMarketing={showDotComMarketing}
            minimalNavLinks={minimalNavLinks}
            location={location}
            history={history}
            isLightTheme={isLightTheme}
            {...props}
        />
    )

    return (
        <div
            className={`global-navbar ${
                variant === 'low-profile' || variant === 'low-profile-with-logo'
                    ? ''
                    : 'global-navbar--bg border-bottom'
            } py-1`}
        >
            {variant === 'low-profile' || variant === 'low-profile-with-logo' ? (
                <>
                    {variant === 'low-profile-with-logo' && logo}
                    <div className="flex-1" />
                    {navLinks}
                </>
            ) : variant === 'no-search-input' ? (
                <>
                    {logo}
                    <div className="nav-item flex-1">
                        <Link to="/search" className="nav-link">
                            Search
                        </Link>
                    </div>
                    {navLinks}
                </>
            ) : (
                <>
                    {logo}
                    {authRequired ? (
                        <div className="flex-1" />
                    ) : (
                        <div className="global-navbar__search-box-container d-none d-sm-flex flex-row">
                            <VersionContextDropdown
                                history={history}
                                navbarSearchQuery={navbarSearchQueryState.query}
                                caseSensitive={caseSensitive}
                                patternType={patternType}
                                versionContext={versionContext}
                                setVersionContext={setVersionContext}
                                availableVersionContexts={availableVersionContexts}
                                selectedSearchContextSpec={props.selectedSearchContextSpec}
                            />
                            <SearchNavbarItem
                                {...props}
                                navbarSearchState={navbarSearchQueryState}
                                onChange={onNavbarQueryChange}
                                location={location}
                                history={history}
                                versionContext={versionContext}
                                isLightTheme={isLightTheme}
                                patternType={patternType}
                                caseSensitive={caseSensitive}
                            />
                        </div>
                    )}
                    {navLinks}
                </>
            )}
        </div>
    )
}
