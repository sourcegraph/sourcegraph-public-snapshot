import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useContext, useEffect, useMemo } from 'react'
import { EMPTY, from } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { Link } from '@sourcegraph/shared/src/components/Link'
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
    CopyQueryButtonProps,
    RepogroupHomepageProps,
    OnboardingTourProps,
    HomePanelsProps,
    ShowQueryBuilderProps,
    ParsedSearchQueryProps,
    SearchContextProps,
} from '..'
import { AuthenticatedUser } from '../../auth'
import { BrandLogo } from '../../components/branding/BrandLogo'
import { SyntaxHighlightedSearchQuery } from '../../components/SyntaxHighlightedSearchQuery'
import { InsightsApiContext, InsightsViewGrid } from '../../insights'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { repogroupList } from '../../repogroups/HomepageConfig'
import { Settings } from '../../schema/settings.schema'
import { VersionContext } from '../../schema/site.schema'
import { ThemePreferenceProps } from '../../theme'
import { HomePanels } from '../panels/HomePanels'

import { LiteralSearchIllustration } from './LiteralSearchIllustration'
import { SearchPageFooter } from './SearchPageFooter'
import { SearchPageInput } from './SearchPageInput'
import { SignUpCta } from './SignUpCta'

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
        CopyQueryButtonProps,
        VersionContextProps,
        Omit<
            SearchContextProps,
            'convertVersionContextToSearchContext' | 'isSearchContextSpecAvailable' | 'fetchSearchContext'
        >,
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

const exampleQueries = [
    { query: 'repo:^github\\.com/sourcegraph/sourcegraph$@3.17 CONTAINER_ID', patternType: 'literal' },
    { query: 'repo:sourcegraph/sourcegraph type:diff after:"1 week ago"', patternType: 'literal' },
    {
        query: 'lang:TypeScript useState OR useMemo',
        patternType: 'literal',
    },
    { query: 'lang:Python return :[v.], :[v.]', patternType: 'structural' },
]

/**
 * The search page
 */
export const SearchPage: React.FunctionComponent<SearchPageProps> = props => {
    const SearchExampleClicked = useCallback(
        (url: string) => (): void => props.telemetryService.log('ExampleSearchClicked', { url }),
        [props.telemetryService]
    )

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
        <div className="web-content search-page d-flex flex-column align-items-center pb-5 px-3">
            <BrandLogo className="search-page__logo" isLightTheme={props.isLightTheme} variant="logo" />
            {props.isSourcegraphDotCom && (
                <div className="text-muted mt-3">Public code search, insights, and automation</div>
            )}
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
                (!props.authenticatedUser || !props.showEnterpriseHomePanels) && (
                    <>
                        <div className="search-page__repogroup-content">
                            <div className="search-page__help-content">
                                <div className="search-page__help-content-example-searches mr-2">
                                    <h3 className="search-page__help-content-header my-3">Example searches</h3>
                                    <div className="mt-2">
                                        {exampleQueries.map(example => (
                                            <div key={example.query} className="pb-2">
                                                <Link
                                                    to={`/search?q=${encodeURIComponent(example.query)}&patternType=${
                                                        example.patternType
                                                    }`}
                                                    className="search-query-link text-monospace mb-2"
                                                    onClick={SearchExampleClicked(
                                                        `/search?q=${encodeURIComponent(example.query)}&patternType=${
                                                            example.patternType
                                                        }`
                                                    )}
                                                >
                                                    <SyntaxHighlightedSearchQuery query={example.query} />
                                                </Link>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                                <div>
                                    <h3 className="search-page__help-content-header my-3">Search basics</h3>
                                    <div className="mt-2">
                                        <div className="mb-2">
                                            Search for code without escaping.{' '}
                                            <span className="search-page__inline-code text-code bg-code p-1">
                                                console.log("
                                            </span>{' '}
                                            results in:
                                        </div>
                                        <LiteralSearchIllustration />
                                    </div>
                                </div>
                            </div>

                            <div className="mt-5 d-flex justify-content-center">
                                <div className="d-flex align-items-center search-page__cta">
                                    <SignUpCta />
                                    <div className="mt-2">
                                        Prefer a local installation?{' '}
                                        <a
                                            href="https://docs.sourcegraph.com"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                        >
                                            Install Sourcegraph locally.
                                        </a>
                                    </div>
                                </div>
                            </div>

                            <div className="mt-5">
                                <div className="d-flex align-items-baseline mt-5 mb-3">
                                    <h3 className="search-page__help-content-header mr-2">Repository pages</h3>
                                    <small className="text-monospace font-weight-normal small">
                                        <span className="search-filter-keyword">repogroup:</span>
                                        <i>name</i>
                                    </small>
                                </div>
                                <div className="search-page__repogroup-list-cards">
                                    {repogroupList.map(repogroup => (
                                        <div className="d-flex align-items-center" key={repogroup.name}>
                                            <img
                                                className="search-page__repogroup-list-icon mr-2"
                                                src={repogroup.homepageIcon}
                                                alt={`${repogroup.name} icon`}
                                            />
                                            <Link
                                                to={repogroup.url}
                                                className="search-page__repogroup-listing-title font-weight-bold"
                                            >
                                                {repogroup.title}
                                            </Link>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </>
                )}

            {props.showEnterpriseHomePanels && props.authenticatedUser && <HomePanels {...props} />}

            <SearchPageFooter className="search-page__footer" />
        </div>
    )
}
