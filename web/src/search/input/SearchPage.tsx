import * as H from 'history'
import React, { useEffect, useState, useCallback, useMemo } from 'react'
import {
    parseSearchURLQuery,
    PatternTypeProps,
    InteractiveSearchProps,
    CaseSensitivityProps,
    SmartSearchFieldProps,
    CopyQueryButtonProps,
} from '..'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { Notices } from '../../global/Notices'
import { Settings } from '../../schema/settings.schema'
import { ThemeProps } from '../../../../shared/src/theme'
import { eventLogger, EventLoggerProps } from '../../tracking/eventLogger'
import { ThemePreferenceProps } from '../../theme'
import { limitString } from '../../util'
import { submitSearch } from '../helpers'
import { QuickLinks } from '../QuickLinks'
import { QueryInput } from './QueryInput'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { SearchButton } from './SearchButton'
import { SearchScopes } from './SearchScopes'
import { InteractiveModeInput } from './interactive/InteractiveModeInput'
import { KeyboardShortcutsProps, KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '../../keyboardShortcuts/keyboardShortcuts'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SearchModeToggle } from './interactive/SearchModeToggle'
import { Link } from '../../../../shared/src/components/Link'
import { BrandLogo } from '../../components/branding/BrandLogo'
import { VersionContextDropdown } from '../../nav/VersionContextDropdown'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { VersionContext } from '../../schema/site.schema'
import { ViewGrid } from '../../repo/tree/ViewGrid'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { getViewsForContainer } from '../../../../shared/src/api/client/services/viewService'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { ContributableViewContainer } from '../../../../shared/src/api/protocol'
import { EMPTY } from 'rxjs'
import classNames from 'classnames'
import { repogroupList } from '../../repogroups/RepogroupList'

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
        VersionContextProps {
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

const languages: { name: string; filterName: string }[] = [
    { name: 'C', filterName: 'c' },
    { name: 'C++', filterName: 'cpp' },
    { name: 'C#', filterName: 'csharp' },
    { name: 'CSS', filterName: 'css' },
    { name: 'Go', filterName: 'go' },
    { name: 'Graphql', filterName: 'graphql' },
    { name: 'Haskell', filterName: 'haskell' },
    { name: 'Html', filterName: 'html' },
    { name: 'Java', filterName: 'java' },
    { name: 'Javascript', filterName: 'javascript' },
    { name: 'Json', filterName: 'json' },
    { name: 'Lua', filterName: 'lua' },
    { name: 'Markdown', filterName: 'markdown' },
    { name: 'Php', filterName: 'php' },
    { name: 'Powershell', filterName: 'powershell' },
    { name: 'Python', filterName: 'python' },
    { name: 'R', filterName: 'r' },
    { name: 'Ruby', filterName: 'ruby' },
    { name: 'Sass', filterName: 'sass' },
    { name: 'Swift', filterName: 'swift' },
    { name: 'Typescript', filterName: 't  ypescript' },
]

/**
 * The search page
 */
export const SearchPage: React.FunctionComponent<Props> = props => {
    useEffect(() => eventLogger.logViewEvent('Home'))

    const queryFromUrl = parseSearchURLQuery(props.location.search) || ''

    /** The query cursor position and value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: queryFromUrl,
        cursorPosition: queryFromUrl.length,
    })

    const quickLinks =
        (isSettingsValid<Settings>(props.settingsCascade) && props.settingsCascade.final.quicklinks) || []

    const onSubmit = useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            // False positive
            // eslint-disable-next-line no-unused-expressions
            event?.preventDefault()
            submitSearch({ ...props, query: userQueryState.query, source: 'home' })
        },
        [props, userQueryState.query]
    )

    const pageTitle = queryFromUrl ? `${limitString(userQueryState.query, 25, true)}` : undefined

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
            <PageTitle title={pageTitle} />
            <BrandLogo className="search-page__logo" isLightTheme={props.isLightTheme} />
            <div
                className={classNames('search-page__container', {
                    'search-page__container--with-repogroups': props.isSourcegraphDotCom,
                })}
            >
                <div className="d-flex flex-row flex-shrink-past-contents">
                    {props.splitSearchModes && props.interactiveSearchMode ? (
                        <InteractiveModeInput
                            {...props}
                            navbarSearchState={userQueryState}
                            onNavbarQueryChange={setUserQueryState}
                            toggleSearchMode={props.toggleSearchMode}
                            lowProfile={false}
                        />
                    ) : (
                        <>
                            <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                                <div className="search-page__input-container">
                                    {props.splitSearchModes && (
                                        <SearchModeToggle
                                            {...props}
                                            interactiveSearchMode={props.interactiveSearchMode}
                                        />
                                    )}
                                    <VersionContextDropdown
                                        history={props.history}
                                        caseSensitive={props.caseSensitive}
                                        patternType={props.patternType}
                                        navbarSearchQuery={userQueryState.query}
                                        versionContext={props.versionContext}
                                        setVersionContext={props.setVersionContext}
                                        availableVersionContexts={props.availableVersionContexts}
                                    />
                                    {props.smartSearchField ? (
                                        <LazyMonacoQueryInput
                                            {...props}
                                            hasGlobalQueryBehavior={true}
                                            queryState={userQueryState}
                                            onChange={setUserQueryState}
                                            onSubmit={onSubmit}
                                            autoFocus={true}
                                        />
                                    ) : (
                                        <QueryInput
                                            {...props}
                                            value={userQueryState}
                                            onChange={setUserQueryState}
                                            autoFocus="cursor-at-end"
                                            hasGlobalQueryBehavior={true}
                                            patternType={props.patternType}
                                            setPatternType={props.setPatternType}
                                            withSearchModeToggle={props.splitSearchModes}
                                            keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR}
                                        />
                                    )}
                                    <SearchButton />
                                </div>
                                <div className="search-page__input-sub-container">
                                    {!props.splitSearchModes && (
                                        <Link className="btn btn-link btn-sm pl-0" to="/search/query-builder">
                                            Query builder
                                        </Link>
                                    )}
                                    <SearchScopes
                                        history={props.history}
                                        query={userQueryState.query}
                                        authenticatedUser={props.authenticatedUser}
                                        settingsCascade={props.settingsCascade}
                                        patternType={props.patternType}
                                        versionContext={props.versionContext}
                                    />
                                </div>
                                <QuickLinks quickLinks={quickLinks} className="search-page__input-sub-container" />
                                <Notices
                                    className="my-3"
                                    location="home"
                                    settingsCascade={props.settingsCascade}
                                    history={props.history}
                                />
                            </Form>
                        </>
                    )}
                </div>
                {views && <ViewGrid {...props} className="mt-5" views={views} />}
            </div>
            {props.isSourcegraphDotCom && (
                <div className="search-page__repogroup-content mt-5">
                    <div className="d-flex align-items-baseline mb-3">
                        <h3 className="search-page__help-content-header mr-1">Search in repository groups</h3>
                        <span className="text-monospace font-weight-normal">
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
                                <li className="list-group-item px-0 py-3">
                                    <Link
                                        to="/search?q=lang:javascript+alert%28:%5Bvariable%5D%29&patternType=literal"
                                        className="text-monospace mb-1"
                                    >
                                        <span className="repogroup-page__keyword-text">lang:</span>javascript
                                        alert(:[variable])
                                    </Link>{' '}
                                    <div>
                                        A collection of top react repositories, including, tooling, ui, testing and key
                                        components.
                                    </div>
                                </li>
                                <li className="list-group-item px-0 py-3">
                                    <Link
                                        to="/search?q=lang:javascript+alert%28:%5Bvariable%5D%29&patternType=literal"
                                        className="text-monospace mb-1"
                                    >
                                        <span className="repogroup-page__keyword-text">lang:</span>javascript
                                        alert(:[variable])
                                    </Link>{' '}
                                    <div>
                                        A collection of top react repositories, including, tooling, ui, testing and key
                                        components.
                                    </div>
                                </li>
                                <li className="list-group-item px-0 py-3">
                                    <Link
                                        to="/search?q=lang:javascript+alert%28:%5Bvariable%5D%29&patternType=literal"
                                        className="text-monospace mb-1"
                                    >
                                        <span className="repogroup-page__keyword-text">lang:</span>javascript
                                        alert(:[variable])
                                    </Link>{' '}
                                    <div>
                                        A collection of top react repositories, including, tooling, ui, testing and key
                                        components.
                                    </div>
                                </li>
                            </ul>
                        </div>
                        <div>
                            <div className="d-flex align-items-baseline">
                                <h3 className="search-page__help-content-header mr-1">Search a language</h3>
                                <span className="text-monospace font-weight-normal">
                                    <span className="repogroup-page__keyword-text">lang:</span>
                                    <i className="repogroup-page__keyword-value-text">name</i>
                                </span>
                            </div>
                            <div className="search-page__lang-list">
                                {languages.map(language => (
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
