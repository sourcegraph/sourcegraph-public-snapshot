import * as H from 'history'
import BarChartIcon from 'mdi-react/BarChartIcon'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import React, { useEffect, useMemo } from 'react'
import { of } from 'rxjs'
import { startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { ContributableMenu } from '@sourcegraph/shared/src/api/protocol'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ActivationDropdown } from '@sourcegraph/shared/src/components/activation/ActivationDropdown'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { WebCommandListPopoverButton } from '@sourcegraph/web/src/components/shared'
import { FeedbackPrompt } from '@sourcegraph/web/src/nav/Feedback/FeedbackPrompt'
import { StatusMessagesNavItem } from '@sourcegraph/web/src/nav/StatusMessagesNavItem'
import { NavGroup, NavItem, NavBar, NavLink, NavActions, NavAction } from '@sourcegraph/wildcard'
import { Box } from '@sourcegraph/wildcard/src/components/Layout/Box/Box'
import { Columns, Column } from '@sourcegraph/wildcard/src/components/Layout/Columns/Columns'
import { Inline } from '@sourcegraph/wildcard/src/components/Layout/Inline/Inline'
import { Stack } from '@sourcegraph/wildcard/src/components/Layout/Stack/Stack'

import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { BatchChangesNavItem } from '../batches/BatchChangesNavItem'
import { CodeMonitoringProps } from '../code-monitoring'
import { CodeMonitoringLogo } from '../code-monitoring/CodeMonitoringLogo'
import { BrandLogo } from '../components/branding/BrandLogo'
import {
    KeyboardShortcutsProps,
    KEYBOARD_SHORTCUT_SHOW_COMMAND_PALETTE,
    KEYBOARD_SHORTCUT_SWITCH_THEME,
} from '../keyboardShortcuts/keyboardShortcuts'
import { LayoutRouteProps } from '../routes'
import { Settings } from '../schema/settings.schema'
import { VersionContext } from '../schema/site.schema'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    OnboardingTourProps,
    ParsedSearchQueryProps,
    isSearchContextSpecAvailable,
    getGlobalSearchContextFilter,
    SearchContextInputProps,
} from '../search'
import { QueryState } from '../search/helpers'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { ThemePreferenceProps } from '../theme'
import { userExternalServicesEnabledFromTags } from '../user/settings/cloud-ga'
import { UserSettingsSidebarItems } from '../user/settings/UserSettingsSidebar'
import { showDotComMarketing } from '../util/features'

import { ExtensionAlertAnimationProps, UserNavItem } from './UserNavItem'

interface Props
    extends SettingsCascadeProps<Settings>,
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
        VersionContextProps,
        SearchContextInputProps,
        CodeMonitoringProps,
        OnboardingTourProps,
        BatchChangesProps {
    history: H.History
    location: H.Location<{ query: string }>
    authenticatedUser: AuthenticatedUser | null
    authRequired: boolean
    navbarSearchQueryState: QueryState
    onNavbarQueryChange: (queryState: QueryState) => void
    isSourcegraphDotCom: boolean
    showSearchBox: boolean
    routes: readonly LayoutRouteProps<{}>[]

    // Whether globbing is enabled for filters.
    globbing: boolean

    userSettingsSideBarItems?: UserSettingsSidebarItems

    /**
     * Which variation of the global navbar to render.
     *
     * 'low-profile' renders the the navbar with no border or background. Used on the search
     * homepage.
     *
     * 'low-profile-with-logo' renders the low-profile navbar but with the homepage logo. Used on repogroup pages.
     */
    variant: 'default' | 'low-profile' | 'low-profile-with-logo'

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
    showSearchBox,
    navbarSearchQueryState,
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
    isSourcegraphDotCom,
    ...props
}) => {
    // Workaround: can't put this in optional parameter value because of https://github.com/babel/babel/issues/11166
    branding = branding ?? window.context?.branding

    const query = props.parsedSearchQuery

    const globalSearchContextSpec = useMemo(() => getGlobalSearchContextFilter(query), [query])

    // UI includes repositories section as part of the user navigation bar
    // This filter makes sure repositories feature flag is active.
    const showRepositorySection = useMemo(
        () => !!props.userSettingsSideBarItems?.find(item => item.label === 'Repositories'),
        [props.userSettingsSideBarItems]
    )

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
        if (!showSearchBox) {
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
        showSearchBox,
        onNavbarQueryChange,
        query,
        globalSearchContextSpec,
        isSearchContextAvailable,
        props.showSearchContext,
    ])

    const settings = !isErrorLike(props.settingsCascade.final) ? props.settingsCascade.final : null
    const codeInsights =
        settings?.experimentalFeatures?.codeInsights && settings?.['insights.displayLocation.insightsPage'] !== false

    const searchNavBar = (
        <SearchNavbarItem
            {...props}
            navbarSearchState={navbarSearchQueryState}
            onChange={onNavbarQueryChange}
            location={location}
            history={history}
            isLightTheme={isLightTheme}
            patternType={patternType}
            caseSensitive={caseSensitive}
            isSourcegraphDotCom={isSourcegraphDotCom}
        />
    )

    return (
        <Stack align="left" space="gutter" dividers={false}>
            <Box paddingLeft="gutter" paddingRight="gutter">
                <Columns>
                    <Column>
                        <Box paddingRight="gutter">
                            <BrandLogo
                                branding={branding}
                                isLightTheme={isLightTheme}
                                variant="symbol"
                                className="global-navbar__logo"
                            />
                        </Box>
                        <Box component="nav" gap="gutter" flexDirection="row">
                            <Inline component="ul" space="gutter" alignY="center">
                                <Box component="li">Home</Box>
                                <Box component="li">Contact</Box>
                                <Box component="li">Info</Box>
                            </Inline>
                        </Box>
                    </Column>
                    <Column width="content">
                        <Box component="nav" gap="gutter" flexDirection="row">
                            <Inline component="ul" space="gutter">
                                <Box component="li">Feedback</Box>
                                <Box component="li">Settings</Box>
                                <Box component="li">Sign Out</Box>
                            </Inline>
                        </Box>
                    </Column>
                </Columns>
            </Box>
        </Stack>
    )
}
