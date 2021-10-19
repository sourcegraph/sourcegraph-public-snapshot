import classNames from 'classnames'
import * as H from 'history'
import React, { useEffect } from 'react'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import {
    PatternTypeProps,
    CaseSensitivityProps,
    OnboardingTourProps,
    HomePanelsProps,
    ParsedSearchQueryProps,
    SearchContextInputProps,
} from '..'
import { AuthenticatedUser } from '../../auth'
import { BrandLogo } from '../../components/branding/BrandLogo'
import { FeatureFlagProps } from '../../featureFlags/featureFlags'
import { CodeInsightsProps } from '../../insights/types'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { Settings } from '../../schema/settings.schema'
import { ThemePreferenceProps } from '../../theme'
import { HomePanels } from '../panels/HomePanels'

import { LoggedOutHomepage } from './LoggedOutHomepage'
import { SearchPageFooter } from './SearchPageFooter'
import { SearchPageInput } from './SearchPageInput'

export interface SearchPageProps
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        PatternTypeProps,
        CaseSensitivityProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL' | 'updateSettings'>,
        SearchContextInputProps,
        OnboardingTourProps,
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
    const { extensionViews: ExtensionViewsSection } = props
    useEffect(() => props.telemetryService.logViewEvent('Home'), [props.telemetryService])

    return (
        <div className="search-page d-flex flex-column align-items-center px-3">
            <BrandLogo className="search-page__logo" isLightTheme={props.isLightTheme} variant="logo" />
            {props.isSourcegraphDotCom && (
                <div className="text-muted text-center font-italic mt-3">
                    Search your code and 2M+ open source repositories
                </div>
            )}
            <div
                className={classNames('search-page__search-container', {
                    'search-page__search-container--with-content-below':
                        props.isSourcegraphDotCom || props.showEnterpriseHomePanels,
                })}
            >
                <SearchPageInput {...props} source="home" />
                <ExtensionViewsSection
                    className="mt-5"
                    telemetryService={props.telemetryService}
                    extensionsController={props.extensionsController}
                    platformContext={props.platformContext}
                    settingsCascade={props.settingsCascade}
                    where="homepage"
                />
            </div>
            <div className="flex-grow-1">
                {props.isSourcegraphDotCom && !props.authenticatedUser && <LoggedOutHomepage {...props} />}

                {props.showEnterpriseHomePanels && props.authenticatedUser && <HomePanels {...props} />}
            </div>

            <SearchPageFooter {...props} />
        </div>
    )
}
