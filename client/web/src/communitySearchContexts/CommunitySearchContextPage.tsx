import classNames from 'classnames'
import * as H from 'history'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import SourceRepositoryMultipleIcon from 'mdi-react/SourceRepositoryMultipleIcon'
import React, { useEffect, useMemo } from 'react'
import { catchError, startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps, Settings } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/web/src/components/SyntaxHighlightedSearchQuery'

import { AuthenticatedUser } from '../auth'
import { SearchPatternType } from '../graphql-operations'
import { KeyboardShortcutsProps } from '../keyboardShortcuts/keyboardShortcuts'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    OnboardingTourProps,
    ParsedSearchQueryProps,
    SearchContextInputProps,
    SearchContextProps,
} from '../search'
import { submitSearch } from '../search/helpers'
import { SearchPageInput } from '../search/home/SearchPageInput'
import { ThemePreferenceProps } from '../theme'
import { eventLogger } from '../tracking/eventLogger'

import styles from './CommunitySearchContextPage.module.scss'
import { CommunitySearchContextMetadata } from './types'

export interface CommunitySearchContextPageProps
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
        SearchContextInputProps,
        Pick<SearchContextProps, 'fetchSearchContextBySpec'>,
        OnboardingTourProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean

    // CommunitySearchContext page metadata
    communitySearchContextMetadata: CommunitySearchContextMetadata

    /** Whether globbing is enabled for filters. */
    globbing: boolean
}

export const CommunitySearchContextPage: React.FunctionComponent<CommunitySearchContextPageProps> = (
    props: CommunitySearchContextPageProps
) => {
    const LOADING = 'loading' as const

    useEffect(
        () =>
            props.telemetryService.logViewEvent(`CommunitySearchContext:${props.communitySearchContextMetadata.spec}`),
        [props.communitySearchContextMetadata.spec, props.telemetryService]
    )

    const contextQuery = `context:${props.communitySearchContextMetadata.spec}`

    const { fetchSearchContextBySpec } = props
    const searchContextOrError = useObservable(
        useMemo(
            () =>
                fetchSearchContextBySpec(props.communitySearchContextMetadata.spec).pipe(
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [props.communitySearchContextMetadata.spec, fetchSearchContextBySpec]
        )
    )

    const onSubmitExample = (query: string, patternType: SearchPatternType) => (
        event?: React.MouseEvent<HTMLButtonElement>
    ): void => {
        eventLogger.log('CommunitySearchContextSuggestionClicked')
        event?.preventDefault()
        submitSearch({ ...props, query, patternType, source: 'communitySearchContextPage' })
    }

    return (
        <div className={styles.communitySearchContextsPage}>
            <PageTitle title={props.communitySearchContextMetadata.title} />
            <CommunitySearchContextPageLogo
                className={styles.logo}
                icon={props.communitySearchContextMetadata.homepageIcon}
                text={props.communitySearchContextMetadata.title}
            />
            <div className={styles.subheading}>
                {props.communitySearchContextMetadata.lowProfile ? (
                    <>{props.communitySearchContextMetadata.description}</>
                ) : (
                    <span className="text-monospace">
                        <span className="search-filter-keyword">context:</span>
                        {props.communitySearchContextMetadata.spec}
                    </span>
                )}
            </div>
            <div className={styles.container}>
                {props.communitySearchContextMetadata.lowProfile ? (
                    <SearchPageInput
                        {...props}
                        selectedSearchContextSpec={props.communitySearchContextMetadata.spec}
                        source="communitySearchContextPage"
                    />
                ) : (
                    <SearchPageInput
                        {...props}
                        selectedSearchContextSpec={props.communitySearchContextMetadata.spec}
                        source="communitySearchContextPage"
                    />
                )}
            </div>
            {!props.communitySearchContextMetadata.lowProfile && (
                <div className="row">
                    <div className={classNames('col-xs-12 col-lg-7', styles.column)}>
                        <p className={classNames('h5 font-weight-normal mb-4', styles.contentDescription)}>
                            {props.communitySearchContextMetadata.description}
                        </p>

                        <h2>Search examples</h2>
                        {props.communitySearchContextMetadata.examples.map(example => (
                            <div className="mt-3" key={example.title}>
                                <h3 className="mb-3">{example.title}</h3>
                                <p>{example.description}</p>
                                <div className="d-flex mb-4">
                                    <small className={classNames('form-control text-monospace ', styles.exampleBar)}>
                                        <SyntaxHighlightedSearchQuery query={`${contextQuery} ${example.query}`} />
                                    </small>
                                    <div className="d-flex">
                                        <button
                                            className={classNames('btn btn-secondary btn-sm', styles.searchButton)}
                                            type="button"
                                            aria-label="Search"
                                            onClick={onSubmitExample(
                                                `${contextQuery} ${example.query}`,
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
                    <div className={classNames('col-xs-12 col-lg-5', styles.column)}>
                        <div className="order-2-lg order-1-xs">
                            <div className={classNames('card', styles.repoCard)}>
                                <h2>
                                    <SourceRepositoryMultipleIcon className="icon-inline mr-2" />
                                    Repositories
                                </h2>
                                <p>
                                    Using the syntax{' '}
                                    <code>
                                        <span className="search-filter-keyword">context:</span>
                                        {props.communitySearchContextMetadata.spec}
                                    </code>{' '}
                                    in a query will search these repositories:
                                </p>
                                {searchContextOrError &&
                                    !isErrorLike(searchContextOrError) &&
                                    searchContextOrError !== LOADING && (
                                        <div className="community-search-contexts-page__repo-list row">
                                            <div className="col-lg-6">
                                                {searchContextOrError.repositories
                                                    .slice(0, Math.ceil(searchContextOrError.repositories.length / 2))
                                                    .map(repo => (
                                                        <RepoLink
                                                            key={repo.repository.name}
                                                            repo={repo.repository.name}
                                                        />
                                                    ))}
                                            </div>
                                            <div className="col-lg-6">
                                                {searchContextOrError.repositories
                                                    .slice(Math.ceil(searchContextOrError.repositories.length / 2))
                                                    .map(repo => (
                                                        <RepoLink
                                                            key={repo.repository.name}
                                                            repo={repo.repository.name}
                                                        />
                                                    ))}
                                            </div>
                                        </div>
                                    )}
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}

const RepoLinkClicked = (repoName: string) => (): void =>
    eventLogger.log('CommunitySearchContextPageRepoLinkClicked', { repo_name: repoName }, { repo_name: repoName })

const RepoLink: React.FunctionComponent<{ repo: string }> = ({ repo }) => (
    <li className={classNames('list-unstyled mb-3', styles.repoItem)} key={repo}>
        {repo.startsWith('github.com') && (
            <>
                <a href={`https://${repo}`} target="_blank" rel="noopener noreferrer" onClick={RepoLinkClicked(repo)}>
                    <GithubIcon className={classNames('icon-inline', styles.repoListIcon)} />
                </a>
                <Link to={`/${repo}`} className="text-monospace search-filter-keyword">
                    {displayRepoName(repo)}
                </Link>
            </>
        )}
        {repo.startsWith('gitlab.com') && (
            <>
                <a href={`https://${repo}`} target="_blank" rel="noopener noreferrer" onClick={RepoLinkClicked(repo)}>
                    <GitlabIcon className={classNames('icon-inline', styles.repoListIcon)} />
                </a>
                <Link to={`/${repo}`} className="text-monospace search-filter-keyword">
                    {displayRepoName(repo)}
                </Link>
            </>
        )}
        {repo.startsWith('bitbucket.com') && (
            <>
                <a href={`https://${repo}`} target="_blank" rel="noopener noreferrer" onClick={RepoLinkClicked(repo)}>
                    <BitbucketIcon className={classNames('icon-inline', styles.repoListIcon)} />
                </a>
                <Link to={`/${repo}`} className="text-monospace search-filter-keyword">
                    {displayRepoName(repo)}
                </Link>
            </>
        )}
    </li>
)

interface CommunitySearchContextPageLogoProps extends Exclude<React.ImgHTMLAttributes<HTMLImageElement>, 'src'> {
    icon: string
    text: string
}

/**
 * The community search context logo image.
 */
const CommunitySearchContextPageLogo: React.FunctionComponent<CommunitySearchContextPageLogoProps> = props => (
    <div className={classNames('d-flex align-items-center', styles.logoContainer)}>
        <img {...props} src={props.icon} alt="" />
        <span className="h3 font-weight-normal mb-0 ml-1">{props.text}</span>
    </div>
)
