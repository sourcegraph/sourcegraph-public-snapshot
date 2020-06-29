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
import { SettingsCascadeProps, Settings } from '../../../shared/src/settings/settings'

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
import { RepogroupMetadata, RepositoryType, CodeHosts } from './types'

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

    // Repogroup page metadata
    repogroupMetadata: RepogroupMetadata
}

export const RepogroupPage: React.FunctionComponent<Props> = (props: Props) => {
    useEffect(() => eventLogger.logViewEvent('Python2To3RepoGroup'))

    const repogroupQuery = `repogroup:${props.repogroupMetadata.name}`

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
            submitSearch({ ...props, query: `${repogroupQuery} ${userQueryState.query}`, source: 'home' })
        },
        [props, repogroupQuery, userQueryState.query]
    )

    const onSubmitExample = (query: string) => (event?: React.MouseEvent<HTMLButtonElement>): void => {
        // eslint-disable-next-line no-unused-expressions
        event?.preventDefault()
        // TODO: update source
        submitSearch({ ...props, query, source: 'home' })
    }

    return (
        <div className="repogroup-page">
            <PageTitle title="Python 2 to 3 migration" />
            <BrandLogo
                className="repogroup-page__logo"
                isLightTheme={props.isLightTheme}
                customLogoUrl={
                    props.repogroupMetadata.customLogoUrl ? props.repogroupMetadata.customLogoUrl : undefined
                }
            />
            <div className="repogroup-page__subheading">
                <span className="text-monospace">
                    <span className="repogroup-page__keyword-text">repogroup:</span>
                    {props.repogroupMetadata.name}
                </span>
            </div>
            <div className="repogroup-page__container">
                <div className="d-flex flex-row flex-shrink-past-contents">
                    <>
                        <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                            <div className="repogroup-page__input-container">
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
                    <p className="h5 font-weight-normal mb-4">{props.repogroupMetadata.description}</p>
                    {props.repogroupMetadata.examples.map(example => (
                        <>
                            <h3 className="mb-3">{example.title}</h3>
                            <div className="d-flex mb-4">
                                <div className="repogroup-page__example-bar form-control text-monospace">
                                    <span className="repogroup-page__keyword-text">repogroup:</span>
                                    {props.repogroupMetadata.name} {example.exampleQuery}
                                </div>
                                <div className="search-button d-flex">
                                    <button
                                        className="btn btn-primary search-button__btn e2e-search-button btn-secondary"
                                        type="button"
                                        aria-label="Search"
                                        onClick={onSubmitExample(`${repogroupQuery} ${example.exampleQuery}`)}
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
                        <p>
                            Using the syntax{' '}
                            <span className="text-monospace">
                                <span className="repogroup-page__keyword-text">repogroup:</span>
                                {props.repogroupMetadata.name}
                            </span>{' '}
                            in a query will search these repositories:
                        </p>
                        <div className="repogroup-page__repo-list row">
                            <div className="col-lg-6">
                                {props.repogroupMetadata.repositories
                                    .slice(0, Math.ceil(props.repogroupMetadata.repositories.length / 2))
                                    .map(repo => (
                                        <RepoLink key={repo.name} repo={repo} />
                                    ))}
                            </div>
                            <div className="col-lg-6">
                                {props.repogroupMetadata.repositories
                                    .slice(
                                        Math.ceil(props.repogroupMetadata.repositories.length / 2),
                                        props.repogroupMetadata.repositories.length
                                    )
                                    .map(repo => (
                                        <RepoLink key={repo.name} repo={repo} />
                                    ))}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

const RepoLink: React.FunctionComponent<{ repo: RepositoryType }> = props => (
    <li className="repogroup-page__repo-item list-unstyled mb-3" key={props.repo.name}>
        {props.repo.codehost === CodeHosts.GITHUB && (
            <GithubIcon className="icon-inline repogroup-page__repo-list-icon" />
        )}
        {props.repo.codehost === CodeHosts.GITLAB && (
            <GitlabIcon className="icon-inline repogroup-page__repo-list-icon" />
        )}
        {props.repo.codehost === CodeHosts.BITBUCKET && (
            <BitbucketIcon className="icon-inline repogroup-page__repo-list-icon" />
        )}
        <Link to={`/${props.repo.name}`}>{props.repo.name}</Link>
    </li>
)
