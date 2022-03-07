import classNames from 'classnames'
import * as H from 'history'
import React, { useEffect, useMemo } from 'react'

import { SearchContextInputProps } from '@sourcegraph/search'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { KeyboardShortcutsProps } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { HomePanelsProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { BrandLogo } from '../../components/branding/BrandLogo'
import { FeatureFlagProps } from '../../featureFlags/featureFlags'
import { CodeInsightsProps } from '../../insights/types'
import { useExperimentalFeatures, useNavbarQueryState } from '../../stores'
import { ThemePreferenceProps } from '../../theme'
import { HomePanels } from '../panels/HomePanels'

import { LoggedOutHomepage } from './LoggedOutHomepage'
import styles from './SearchPage.module.scss'
import { SearchPageFooter } from './SearchPageFooter'
import { SearchPageInput } from './SearchPageInput'

export interface SearchPageProps
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'>,
        PlatformContextProps<
            'forceUpdateTooltip' | 'settings' | 'sourcegraphURL' | 'updateSettings' | 'requestGraphQL'
        >,
        SearchContextInputProps,
        HomePanelsProps,
        CodeInsightsProps,
        FeatureFlagProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    autoFocus?: boolean

    // Whether globbing is enabled for filters.
    globbing: boolean
}

/**
 * The search page
 */
export const SearchPage: React.FunctionComponent<SearchPageProps> = props => {
    const showEnterpriseHomePanels = useExperimentalFeatures(features => features.showEnterpriseHomePanels ?? false)

    const isExperimentalOnboardingTourEnabled = useExperimentalFeatures(
        features => features.showOnboardingTour ?? false
    )
    const hasSearchQuery = useNavbarQueryState(state => state.searchQueryFromURL !== '')
    const isGettingStartedTourEnabled = props.featureFlags.get('getting-started-tour')
    const showOnboardingTour = useMemo(
        () => isExperimentalOnboardingTourEnabled && !hasSearchQuery && !isGettingStartedTourEnabled,
        [hasSearchQuery, isGettingStartedTourEnabled, isExperimentalOnboardingTourEnabled]
    )
    const showCollaborators = useExperimentalFeatures(features => features.homepageUserInvitation) ?? false

    useEffect(() => props.telemetryService.logViewEvent('Home'), [props.telemetryService])

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', styles.searchPage)}>
            <BrandLogo className={styles.logo} isLightTheme={props.isLightTheme} variant="logo" />
            {props.isSourcegraphDotCom && (
                <div className="text-muted text-center font-italic mt-3">
                    Search your code and 2M+ open source repositories
                </div>
            )}
            <div
                className={classNames(styles.searchContainer, {
                    [styles.searchContainerWithContentBelow]: props.isSourcegraphDotCom || showEnterpriseHomePanels,
                })}
            >
                <SearchPageInput {...props} showOnboardingTour={showOnboardingTour} source="home" />
            </div>
            <div
                className={classNames(styles.panelsContainer, {
                    [styles.panelsContainerWithCollaborators]: showCollaborators,
                })}
            >
                {props.isSourcegraphDotCom && !props.authenticatedUser && <LoggedOutHomepage {...props} />}

                {showEnterpriseHomePanels && props.authenticatedUser && (
                    <HomePanels showCollaborators={showCollaborators} {...props} />
                )}
            </div>

            <SearchPageFooter {...props} />
        </div>
    )
}
