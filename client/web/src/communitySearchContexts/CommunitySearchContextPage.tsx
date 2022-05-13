import React, { useEffect, useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import SourceRepositoryMultipleIcon from 'mdi-react/SourceRepositoryMultipleIcon'
import { catchError, startWith } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { SearchContextInputProps, SearchContextProps } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { KeyboardShortcutsProps } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps, Settings } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, useObservable, Link, Card, Icon, Typography } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { PageTitle } from '../components/PageTitle'
import { SearchPatternType } from '../graphql-operations'
import { submitSearch } from '../search/helpers'
import { SearchPageInput } from '../search/home/SearchPageInput'
import { useNavbarQueryState } from '../stores'
import { ThemePreferenceProps } from '../theme'
import { eventLogger } from '../tracking/eventLogger'

import { CommunitySearchContextMetadata } from './types'

import styles from './CommunitySearchContextPage.module.scss'

export interface CommunitySearchContextPageProps
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        TelemetryProps,
        KeyboardShortcutsProps,
        ExtensionsControllerProps<'executeCommand'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL' | 'requestGraphQL'>,
        SearchContextInputProps,
        Pick<SearchContextProps, 'fetchSearchContextBySpec'> {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean

    // CommunitySearchContext page metadata
    communitySearchContextMetadata: CommunitySearchContextMetadata

    /** Whether globbing is enabled for filters. */
    globbing: boolean
}

export const CommunitySearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<CommunitySearchContextPageProps>
> = (props: CommunitySearchContextPageProps) => {
    const LOADING = 'loading' as const

    useEffect(
        () =>
            props.telemetryService.logViewEvent(`CommunitySearchContext:${props.communitySearchContextMetadata.spec}`),
        [props.communitySearchContextMetadata.spec, props.telemetryService]
    )
    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)

    const contextQuery = `context:${props.communitySearchContextMetadata.spec}`

    const { fetchSearchContextBySpec } = props
    const searchContextOrError = useObservable(
        useMemo(
            () =>
                fetchSearchContextBySpec(props.communitySearchContextMetadata.spec, props.platformContext).pipe(
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [props.communitySearchContextMetadata.spec, fetchSearchContextBySpec, props.platformContext]
        )
    )

    const onSubmitExample = (query: string, patternType: SearchPatternType) => (
        event?: React.MouseEvent<HTMLButtonElement>
    ): void => {
        eventLogger.log('CommunitySearchContextSuggestionClicked')
        event?.preventDefault()
        submitSearch({ ...props, query, caseSensitive, patternType, source: 'communitySearchContextPage' })
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
                        {/*
                           a11y-ignore
                           Rule: "color-contrast" (Elements must have sufficient color contrast)
                           GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                          */}
                        <span className="search-filter-keyword a11y-ignore">context:</span>
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

                        <Typography.H2>Search examples</Typography.H2>
                        {props.communitySearchContextMetadata.examples.map(example => (
                            <div className="mt-3" key={example.title}>
                                <Typography.H3 className="mb-3">{example.title}</Typography.H3>
                                <p>{example.description}</p>
                                <div className="d-flex mb-4">
                                    <small className={classNames('form-control text-monospace ', styles.exampleBar)}>
                                        <SyntaxHighlightedSearchQuery query={`${contextQuery} ${example.query}`} />
                                    </small>
                                    <div className="d-flex">
                                        <Button
                                            className={styles.searchButton}
                                            aria-label="Search"
                                            onClick={onSubmitExample(
                                                `${contextQuery} ${example.query}`,
                                                example.patternType
                                            )}
                                            variant="secondary"
                                            size="sm"
                                        >
                                            Search
                                        </Button>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                    <div className={classNames('col-xs-12 col-lg-5', styles.column)}>
                        <div className="order-2-lg order-1-xs">
                            <Card className={styles.repoCard}>
                                <Typography.H2>
                                    <Icon className="mr-2" as={SourceRepositoryMultipleIcon} />
                                    Repositories
                                </Typography.H2>
                                <p>
                                    Using the syntax{' '}
                                    <code>
                                        {/*
                                            a11y-ignore
                                            Rule: "color-contrast" (Elements must have sufficient color contrast)
                                            GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                                          */}
                                        <span className="search-filter-keyword a11y-ignore">context:</span>
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
                            </Card>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}

const RepoLinkClicked = (repoName: string) => (): void =>
    eventLogger.log('CommunitySearchContextPageRepoLinkClicked', { repo_name: repoName }, { repo_name: repoName })

const RepoLink: React.FunctionComponent<React.PropsWithChildren<{ repo: string }>> = ({ repo }) => (
    <li className={classNames('list-unstyled mb-3', styles.repoItem)} key={repo}>
        {repo.startsWith('github.com') && (
            <>
                <Link to={`https://${repo}`} target="_blank" rel="noopener noreferrer" onClick={RepoLinkClicked(repo)}>
                    <Icon className={styles.repoListIcon} as={GithubIcon} />
                </Link>
                <Link to={`/${repo}`} className="text-monospace search-filter-keyword">
                    {displayRepoName(repo)}
                </Link>
            </>
        )}
        {repo.startsWith('gitlab.com') && (
            <>
                <Link to={`https://${repo}`} target="_blank" rel="noopener noreferrer" onClick={RepoLinkClicked(repo)}>
                    <Icon className={styles.repoListIcon} as={GitlabIcon} />
                </Link>
                <Link to={`/${repo}`} className="text-monospace search-filter-keyword">
                    {displayRepoName(repo)}
                </Link>
            </>
        )}
        {repo.startsWith('bitbucket.com') && (
            <>
                <Link to={`https://${repo}`} target="_blank" rel="noopener noreferrer" onClick={RepoLinkClicked(repo)}>
                    <Icon className={styles.repoListIcon} as={BitbucketIcon} />
                </Link>
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
const CommunitySearchContextPageLogo: React.FunctionComponent<
    React.PropsWithChildren<CommunitySearchContextPageLogoProps>
> = props => (
    <div className={classNames('d-flex align-items-center', styles.logoContainer)}>
        <img {...props} src={props.icon} alt="" />
        <Typography.H3 as="span" className="font-weight-normal mb-0 ml-1">
            {props.text}
        </Typography.H3>
    </div>
)
