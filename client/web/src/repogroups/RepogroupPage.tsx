import * as H from 'history'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import SourceRepositoryMultipleIcon from 'mdi-react/SourceRepositoryMultipleIcon'
import React, { useEffect, useMemo } from 'react'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps, Settings, isSettingsValid } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../auth'
import { PageTitle } from '../components/PageTitle'
import { SyntaxHighlightedSearchQuery } from '../components/SyntaxHighlightedSearchQuery'
import { SearchPatternType } from '../graphql-operations'
import { KeyboardShortcutsProps } from '../keyboardShortcuts/keyboardShortcuts'
import { VersionContext } from '../schema/site.schema'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    OnboardingTourProps,
    ShowQueryBuilderProps,
    ParsedSearchQueryProps,
    SearchContextInputProps,
} from '../search'
import { submitSearch } from '../search/helpers'
import { SearchPageInput } from '../search/input/SearchPageInput'
import { ThemePreferenceProps } from '../theme'
import { eventLogger } from '../tracking/eventLogger'

import { RepogroupMetadata } from './types'

export interface RepogroupPageProps
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        TelemetryProps,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        PatternTypeProps,
        CaseSensitivityProps,
        KeyboardShortcutsProps,
        ExtensionsControllerProps<'executeCommand'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL'>,
        VersionContextProps,
        SearchContextInputProps,
        OnboardingTourProps,
        ShowQueryBuilderProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => Promise<void>
    availableVersionContexts: VersionContext[] | undefined

    // Repogroup page metadata
    repogroupMetadata: RepogroupMetadata

    /** Whether globbing is enabled for filters. */
    globbing: boolean

    // Whether to additionally highlight or provide hovers for tokens, e.g., regexp character sets.
    enableSmartQuery: boolean
}

export const RepogroupPage: React.FunctionComponent<RepogroupPageProps> = (props: RepogroupPageProps) => {
    useEffect(() => props.telemetryService.logViewEvent(`Repogroup:${props.repogroupMetadata.name}`), [
        props.repogroupMetadata.name,
        props.telemetryService,
    ])

    const repogroupQuery = `repogroup:${props.repogroupMetadata.name}`

    // Get repogroups from settings.
    const repogroups: { [name: string]: string[] } | undefined = useMemo(
        () =>
            isSettingsValid<Settings>(props.settingsCascade) && props.settingsCascade.final['search.repositoryGroups'],
        [props.settingsCascade]
    )

    // Find the repositories for this specific repogroup.
    const repogroupRepoList = repogroups?.[props.repogroupMetadata.name]

    const onSubmitExample = (query: string, patternType: SearchPatternType) => (
        event?: React.MouseEvent<HTMLButtonElement>
    ): void => {
        eventLogger.log('RepositoryGroupSuggestionClicked')
        // eslint-disable-next-line no-unused-expressions
        event?.preventDefault()
        submitSearch({ ...props, query, patternType, source: 'repogroupPage' })
    }

    return (
        <div className="repogroup-page">
            <PageTitle title={props.repogroupMetadata.title} />
            <RepogroupPageLogo
                className="repogroup-page__logo"
                icon={props.repogroupMetadata.homepageIcon}
                text={props.repogroupMetadata.title}
            />
            <div className="repogroup-page__subheading">
                {props.repogroupMetadata.lowProfile ? (
                    <>{props.repogroupMetadata.description}</>
                ) : (
                    <span className="text-monospace">
                        <span className="search-filter-keyword">repogroup:</span>
                        {props.repogroupMetadata.name}
                    </span>
                )}
            </div>
            <div className="repogroup-page__container">
                {props.repogroupMetadata.lowProfile ? (
                    <SearchPageInput
                        {...props}
                        hiddenQueryPrefix={repogroupQuery}
                        source="repogroupPage"
                        hideVersionContexts={true}
                        showQueryBuilder={false}
                    />
                ) : (
                    <SearchPageInput {...props} queryPrefix={repogroupQuery} source="repogroupPage" />
                )}
            </div>
            {!props.repogroupMetadata.lowProfile && (
                <div className="row">
                    <div className="repogroup-page__column col-xs-12 col-lg-7">
                        <p className="repogroup-page__content-description h5 font-weight-normal mb-4">
                            {props.repogroupMetadata.description}
                        </p>

                        <h2>Search examples</h2>
                        {props.repogroupMetadata.examples.map(example => (
                            <div className="mt-3" key={example.title}>
                                <h3 className="mb-3">{example.title}</h3>
                                <p>{example.description}</p>
                                <div className="d-flex mb-4">
                                    <small className="repogroup-page__example-bar form-control text-monospace ">
                                        <SyntaxHighlightedSearchQuery query={`${repogroupQuery} ${example.query}`} />
                                    </small>
                                    <div className="d-flex">
                                        <button
                                            className="btn btn-secondary btn-sm repogroup-page__search-button"
                                            type="button"
                                            aria-label="Search"
                                            onClick={onSubmitExample(
                                                `${repogroupQuery} ${example.query}`,
                                                example.patternType
                                            )}
                                        >
                                            Search
                                        </button>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                    <div className="repogroup-page__column col-xs-12 col-lg-5">
                        <div className="order-2-lg order-1-xs">
                            <div className="repogroup-page__repo-card card">
                                <h2>
                                    <SourceRepositoryMultipleIcon className="icon-inline mr-2" />
                                    Repositories
                                </h2>
                                <p>
                                    Using the syntax{' '}
                                    <code>
                                        <span className="search-filter-keyword ">repogroup:</span>
                                        {props.repogroupMetadata.name}
                                    </code>{' '}
                                    in a query will search these repositories:
                                </p>
                                <div className="repogroup-page__repo-list row">
                                    <div className="col-lg-6">
                                        {repogroupRepoList
                                            ?.slice(0, Math.ceil(repogroupRepoList.length / 2))
                                            .map(repo => (
                                                <RepoLink key={repo} repo={repo} />
                                            ))}
                                    </div>
                                    <div className="col-lg-6">
                                        {repogroupRepoList
                                            ?.slice(Math.ceil(repogroupRepoList.length / 2), repogroupRepoList.length)
                                            .map(repo => (
                                                <RepoLink key={repo} repo={repo} />
                                            ))}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}

const RepoLinkClicked = (repoName: string) => (): void =>
    eventLogger.log('RepogroupPageRepoLinkClicked', { repo_name: repoName })

const RepoLink: React.FunctionComponent<{ repo: string }> = ({ repo }) => (
    <li className="repogroup-page__repo-item list-unstyled mb-3" key={repo}>
        {repo.startsWith('github.com') && (
            <>
                <a href={`https://${repo}`} target="_blank" rel="noopener noreferrer" onClick={RepoLinkClicked(repo)}>
                    <GithubIcon className="icon-inline repogroup-page__repo-list-icon" />
                </a>
                <Link to={`/${repo}`} className="text-monospace search-filter-keyword">
                    {displayRepoName(repo)}
                </Link>
            </>
        )}
        {repo.startsWith('gitlab.com') && (
            <>
                <a href={`https://${repo}`} target="_blank" rel="noopener noreferrer" onClick={RepoLinkClicked(repo)}>
                    <GitlabIcon className="icon-inline repogroup-page__repo-list-icon" />
                </a>
                <Link to={`/${repo}`} className="text-monospace search-filter-keyword">
                    {displayRepoName(repo)}
                </Link>
            </>
        )}
        {repo.startsWith('bitbucket.com') && (
            <>
                <a href={`https://${repo}`} target="_blank" rel="noopener noreferrer" onClick={RepoLinkClicked(repo)}>
                    <BitbucketIcon className="icon-inline repogroup-page__repo-list-icon" />
                </a>
                <Link to={`/${repo}`} className="text-monospace search-filter-keyword">
                    {displayRepoName(repo)}
                </Link>
            </>
        )}
    </li>
)

interface RepogroupPageLogoProps extends Exclude<React.ImgHTMLAttributes<HTMLImageElement>, 'src'> {
    icon: string
    text: string
}

/**
 * The repogroup logo image.
 */
const RepogroupPageLogo: React.FunctionComponent<RepogroupPageLogoProps> = props => (
    <div className="repogroup-page__logo-container d-flex align-items-center">
        <img {...props} src={props.icon} alt="" />
        <span className="h3 font-weight-normal mb-0 ml-1">{props.text}</span>
    </div>
)
