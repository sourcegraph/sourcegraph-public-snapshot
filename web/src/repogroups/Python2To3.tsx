import React, { useEffect, useState } from 'react'
import * as H from 'history'
import { PageTitle } from '../components/PageTitle'
import { BrandLogo } from '../components/branding/BrandLogo'
import { Form } from 'reactstrap'
import { SearchModeToggle } from '../search/input/interactive/SearchModeToggle'
import { VersionContextDropdown } from '../nav/VersionContextDropdown'
import { LazyMonacoQueryInput } from '../search/input/LazyMonacoQueryInput'
import { QueryInput } from '../search/input/QueryInput'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR, KeyboardShortcutsProps } from '../keyboardShortcuts/keyboardShortcuts'
import { SearchButton } from '../search/input/SearchButton'
import { Link } from '../../../shared/src/components/Link'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { Settings } from 'http2'
import { ThemeProps } from '../../../shared/src/theme'
import { ThemePreferenceProps } from '../theme'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    InteractiveSearchProps,
    SmartSearchFieldProps,
    CopyQueryButtonProps,
} from '../search'
import { EventLoggerProps, eventLogger } from '../tracking/eventLogger'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { VersionContextProps } from '../../../shared/src/search/util'
import { VersionContext } from '../schema/site.schema'
import { submitSearch } from '../search/helpers'
import * as GQL from '../../../shared/src/graphql/schema'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import BitbucketIcon from 'mdi-react/BitbucketIcon'

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

enum CodeHosts {
    GITHUB = 'github',
    GITLAB = 'gitlab',
    BITBUCKET = 'bitbucket',
}

interface RepositoryType {
    name: string
    codehost: CodeHosts
}

interface ExampleQuery {
    title: string
    exampleQuery: string
}
export const Python2To3: React.FunctionComponent<Props> = (props: Props) => {
    useEffect(() => eventLogger.logViewEvent('Python2To3RepoGroup'))

    const fixedQuery = 'repogroup:python-2-to-3'

    /** The query cursor position and value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: '',
        cursorPosition: 0,
    })

    const onSubmit = React.useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            // False positive
            // eslint-disable-next-line no-unused-expressions
            event?.preventDefault()
            submitSearch({ ...props, query: `${fixedQuery} ${userQueryState.query}`, source: 'home' })
        },
        [props, userQueryState.query]
    )

    const repositories: RepositoryType[] = [
        { name: 'github.com/sourcegraph/sourcegraph', codehost: CodeHosts.GITHUB },
        { name: 'github.com/test/test', codehost: CodeHosts.GITHUB },
        { name: 'github.com/test/test', codehost: CodeHosts.GITHUB },
        { name: 'github.com/sourcegraph/src-cli', codehost: CodeHosts.GITHUB },
    ]

    const examples: ExampleQuery[] = [
        {
            title: 'Python 2 imports',
            exampleQuery: 'repogroup:refactor-python2-to-3 "from :[package.] import :[function.]”',
        },
        { title: 'Python 3 imports', exampleQuery: 'repogroup:refactor-python2-to-3 from B.w+ import w+' },
        { title: 'Python 2 prints', exampleQuery: 'repogroup:refactor-python2-to-3 \'print ":[string]"\'' },
        { title: 'Python 3 prints', exampleQuery: 'repogroup:refactor-python2-to-3 \'print ":[string]"\'' },
        {
            title: 'Python 2 integer conversion',
            exampleQuery: 'repogroup:refactor-python2-to-3 float(:[arg]) / float(:[arg])',
        },
        {
            title: 'Python 3 integer conversion',
            exampleQuery: 'repogroup:refactor-python2-to-3 lang:python \\sint\\(-*\\d+\\)',
        },
    ]

    const repogroupName = 'repogroup:refactor-python-2-to-3'

    return (
        <div className="repogroup-page">
            <PageTitle title="Python 2 to 3 migration" />
            <BrandLogo className="search-page__logo" isLightTheme={props.isLightTheme} />
            <div className="repogroup-page__subheading">
                <span className="text-monospace">
                    <span className="repogroup-page__repogroup-text">repogroup:</span>
                    {repogroupName}
                </span>
            </div>
            <div className="repogroup-page__container">
                <div className="d-flex flex-row flex-shrink-past-contents">
                    <>
                        <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                            <div className="search-page__input-container">
                                {props.splitSearchModes && (
                                    <SearchModeToggle {...props} interactiveSearchMode={props.interactiveSearchMode} />
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
                            </div>
                        </Form>
                    </>
                </div>
            </div>
            <div className="repogroup-page__content">
                <div className="repogroup-page__column">
                    <p className="mb-4">
                        This repository group contains python 2 and 3 repositories and corresponding search examples.
                        The search examples will help you find code that requires refactoring and review example Python
                        3 syntax.
                    </p>
                    {examples.map(example => (
                        <>
                            <h3 className="mb-3">{example.title}</h3>
                            <div className="d-flex mb-4">
                                <div className="repogroup-page__example-bar form-control">{example.exampleQuery}</div>
                                <div className="search-button d-flex">
                                    <button
                                        className="btn btn-primary search-button__btn e2e-search-button btn-secondary"
                                        type="submit"
                                        aria-label="Search"
                                    >
                                        Search
                                    </button>
                                </div>
                            </div>
                        </>
                    ))}
                </div>
                <div className="repogroup-page__column">
                    <div className="repogroup-page__repo-card card">
                        <h2 className="font-weight-normal">
                            <SourceRepositoryIcon className="icon-inline" />
                            Repositories
                        </h2>
                        <p className="mb-1">
                            Using the syntax{' '}
                            <span className="text-monospace">
                                <span className="repogroup-page__repogroup-text">repogroup:</span>python-2-to-migration
                            </span>{' '}
                            in a query will search these repositories:
                        </p>
                        <div className="repogroup-page__repo-list row">
                            <div className="col-lg-6">
                                {repositories.slice(0, Math.ceil(repositories.length / 2)).map(repo => (
                                    <li className="repogroup-page__repo-item list-unstyled mb-3" key={repo.name}>
                                        {repo.codehost === CodeHosts.GITHUB && (
                                            <GithubIcon className="icon-inline repogroup-page__repo-list-icon" />
                                        )}
                                        {repo.codehost === CodeHosts.GITLAB && (
                                            <GitlabIcon className="icon-inline repogroup-page__repo-list-icon" />
                                        )}
                                        {repo.codehost === CodeHosts.BITBUCKET && (
                                            <BitbucketIcon className="icon-inline repogroup-page__repo-list-icon" />
                                        )}
                                        {repo.name}
                                    </li>
                                ))}
                            </div>
                            <div className="col-lg-6">
                                {repositories
                                    .slice(Math.ceil(repositories.length / 2), repositories.length)
                                    .map(repo => (
                                        <li className="repogroup-page__repo-item list-unstyled mb-3" key={repo.name}>
                                            {repo.codehost === CodeHosts.GITHUB && (
                                                <GithubIcon className="icon-inline repogroup-page__repo-list-icon" />
                                            )}
                                            {repo.codehost === CodeHosts.GITLAB && (
                                                <GitlabIcon className="icon-inline repogroup-page__repo-list-icon" />
                                            )}
                                            {repo.codehost === CodeHosts.BITBUCKET && (
                                                <BitbucketIcon className="icon-inline repogroup-page__repo-list-icon" />
                                            )}
                                            {repo.name}
                                        </li>
                                    ))}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}
