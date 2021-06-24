import classNames from 'classnames'
import * as H from 'history'
import React, { useContext, useEffect, useMemo } from 'react'
import { EMPTY, from } from 'rxjs'
import { switchMap } from 'rxjs/operators'

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
import { InsightsApiContext, InsightsViewGrid } from '../../insights'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { Settings } from '../../schema/settings.schema'
import { VersionContext } from '../../schema/site.schema'
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
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL'>,
        VersionContextProps,
        SearchContextInputProps,
        RepogroupHomepageProps,
        OnboardingTourProps,
        HomePanelsProps,
        ShowQueryBuilderProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => Promise<void>
    availableVersionContexts: VersionContext[] | undefined
    autoFocus?: boolean

    // Whether globbing is enabled for filters.
    globbing: boolean

    // Whether to additionally highlight or provide hovers for tokens, e.g., regexp character sets.
    enableSmartQuery: boolean
}

/**
 * The search page
 */
export const SearchPage: React.FunctionComponent<SearchPageProps> = props => {
    useEffect(() => props.telemetryService.logViewEvent('Home'), [props.telemetryService])

    const showCodeInsights =
        !isErrorLike(props.settingsCascade.final) &&
        !!props.settingsCascade.final?.experimentalFeatures?.codeInsights &&
        props.settingsCascade.final['insights.displayLocation.homepage'] !== false

    const { getCombinedViews } = useContext(InsightsApiContext)
    const views = useObservable(
        useMemo(
            () =>
                showCodeInsights
                    ? getCombinedViews(() =>
                          from(props.extensionsController.extHostAPI).pipe(
                              switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getHomepageViews({})))
                          )
                      )
                    : EMPTY,
            [getCombinedViews, showCodeInsights, props.extensionsController]
        )
    )
    return (
        <div className="search-page d-flex flex-column align-items-center pb-5 px-3">
            <BrandLogo className="search-page__logo" isLightTheme={props.isLightTheme} variant="logo" />
            {props.isSourcegraphDotCom && <div className="text-muted text-center mt-3">Search public code</div>}
            <div
                className={classNames('search-page__search-container', {
                    'search-page__search-container--with-content-below':
                        props.isSourcegraphDotCom || props.showEnterpriseHomePanels,
                })}
            >
                <SearchPageInput {...props} source="home" />
                {views && <InsightsViewGrid {...props} className="mt-5" views={views} />}
            </div>
            {props.isSourcegraphDotCom &&
                props.showRepogroupHomepage &&
                (!props.authenticatedUser || !props.showEnterpriseHomePanels) && <LoggedOutHomepage {...props} />}

            {props.showEnterpriseHomePanels && props.authenticatedUser && <HomePanels {...props} />}

            <SearchPageFooter className="search-page__footer" />
        </div>
    )
}
