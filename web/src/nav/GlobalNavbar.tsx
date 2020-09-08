import * as H from 'history'
import React, { useEffect, useMemo } from 'react'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { authRequired as authRequiredObservable, AuthenticatedUser } from '../auth'
import {
    parseSearchURLQuery,
    PatternTypeProps,
    InteractiveSearchProps,
    CaseSensitivityProps,
    CopyQueryButtonProps,
    OnboardingTourProps,
} from '../search'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { showDotComMarketing } from '../util/features'
import { NavLinks } from './NavLinks'
import { ThemeProps } from '../../../shared/src/theme'
import { ThemePreferenceProps } from '../theme'
import { KeyboardShortcutsProps } from '../keyboardShortcuts/keyboardShortcuts'
import { QueryState } from '../search/helpers'
import { InteractiveModeInput } from '../search/input/interactive/InteractiveModeInput'
import { FiltersToTypeAndValue } from '../../../shared/src/search/interactive/util'
import { SearchModeToggle } from '../search/input/interactive/SearchModeToggle'
import { Link } from '../../../shared/src/components/Link'
import { convertPlainTextToInteractiveQuery } from '../search/input/helpers'
import { VersionContextDropdown } from './VersionContextDropdown'
import { VersionContextProps } from '../../../shared/src/search/util'
import { VersionContext } from '../schema/site.schema'
import { TelemetryProps } from '../../../shared/src/telemetry/telemetryService'
import { useObservable } from '../../../shared/src/util/useObservable'

interface Props
    extends SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        InteractiveSearchProps,
        CopyQueryButtonProps,
        VersionContextProps,
        OnboardingTourProps {
    history: H.History
    location: H.Location<{ query: string }>
    authenticatedUser: AuthenticatedUser | null
    navbarSearchQueryState: QueryState
    onNavbarQueryChange: (queryState: QueryState) => void
    isSourcegraphDotCom: boolean
    isSearchRelatedPage: boolean
    showCampaigns: boolean

    // Whether globbing is enabled for filters.
    globbing: boolean

    /**
     * Which variation of the global navbar to render.
     *
     * 'low-profile' renders the the navbar with no border or background. Used on the search
     * homepage.
     *
     * 'low-profile-with-logo' renders the low-profile navbar but with the homepage logo. Used on repogroup pages.
     */
    variant: 'default' | 'low-profile' | 'low-profile-with-logo' | 'no-search-input'

    splitSearchModes: boolean
    interactiveSearchMode: boolean
    toggleSearchMode: (event: React.MouseEvent<HTMLAnchorElement>) => void
    setVersionContext: (versionContext: string | undefined) => void
    availableVersionContexts: VersionContext[] | undefined

    /** For testing only. Used because reactstrap's Popover is incompatible with react-test-renderer. */
    hideNavLinks: boolean
}

export const GlobalNavbar: React.FunctionComponent<Props> = ({
    isSearchRelatedPage,
    splitSearchModes,
    interactiveSearchMode,
    navbarSearchQueryState,
    versionContext,
    setVersionContext,
    availableVersionContexts,
    caseSensitive,
    patternType,
    onNavbarQueryChange,
    onFiltersInQueryChange,
    hideNavLinks,
    variant,
    isLightTheme,
    location,
    history,
    ...props
}) => {
    const authRequired = useObservable(authRequiredObservable)

    const query = useMemo(() => parseSearchURLQuery(location.search || ''), [location.search])

    useEffect(() => {
        // In interactive search mode, the InteractiveModeInput component will handle updating the inputs.
        if (!interactiveSearchMode) {
            if (query) {
                onNavbarQueryChange({ query, cursorPosition: query.length })
            } else {
                // If we have no component state, then we may have gotten unmounted during a route change.
                const query = location.state ? location.state.query : ''
                onNavbarQueryChange({
                    query,
                    cursorPosition: query.length,
                })
            }
        }

        if (query) {
            if (!isSearchRelatedPage) {
                // On a non-search related page or non-repo page, we clear the query in
                // the main query input and interactive mode UI to avoid misleading users
                // that the query is relevant in any way on those pages.
                onNavbarQueryChange({ query: '', cursorPosition: 0 })
                onFiltersInQueryChange({})
            }

            if (interactiveSearchMode) {
                let filtersInQuery: FiltersToTypeAndValue = {}
                const { filtersInQuery: newFiltersInQuery, navbarQuery } = convertPlainTextToInteractiveQuery(query)
                filtersInQuery = { ...filtersInQuery, ...newFiltersInQuery }
                onNavbarQueryChange({ query: navbarQuery, cursorPosition: navbarQuery.length })
                onFiltersInQueryChange(filtersInQuery)
            }
        }
    }, [interactiveSearchMode, isSearchRelatedPage, location, onFiltersInQueryChange, onNavbarQueryChange, query])

    let logoSource = '/.assets/img/sourcegraph-mark.svg'
    let logoLinkClassName = 'global-navbar__logo-link global-navbar__logo-animated'
    const logoWithNameSource = '/.assets/img/sourcegraph-head-logo.svg'
    const logoWithNameLightSource = '/.assets/img/sourcegraph-light-head-logo.svg'

    const branding = window.context ? window.context.branding : null
    if (branding) {
        if (isLightTheme) {
            if (branding.light?.symbol) {
                logoSource = branding.light.symbol
            }
        } else if (branding.dark?.symbol) {
            logoSource = branding.dark.symbol
        }
        if (branding.disableSymbolSpin) {
            logoLinkClassName = 'global-navbar__logo-link'
        }
    }

    const logo = <img className="global-navbar__logo" src={logoSource} />
    const logoWithNameLink = (
        <Link to="/search">
            <img
                className="global-navbar__logo-with-name pl-2"
                src={isLightTheme ? logoWithNameLightSource : logoWithNameSource}
            />
        </Link>
    )

    const logoLink = !authRequired ? (
        <Link to="/search" className={logoLinkClassName}>
            {logo}
        </Link>
    ) : (
        <div className={logoLinkClassName}>{logo}</div>
    )
    const navLinks = !authRequired && !hideNavLinks && (
        <NavLinks
            showDotComMarketing={showDotComMarketing}
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
                    {variant === 'low-profile-with-logo' && <div className="nav-item flex-1">{logoWithNameLink}</div>}
                    <div className="flex-1" />
                    {navLinks}
                </>
            ) : variant === 'no-search-input' ? (
                <>
                    {logoLink}
                    <div className="nav-item flex-1">
                        <Link to="/search" className="nav-link">
                            Search
                        </Link>
                    </div>
                    {navLinks}
                </>
            ) : (
                <>
                    {logoLink}
                    {authRequired ? (
                        <div className="flex-1" />
                    ) : splitSearchModes && interactiveSearchMode ? (
                        <InteractiveModeInput
                            {...props}
                            navbarSearchState={navbarSearchQueryState}
                            onNavbarQueryChange={onNavbarQueryChange}
                            lowProfile={!isSearchRelatedPage}
                            versionContext={versionContext}
                            location={location}
                            history={history}
                            setVersionContext={setVersionContext}
                            availableVersionContexts={availableVersionContexts}
                            isLightTheme={isLightTheme}
                            patternType={patternType}
                            caseSensitive={caseSensitive}
                            onFiltersInQueryChange={onFiltersInQueryChange}
                        />
                    ) : (
                        <div className="global-navbar__search-box-container d-none d-sm-flex flex-row">
                            {splitSearchModes && (
                                <SearchModeToggle {...props} interactiveSearchMode={interactiveSearchMode} />
                            )}
                            <VersionContextDropdown
                                history={history}
                                navbarSearchQuery={navbarSearchQueryState.query}
                                caseSensitive={caseSensitive}
                                patternType={patternType}
                                versionContext={versionContext}
                                setVersionContext={setVersionContext}
                                availableVersionContexts={availableVersionContexts}
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
