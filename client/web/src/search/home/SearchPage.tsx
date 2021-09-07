import classNames from 'classnames'
import * as H from 'history'
import React, { useEffect, useMemo } from 'react'
import { EMPTY, from } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import {
    PatternTypeProps,
    CaseSensitivityProps,
    RepogroupHomepageProps,
    OnboardingTourProps,
    HomePanelsProps,
    ShowQueryBuilderProps,
    ParsedSearchQueryProps,
    SearchContextInputProps,
} from '..'
import { AuthenticatedUser } from '../../auth'
import { BrandLogo } from '../../components/branding/BrandLogo'
import { FeatureFlagProps } from '../../featureFlags/featureFlags'
import { SmartInsight } from '../../insights/components/insights-view-grid/components/smart-insight/SmartInsight'
import { createExtensionInsight } from '../../insights/core/backend/utils/create-extension-insight'
import { useAllInsights } from '../../insights/hooks/use-insight/use-insight'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { Settings } from '../../schema/settings.schema'
import { VersionContext } from '../../schema/site.schema'
import { ThemePreferenceProps } from '../../theme'
import { StaticView, ViewGrid } from '../../views'
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
        VersionContextProps,
        SearchContextInputProps,
        RepogroupHomepageProps,
        OnboardingTourProps,
        HomePanelsProps,
        ShowQueryBuilderProps,
        FeatureFlagProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => Promise<void>
    availableVersionContexts: VersionContext[] | undefined
    autoFocus?: boolean

    // Whether globbing is enabled for filters.
    globbing: boolean
}

/**
 * The search page
 */
export const SearchPage: React.FunctionComponent<SearchPageProps> = props => {
    useEffect(() => props.telemetryService.logViewEvent('Home'), [props.telemetryService])

    const showCodeInsights =
        !isErrorLike(props.settingsCascade.final) &&
        !!props.settingsCascade.final?.experimentalFeatures?.codeInsights &&
        props.settingsCascade.final['insights.displayLocation.homepage'] === true

    const insights = useAllInsights({ settingsCascade: props.settingsCascade })
    const views = useObservable(
        useMemo(
            () =>
                showCodeInsights
                    ? from(props.extensionsController.extHostAPI).pipe(
                          switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getHomepageViews({}))),
                          map(extensionViews => extensionViews.map(createExtensionInsight))
                      )
                    : EMPTY,
            [showCodeInsights, props.extensionsController]
        )
    )

    const extensionViews = views ?? []
    const allViewIds = useMemo(() => [...(views ?? []), ...insights].map(view => view.id), [views, insights])

    return (
        <div className="search-page d-flex flex-column align-items-center px-3">
            <BrandLogo className="search-page__logo" isLightTheme={props.isLightTheme} variant="logo" />
            {props.isSourcegraphDotCom && (
                <div className="text-muted text-center font-italic mt-3">
                    Search your code and 1M+ open source repositories
                </div>
            )}
            <div
                className={classNames('search-page__search-container', {
                    'search-page__search-container--with-content-below':
                        props.isSourcegraphDotCom || props.showEnterpriseHomePanels,
                })}
            >
                <SearchPageInput {...props} source="home" />
                {showCodeInsights && (
                    <ViewGrid viewIds={allViewIds} telemetryService={props.telemetryService} className="mt-5">
                        {/* Render extension views for the search page */}
                        {extensionViews.map(view => (
                            <StaticView key={view.id} view={view} telemetryService={props.telemetryService} />
                        ))}

                        {/* Render all code insights with proper directory page context */}
                        {insights.map(insight => (
                            <SmartInsight
                                key={insight.id}
                                insight={insight}
                                telemetryService={props.telemetryService}
                                platformContext={props.platformContext}
                                settingsCascade={props.settingsCascade}
                                where="homepage"
                                context={{}}
                            />
                        ))}
                    </ViewGrid>
                )}
            </div>
            <div className="flex-grow-1">
                {props.isSourcegraphDotCom &&
                    props.showRepogroupHomepage &&
                    (!props.authenticatedUser || !props.showEnterpriseHomePanels) && <LoggedOutHomepage {...props} />}

                {props.showEnterpriseHomePanels && props.authenticatedUser && <HomePanels {...props} />}
            </div>

            <SearchPageFooter {...props} />
        </div>
    )
}
