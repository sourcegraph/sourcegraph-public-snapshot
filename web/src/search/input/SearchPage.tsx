import * as H from 'history'
import React, { useEffect, useMemo } from 'react'
import {
    PatternTypeProps,
    InteractiveSearchProps,
    CaseSensitivityProps,
    SmartSearchFieldProps,
    CopyQueryButtonProps,
    RepogroupHomepageProps,
} from '..'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { Settings } from '../../schema/settings.schema'
import { ThemeProps } from '../../../../shared/src/theme'
import { eventLogger, EventLoggerProps } from '../../tracking/eventLogger'
import { ThemePreferenceProps } from '../../theme'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { Link } from '../../../../shared/src/components/Link'
import { BrandLogo } from '../../components/branding/BrandLogo'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { VersionContext } from '../../schema/site.schema'
import { ViewGrid } from '../../repo/tree/ViewGrid'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { getViewsForContainer } from '../../../../shared/src/api/client/services/viewService'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { ContributableViewContainer } from '../../../../shared/src/api/protocol'
import { EMPTY } from 'rxjs'
import classNames from 'classnames'
import { repogroupList, homepageLanguageList } from '../../repogroups/HomepageConfig'
import { SearchPageInput } from './SearchPageInput'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'

interface Props
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        KeyboardShortcutsProps,
        EventLoggerProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        InteractiveSearchProps,
        SmartSearchFieldProps,
        CopyQueryButtonProps,
        VersionContextProps,
        RepogroupHomepageProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => void
    availableVersionContexts: VersionContext[] | undefined

    // For NavLinks
    authRequired?: boolean
    showCampaigns: boolean
}

/**
 * The search page
 */
export const SearchPage: React.FunctionComponent<Props> = props => {
    useEffect(() => eventLogger.logViewEvent('Home'))

    const codeInsightsEnabled =
        !isErrorLike(props.settingsCascade.final) && !!props.settingsCascade.final?.experimentalFeatures?.codeInsights

    const views = useObservable(
        useMemo(
            () =>
                codeInsightsEnabled
                    ? getViewsForContainer(
                          ContributableViewContainer.Homepage,
                          {},
                          props.extensionsController.services.view
                      )
                    : EMPTY,
            [codeInsightsEnabled, props.extensionsController.services.view]
        )
    )

    return (
        <div className="search-page">
            <BrandLogo className="search-page__logo" isLightTheme={props.isLightTheme} />
            <div className="search-page__cloud-tag-line">Search public code</div>
            <div
                className={classNames('search-page__container', {
                    'search-page__container--with-repogroups': props.isSourcegraphDotCom,
                })}
            >
                <SearchPageInput {...props} source="home" />
                {views && <ViewGrid {...props} className="mt-5" views={views} />}
            </div>
            {props.isSourcegraphDotCom && props.showRepogroupHomepage && (
                <div className="search-page__repogroup-content mt-5">
                    <div className="d-flex align-items-baseline mb-3">
                        <h3 className="search-page__help-content-header mr-2">Search in repository groups</h3>
                        <span className="text-monospace font-weight-normal search-page__lang-ref">
                            <span className="repogroup-page__keyword-text">repogroup:</span>
                            <i>name</i>
                        </span>
                    </div>
                    <div className="search-page__repogroup-list-cards">
                        {repogroupList.map(repogroup => (
                            <div className="d-flex" key={repogroup.name}>
                                <img className="search-page__repogroup-list-icon mr-2" src={repogroup.homepageIcon} />
                                <div className="d-flex flex-column">
                                    <Link
                                        to={repogroup.url}
                                        className="search-page__repogroup-listing-title search-page__web-link font-weight-bold"
                                    >
                                        {repogroup.title}
                                    </Link>
                                    <p className="search-page__repogroup-listing-description">
                                        {repogroup.homepageDescription}
                                    </p>
                                </div>
                            </div>
                        ))}
                    </div>
                    <div className="search-page__help-content mt-5">
                        <div>
                            <h3 className="search-page__help-content-header">Example searches</h3>
                            <ul className="list-group-flush p-0">
                                <li className="list-group-item px-0 pb-3">
                                    <Link
                                        to="/search?q=lang:javascript+alert%28:%5Bvariable%5D%29&patternType=structural"
                                        className="text-monospace mb-1"
                                    >
                                        <span className="repogroup-page__keyword-text">lang:</span>javascript
                                        alert(:[variable])
                                    </Link>{' '}
                                    <div>Find usages of the alert() method that displays an alert box.</div>
                                </li>
                                <li className="list-group-item px-0 py-3">
                                    <Link
                                        to="/search?q=lang:python+from+%5CB%5C.%5Cw%2B+import+%5Cw%2B&patternType=regexp"
                                        className="text-monospace mb-1"
                                    >
                                        <span className="repogroup-page__keyword-text">lang:</span>python from \B\.\w+
                                        import \w+
                                    </Link>{' '}
                                    <div>
                                        Search for explicit imports with one or more leading dots that indicate current
                                        and parent packages involved.
                                    </div>
                                </li>
                                <li className="list-group-item px-0 py-3">
                                    <Link
                                        to='/search?q=repo:%5Egithub%5C.com/golang/go%24+type:diff+after:"1+week+ago"&patternType=literal"'
                                        className="text-monospace mb-1"
                                    >
                                        <span className="repogroup-page__keyword-text">repo:</span>
                                        ^github\.com/golang/go${' '}
                                        <span className="repogroup-page__keyword-text">type:</span>diff{' '}
                                        <span className="repogroup-page__keyword-text">after:</span>"1 week ago"
                                    </Link>{' '}
                                    <div>
                                        Browse diffs for recent code changes in the 'golang/go' GitHub repository.
                                    </div>
                                </li>
                                <li className="list-group-item px-0 py-3">
                                    <Link
                                        to='/search?q=file:pod.yaml+content:"kind:+ReplicationController"&patternType=literal'
                                        className="text-monospace mb-1"
                                    >
                                        <span className="repogroup-page__keyword-text">file:</span>pod.yaml{' '}
                                        <span className="repogroup-page__keyword-text">content:</span>"kind:
                                        ReplicationController"
                                    </Link>{' '}
                                    <div>
                                        Use a ReplicationController configuration to ensure specified number of pod
                                        replicas are running at any one time.
                                    </div>
                                </li>
                            </ul>
                        </div>
                        <div>
                            <div className="d-flex align-items-baseline">
                                <h3 className="search-page__help-content-header mr-2">Search a language</h3>
                                <span className="text-monospace font-weight-normal search-page__lang-ref">
                                    <span className="repogroup-page__keyword-text">lang:</span>
                                    <i className="repogroup-page__keyword-value-text">name</i>
                                </span>
                            </div>
                            <div className="search-page__lang-list">
                                {homepageLanguageList.map(language => (
                                    <Link
                                        className="text-monospace search-page__web-link"
                                        to={`/search?q=lang:${language.filterName}`}
                                        key={language.name}
                                    >
                                        {language.name}
                                    </Link>
                                ))}
                            </div>
                        </div>
                        <div>
                            <h3 className="search-page__help-content-header">Search syntax</h3>
                            <div className="search-page__lang-list">
                                <dl>
                                    <dt className="search-page__help-content-subheading">Common search keywords</dt>
                                    <dd className="text-monospace">repo:my/repo</dd>
                                    <dd className="text-monospace">repo:github.com/myorg/</dd>
                                    <dd className="text-monospace">file:my/file</dd>
                                    <dd className="text-monospace">lang:javascript</dd>
                                </dl>
                                <dl>
                                    <dt className="search-page__help-content-subheading">
                                        Diff/commit search keywords:
                                    </dt>
                                    <dd className="text-monospace">type:diff or type:commit</dd>
                                    <dd className="text-monospace">after:”2 weeks ago”</dd>
                                    <dd className="text-monospace">author:alice@example.com</dd>{' '}
                                    <dd className="text-monospace">repo:r@*refs/heads/ (all branches)</dd>
                                </dl>
                                <dl>
                                    <dt className="search-page__help-content-subheading">Finding matches</dt>
                                    <dd className="text-monospace">Regexp: (read|write)File</dd>{' '}
                                    <dd className="text-monospace">Exact: “fs.open(f)”</dd>
                                </dl>
                                <dl>
                                    <dt className="search-page__help-content-subheading">Structural Searches</dt>
                                    <dd className="text-monospace">:[arg] matches arguments</dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
