import React, { useState, useCallback, useEffect, useMemo } from 'react'

import {
    mdiChevronUp,
    mdiChevronDown,
    mdiMinusCircleOutline,
    mdiCheck,
    mdiCloseCircle,
    mdiDatabaseOutline,
    mdiDatabaseCheckOutline,
    mdiDatabaseClockOutline,
    mdiDatabaseRefreshOutline,
    mdiDatabaseRemoveOutline,
} from '@mdi/js'
import classNames from 'classnames'

import type { TranscriptJSON } from '@sourcegraph/cody-shared/dist/chat/transcript'
import { useLazyQuery, useQuery } from '@sourcegraph/http-client'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import {
    Icon,
    Popover,
    PopoverTrigger,
    PopoverContent,
    Button,
    Card,
    Text,
    Input,
    Tooltip,
    Link,
    useDebounce,
    Position,
    Flipping,
    Overlapping,
} from '@sourcegraph/wildcard'

import { useUserHistory } from '../../../components/useUserHistory'
import type {
    ContextSelectorRepoFields,
    SuggestedReposResult,
    SuggestedReposVariables,
    ReposSelectorSearchResult,
    ReposSelectorSearchVariables,
} from '../../../graphql-operations'
import { ExternalRepositoryIcon } from '../../../site-admin/components/ExternalRepositoryIcon'

import { SuggestedReposQuery, ReposSelectorSearchQuery } from './backend'
import { Callout } from './Callout'

import styles from './ScopeSelector.module.scss'

export interface IRepo {
    id: string
    name: string
    embeddingExists: boolean
    externalRepository: {
        serviceType: string
    }
    embeddingJobs?: {
        nodes: {
            id: string
            state: string
            failureMessage: string | null
        }[]
    } | null
}

export const MAX_ADDITIONAL_REPOSITORIES = 10
const MAX_SUGGESTED_REPOSITORIES = 10

// NOTE: The actual 10 most popular OSS repositories on GitHub are typically just
// plaintext resource collections like awesome lists or how-to-program tutorials, which
// are not likely to be as interesting to Cody users. Instead, we hardcode a list of
// repositories that are among the top 100 that contain actual source code.
const TOP_TEN_DOTCOM_REPOS = [
    'github.com/sourcegraph/sourcegraph',
    'github.com/freeCodeCamp/freeCodeCamp',
    'github.com/facebook/react',
    'github.com/tensorflow/tensorflow',
    'github.com/torvalds/linux',
    'github.com/microsoft/vscode',
    'github.com/flutter/flutter',
    'github.com/golang/go',
    'github.com/d3/d3',
    'github.com/kubernetes/kubernetes',
]

/**
 * removeDupes is an `Array.filter` predicate function which removes duplicate entries
 * from an array of objects based on the `name` property. It filters out any entries which
 * are not the first occurrence with a given `name`, which means it will preserve the
 * earliest occurrence of each.
 */
const removeDupes = (first: { name: string }, index: number, self: { name: string }[]): boolean =>
    index === self.findIndex(entry => entry.name === first.name)

export const RepositoriesSelectorPopover: React.FC<{
    includeInferredRepository: boolean
    includeInferredFile: boolean
    inferredRepository: IRepo | null
    inferredFilePath: string | null
    additionalRepositories: IRepo[]
    resetScope: (() => Promise<void>) | null
    addRepository: (repoName: string) => void
    removeRepository: (repoName: string) => void
    toggleIncludeInferredRepository: () => void
    toggleIncludeInferredFile: () => void
    // Whether to encourage the popover to overlap its trigger if necessary, rather than
    // collapsing or flipping position.
    encourageOverlap?: boolean
    transcriptHistory: TranscriptJSON[]
    authenticatedUser: AuthenticatedUser | null
}> = React.memo(function RepositoriesSelectorPopoverContent({
    inferredRepository,
    inferredFilePath,
    additionalRepositories,
    resetScope,
    addRepository,
    removeRepository,
    includeInferredRepository,
    includeInferredFile,
    toggleIncludeInferredRepository,
    toggleIncludeInferredFile,
    encourageOverlap = false,
    transcriptHistory,
    authenticatedUser,
}) {
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)
    const [searchText, setSearchText] = useState('')
    const searchTextDebounced = useDebounce(searchText, 300)

    const [searchRepositories, { data: searchResultsData, loading: loadingSearchResults, stopPolling }] = useLazyQuery<
        ReposSelectorSearchResult,
        ReposSelectorSearchVariables
    >(ReposSelectorSearchQuery, {})

    // TODO: Refactor out into custom hook
    const userHistory = useUserHistory(authenticatedUser?.id, false)
    const suggestedRepoNames: string[] = useMemo(() => {
        const flattenedTranscriptHistoryEntries = transcriptHistory
            .map(item => {
                const { scope, lastInteractionTimestamp } = item
                return (
                    // Return a new item for each repository in the scope.
                    scope?.repositories.map(name => ({
                        // Parse a date from the last interaction timestamp.
                        lastAccessed: new Date(lastInteractionTimestamp),
                        name,
                    })) || []
                )
            })
            .flat()
            // Remove duplicates.
            .filter(removeDupes)
            // We only need up to the first MAX_SUGGESTED_REPOSITORIES.
            .slice(0, MAX_SUGGESTED_REPOSITORIES)

        const userHistoryEntries =
            userHistory
                .loadEntries()
                .map(item => ({
                    name: item.repoName,
                    // Parse a date from the last acessed timestamp.
                    lastAccessed: new Date(item.lastAccessed),
                }))
                // We only need up to the first MAX_SUGGESTED_REPOSITORIES.
                .slice(0, MAX_SUGGESTED_REPOSITORIES) || []

        // We also take a list of 10 of the most popular OSS repositories on GitHub to
        // fill in if we have fewer than MAX_SUGGESTED_REPOSITORIES. This is mostly
        // relevant for dotcom but could fill in for any instance which indexes GitHub OSS
        // repositories. If the repositories are not indexed on the instance, they will
        // not return any results from the search API and thus will not be included in the
        // final list of suggestions.
        const topTenDotcomRepos = TOP_TEN_DOTCOM_REPOS.map(name => ({
            name,
            // We order by most recently accessed; these should always be ranked last.
            lastAccessed: new Date(0),
        }))

        // Merge the lists.
        const merged = [...flattenedTranscriptHistoryEntries, ...userHistoryEntries, ...topTenDotcomRepos]
            // Sort by most recently accessed.
            .sort((a, b) => b.lastAccessed.getTime() - a.lastAccessed.getTime())
            // Remove duplicates.
            .filter(removeDupes)
            // Take the most recent MAX_SUGGESTED_REPOSITORIES.
            .slice(0, MAX_SUGGESTED_REPOSITORIES)

        // Return just the names.
        return merged.map(({ name }) => name)
    }, [transcriptHistory, userHistory])

    // Query for the suggested repositories.
    const { data: suggestedReposData } = useQuery<SuggestedReposResult, SuggestedReposVariables>(SuggestedReposQuery, {
        variables: {
            names: suggestedRepoNames,
            includeJobs: !!authenticatedUser?.siteAdmin,
        },
        fetchPolicy: 'cache-first',
    })
    // Filter out and reorder the suggested repository results.
    const displaySuggestedRepos: ContextSelectorRepoFields[] = useMemo(() => {
        if (!suggestedReposData) {
            return []
        }

        const nodes = [...suggestedReposData.byName.nodes]
        // The order of the by-name repos returned by the search API will not match the
        // order of suggestions we intend to display (the ordering of suggestedRepoNames),
        // since the default ordering of the search API is alphabetical. Thus, we reorder
        // them to match the initial ordering of suggestedRepoNames.
        const sortedByNameNodes = nodes.sort(
            (a, b) => suggestedRepoNames.indexOf(a.name) - suggestedRepoNames.indexOf(b.name)
        )
        // Make sure we have a full MAX_SUGGESTED_REPOSITORIES to display in the
        // suggestions. We'll prioritize the repositories we looked up by name, and then
        // fill in the rest from the first 10 embedded repositories returned by the search
        // API.
        return (
            [...sortedByNameNodes, ...suggestedReposData.firstTen.nodes]
                // Remove any duplicates.
                .filter(removeDupes)
                // Take the first MAX_SUGGESTED_REPOSITORIES.
                .slice(0, MAX_SUGGESTED_REPOSITORIES)
                // Finally, filter out repositories that are already selected in the chat
                // context.
                .filter(
                    suggestion => !additionalRepositories.find(alreadyAdded => alreadyAdded.name === suggestion.name)
                )
        )
    }, [suggestedReposData, suggestedRepoNames, additionalRepositories])

    const searchResults = useMemo(() => searchResultsData?.repositories.nodes || [], [searchResultsData])

    const onSearch = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            setSearchText(event.target.value.trim())
        },
        [setSearchText]
    )

    const clearSearchText = useCallback(() => {
        setSearchText('')
        stopPolling()
    }, [setSearchText, stopPolling])

    useEffect(() => {
        if (searchTextDebounced) {
            /* eslint-disable no-console */
            searchRepositories({
                variables: { query: searchTextDebounced, includeJobs: !!authenticatedUser?.siteAdmin },
                pollInterval: 5000,
            }).catch(console.error)
            /* eslint-enable no-console */
        }
    }, [searchTextDebounced, searchRepositories, authenticatedUser?.siteAdmin])

    const [isCalloutDismissed = true, setIsCalloutDismissed] = useTemporarySetting(
        'cody.contextCallout.dismissed',
        false
    )

    const onOpenChange = useCallback(
        (event: { isOpen: boolean }) => {
            setIsPopoverOpen(event.isOpen)
            if (!event.isOpen) {
                setSearchText('')
                stopPolling()
            }

            if (!isCalloutDismissed) {
                setIsCalloutDismissed(true)
            }
        },
        [setIsPopoverOpen, setSearchText, isCalloutDismissed, setIsCalloutDismissed, stopPolling]
    )

    const netRepositories: IRepo[] = useMemo(() => {
        const names = []
        if (
            includeInferredRepository &&
            inferredRepository &&
            !additionalRepositories.find(repo => repo.name === inferredRepository.name)
        ) {
            names.push(inferredRepository)
        }
        return [...names, ...additionalRepositories]
    }, [includeInferredRepository, inferredRepository, additionalRepositories])

    const additionalRepositoriesLeft = Math.max(MAX_ADDITIONAL_REPOSITORIES - additionalRepositories.length, 0)

    const scopeChanged = additionalRepositories.length !== 0 || !includeInferredFile || !includeInferredRepository

    return (
        <>
            <Popover isOpen={isPopoverOpen} onOpenChange={onOpenChange}>
                <PopoverTrigger
                    as={Button}
                    variant="secondary"
                    size="sm"
                    outline={true}
                    className={classNames(
                        'd-flex p-1 align-items-center w-100 text-muted font-weight-normal',
                        styles.repositoryNamesText
                    )}
                >
                    <div className="mr-1">
                        <EmbeddingStatusIndicator repos={netRepositories} />
                    </div>
                    <div className="text-truncate mr-1">
                        {netRepositories.length > 1 ? (
                            <>
                                {netRepositories.length} Repositories (
                                {netRepositories.map(({ name }) => getFileName(name)).join(', ')})
                            </>
                        ) : netRepositories.length ? (
                            getFileName(netRepositories[0].name)
                        ) : (
                            'Add repositories to the chat context'
                        )}
                    </div>
                    <Icon
                        aria-hidden={true}
                        svgPath={isPopoverOpen ? mdiChevronUp : mdiChevronDown}
                        className="ml-auto"
                    />
                </PopoverTrigger>

                {/* We can try to explicitly encourage the popover to only appear beneath its
                    trigger by restricting it only permitting the Flipping.opposite strategy
                    and allowing overlap if necessary. Otherwise, on smaller viewports, the
                    popover may wind up partially below the initially visible scroll area, or
                    else awkwardly scrunched up to the left or right of the trigger. */}
                <PopoverContent
                    position={Position.bottomStart}
                    flipping={encourageOverlap ? Flipping.opposite : undefined}
                    overlapping={encourageOverlap ? Overlapping.all : undefined}
                >
                    <Card
                        className={classNames(
                            'd-flex flex-column justify-content-between',
                            styles.repositorySelectorContent
                        )}
                    >
                        {!searchText && (
                            <>
                                <div className="d-flex justify-content-between p-2 border-bottom mb-1">
                                    <Text className={classNames('m-0', styles.header)}>Chat Context</Text>
                                    {resetScope && scopeChanged && (
                                        <Button
                                            onClick={resetScope}
                                            variant="icon"
                                            aria-label="Reset scope"
                                            title="Reset scope"
                                            className={styles.header}
                                        >
                                            Reset
                                        </Button>
                                    )}
                                </div>
                                <div className={classNames('d-flex flex-column', styles.contextItemsContainer)}>
                                    {inferredRepository && (
                                        <div className="d-flex flex-column">
                                            {inferredFilePath && (
                                                <button
                                                    type="button"
                                                    className={classNames(
                                                        'd-flex justify-content-between flex-row text-truncate px-2 py-1 mt-1',
                                                        styles.repositoryListItem,
                                                        {
                                                            [styles.notIncludedInContext]: !includeInferredFile,
                                                        }
                                                    )}
                                                    onClick={toggleIncludeInferredFile}
                                                >
                                                    <div className="d-flex align-items-center text-truncate">
                                                        <Icon
                                                            aria-hidden={true}
                                                            className={classNames('mr-1 text-muted', {
                                                                [styles.visibilityHidden]: !includeInferredFile,
                                                            })}
                                                            svgPath={mdiCheck}
                                                        />
                                                        <ExternalRepositoryIcon
                                                            externalRepo={inferredRepository.externalRepository}
                                                            className={styles.repoIcon}
                                                        />
                                                        <span className="text-truncate">
                                                            {getFileName(inferredFilePath)}
                                                        </span>
                                                    </div>
                                                    <EmbeddingExistsIcon
                                                        repo={inferredRepository}
                                                        authenticatedUser={authenticatedUser}
                                                    />
                                                </button>
                                            )}
                                            <button
                                                type="button"
                                                className={classNames(
                                                    'd-flex justify-content-between flex-row text-truncate px-2 py-1',
                                                    styles.repositoryListItem,
                                                    {
                                                        [styles.notIncludedInContext]: !includeInferredRepository,
                                                    }
                                                )}
                                                onClick={toggleIncludeInferredRepository}
                                            >
                                                <div className="d-flex align-items-center text-truncate">
                                                    <Icon
                                                        aria-hidden={true}
                                                        className={classNames('mr-1 text-muted', {
                                                            [styles.visibilityHidden]: !includeInferredRepository,
                                                        })}
                                                        svgPath={mdiCheck}
                                                    />
                                                    <ExternalRepositoryIcon
                                                        externalRepo={inferredRepository.externalRepository}
                                                        className={styles.repoIcon}
                                                    />
                                                    <span className="text-truncate">
                                                        {getRepoName(inferredRepository.name)}
                                                    </span>
                                                </div>
                                                <EmbeddingExistsIcon
                                                    repo={inferredRepository}
                                                    authenticatedUser={authenticatedUser}
                                                />
                                            </button>
                                        </div>
                                    )}
                                    {!!additionalRepositories.length && (
                                        <div className="d-flex flex-column">
                                            <Text
                                                className={classNames(
                                                    'mb-0 px-2 py-1 text-muted d-flex justify-content-between',
                                                    styles.subHeader,
                                                    {
                                                        'mt-1': inferredRepository || inferredFilePath,
                                                    }
                                                )}
                                            >
                                                <span className="small">
                                                    {inferredRepository ? 'Additional repositories' : 'Repositories'}
                                                </span>
                                                <span className="small">{additionalRepositories.length}/10</span>
                                            </Text>
                                            {additionalRepositories.map(repository => (
                                                <AdditionalRepositoriesListItem
                                                    key={repository.id}
                                                    repository={repository}
                                                    removeRepository={removeRepository}
                                                    authenticatedUser={authenticatedUser}
                                                />
                                            ))}
                                        </div>
                                    )}

                                    {!inferredRepository && !inferredFilePath && !additionalRepositories.length && (
                                        <div className="d-flex align-items-center justify-content-center flex-column p-4 mt-4">
                                            <Text size="small" className="m-0 text-center text-muted">
                                                Add up to 10 repositories for Cody to reference when providing answers.
                                            </Text>
                                        </div>
                                    )}
                                </div>
                            </>
                        )}
                        {searchText && (
                            <>
                                <div className="d-flex justify-content-between p-2 border-bottom mb-1">
                                    <Text className={classNames('m-0', styles.header)}>
                                        {additionalRepositoriesLeft
                                            ? `Add up to ${additionalRepositoriesLeft} additional repositories`
                                            : 'Maximum additional repositories added'}
                                    </Text>
                                </div>
                                <div className={classNames('d-flex flex-column', styles.contextItemsContainer)}>
                                    {searchResults.length ? (
                                        searchResults.map(repository => (
                                            <SearchResultsListItem
                                                additionalRepositories={additionalRepositories}
                                                key={repository.id}
                                                repository={repository}
                                                searchText={searchText}
                                                addRepository={addRepository}
                                                removeRepository={removeRepository}
                                                authenticatedUser={authenticatedUser}
                                            />
                                        ))
                                    ) : !loadingSearchResults ? (
                                        <div className="d-flex align-items-center justify-content-center flex-column p-4 mt-4">
                                            <Text size="small" className="m-0 d-flex text-center">
                                                No matching repositories found
                                            </Text>
                                        </div>
                                    ) : null}
                                </div>
                            </>
                        )}
                        <div className={classNames('relative p-2 border-top mt-auto', styles.inputContainer)}>
                            <Input
                                role="combobox"
                                autoFocus={true}
                                autoComplete="off"
                                spellCheck="false"
                                placeholder={
                                    additionalRepositoriesLeft
                                        ? inferredRepository
                                            ? 'Add additional repositories...'
                                            : 'Add repositories...'
                                        : 'Maximum additional repositories added'
                                }
                                variant="small"
                                disabled={!searchText && !additionalRepositoriesLeft}
                                value={searchText}
                                onChange={onSearch}
                            />
                            {!!searchText && (
                                <Button
                                    className={classNames(
                                        'd-flex p-1 align-items-center justify-content-center',
                                        styles.clearButton
                                    )}
                                    variant="icon"
                                    onClick={clearSearchText}
                                    aria-label="Clear"
                                >
                                    <Icon aria-hidden={true} svgPath={mdiCloseCircle} />
                                </Button>
                            )}
                        </div>
                    </Card>
                </PopoverContent>
            </Popover>
            {!isCalloutDismissed && <Callout dismiss={() => setIsCalloutDismissed(true)} />}
        </>
    )
})

const AdditionalRepositoriesListItem: React.FC<{
    repository: IRepo
    removeRepository: (repoName: string) => void
    authenticatedUser: AuthenticatedUser | null
}> = React.memo(function RepositoryListItemContent({ repository, removeRepository, authenticatedUser }) {
    const onClick = useCallback(() => {
        removeRepository(repository.name)
    }, [repository, removeRepository])

    return (
        <button
            type="button"
            className={classNames(
                'd-flex justify-content-between flex-row text-truncate px-2 py-1 mb-1',
                styles.repositoryListItem
            )}
            onClick={onClick}
        >
            <div className="d-flex align-items-center text-truncate">
                <Icon
                    aria-hidden={true}
                    className={classNames('mr-1 text-muted', styles.removeRepoIcon)}
                    svgPath={mdiMinusCircleOutline}
                />
                <ExternalRepositoryIcon externalRepo={repository.externalRepository} className={styles.repoIcon} />
                <span className="text-truncate">{getRepoName(repository.name)}</span>
            </div>
            <EmbeddingExistsIcon repo={repository} authenticatedUser={authenticatedUser} />
        </button>
    )
})

const SearchResultsListItem: React.FC<{
    additionalRepositories: IRepo[]
    repository: IRepo
    searchText: string
    addRepository: (repoName: string) => void
    removeRepository: (repoName: string) => void
    authenticatedUser: AuthenticatedUser | null
}> = React.memo(function RepositoryListItemContent({
    additionalRepositories,
    repository,
    searchText,
    addRepository,
    removeRepository,
    authenticatedUser,
}) {
    const selected = useMemo(
        () => !!additionalRepositories.find(({ name }) => name === repository.name),
        [additionalRepositories, repository]
    )

    const onClick = useCallback(() => {
        if (selected) {
            removeRepository?.(repository.name)
        } else {
            addRepository?.(repository.name)
        }
    }, [selected, repository, addRepository, removeRepository])

    const disabled = additionalRepositories.length >= MAX_ADDITIONAL_REPOSITORIES && !selected

    return (
        <button
            type="button"
            className={classNames(
                'd-flex justify-content-between flex-row text-truncate px-2 py-1 mb-1',
                styles.repositoryListItem,
                { [styles.disabledSearchResult]: disabled }
            )}
            disabled={disabled}
            onClick={onClick}
        >
            <div className="text-truncate">
                <Icon
                    aria-hidden={true}
                    className={classNames('mr-1 text-muted', { [styles.visibilityHidden]: !selected })}
                    svgPath={mdiCheck}
                />
                <ExternalRepositoryIcon externalRepo={repository.externalRepository} className={styles.repoIcon} />
                {getTintedText(getRepoName(repository.name), searchText)}
            </div>
            <EmbeddingExistsIcon repo={repository} authenticatedUser={authenticatedUser} />
        </button>
    )
})

const getTintedText = (item: string, searchText: string): React.ReactNode => {
    const searchRegex = new RegExp(`(${searchText})`, 'gi')

    const matches = item.match(searchRegex)
    return (
        <span className="text-truncate">
            {item
                .replace(searchRegex, '$')
                .split('$')
                .reduce((spans, unmatched, index) => {
                    spans.push(<span key={`unmatched-${index}`}>{unmatched}</span>)
                    if (matches?.[index]) {
                        spans.push(
                            <span key={`matched-${index}`} className={styles.tintedSearch}>
                                {matches[index]}
                            </span>
                        )
                    }
                    return spans
                }, [] as React.ReactElement[])}
        </span>
    )
}

export const getFileName = (path: string): string => {
    const parts = path.split('/')
    return parts[parts.length - 1]
}

export const getRepoName = (path: string): string => {
    const parts = path.split('/')
    return parts.slice(-2).join('/')
}

enum RepoEmbeddingStatus {
    NOT_INDEXED = 'NOT_INDEXED',
    QUEUED = 'QUEUED',
    INDEXING = 'INDEXING',
    INDEXED = 'INDEXED',
    FAILED = 'FAILED',
}

const getEmbeddingStatus = ({
    embeddingExists,
    embeddingJobs,
}: IRepo): { status: RepoEmbeddingStatus; tooltip: string; icon: string; className: string } => {
    if (embeddingExists) {
        return {
            status: RepoEmbeddingStatus.INDEXED,
            tooltip: 'Repository is indexed',
            icon: mdiDatabaseCheckOutline,
            className: 'text-success',
        }
    }

    if (!embeddingJobs?.nodes.length) {
        return {
            status: RepoEmbeddingStatus.NOT_INDEXED,
            tooltip: 'Repository is not indexed',
            icon: mdiDatabaseRemoveOutline,
            className: 'text-warning',
        }
    }

    const job = embeddingJobs.nodes[0]
    switch (job.state) {
        case 'QUEUED':
            return {
                status: RepoEmbeddingStatus.QUEUED,
                tooltip: 'Repository is queued for indexing',
                icon: mdiDatabaseClockOutline,
                className: 'text-warning',
            }
        case 'PROCESSING':
            return {
                status: RepoEmbeddingStatus.INDEXING,
                tooltip: 'Repository is being indexed',
                icon: mdiDatabaseRefreshOutline,
                className: 'text-warning',
            }
        case 'COMPLETED':
            return {
                status: RepoEmbeddingStatus.INDEXED,
                tooltip: 'Repository is indexed',
                icon: mdiDatabaseCheckOutline,
                className: 'text-success',
            }
        case 'ERRORED':
            return {
                status: RepoEmbeddingStatus.FAILED,
                tooltip: `Repository indexing failed: ${job.failureMessage || 'unknown error'}`,
                icon: mdiDatabaseRemoveOutline,
                className: 'text-danger',
            }
        case 'FAILED':
            return {
                status: RepoEmbeddingStatus.FAILED,
                tooltip: `Repository indexing failed: ${job.failureMessage || 'unknown error'}`,
                icon: mdiDatabaseRemoveOutline,
                className: 'text-danger',
            }
        default:
            return {
                status: RepoEmbeddingStatus.NOT_INDEXED,
                tooltip: 'Repository is not indexed',
                icon: mdiDatabaseRemoveOutline,
                className: 'text-warning',
            }
    }
}

export const isRepoIndexed = (repo: IRepo): boolean => getEmbeddingStatus(repo).status === RepoEmbeddingStatus.INDEXED

const EmbeddingExistsIcon: React.FC<{
    repo: IRepo
    authenticatedUser: AuthenticatedUser | null
}> = React.memo(function EmbeddingExistsIconContent({ repo, authenticatedUser }) {
    const { tooltip, icon, className } = getEmbeddingStatus(repo)

    return (
        <Tooltip content={tooltip}>
            {authenticatedUser?.siteAdmin ? (
                <Link to="/site-admin/embeddings" className="text-body" onClick={event => event.stopPropagation()}>
                    <Icon aria-hidden={true} className={classNames(styles.icon, className)} svgPath={icon} />
                </Link>
            ) : (
                <Icon aria-hidden={true} className={classNames(styles.icon, className)} svgPath={icon} />
            )}
        </Tooltip>
    )
})

const EmbeddingStatusIndicator: React.FC<{ repos: IRepo[] }> = React.memo(function EmbeddingsStatusIndicatorContent({
    repos,
}) {
    const repoStatusCounts = useMemo(
        () =>
            repos.reduce(
                (statuses, repo) => {
                    const { status } = getEmbeddingStatus(repo)
                    statuses[status] = statuses[status] + 1
                    return statuses
                },
                {
                    [RepoEmbeddingStatus.NOT_INDEXED]: 0,
                    [RepoEmbeddingStatus.INDEXING]: 0,
                    [RepoEmbeddingStatus.INDEXED]: 0,
                    [RepoEmbeddingStatus.QUEUED]: 0,
                    [RepoEmbeddingStatus.FAILED]: 0,
                } as Record<RepoEmbeddingStatus, number>
            ),
        [repos]
    )

    if (!repos.length) {
        return (
            <Icon
                aria-label="Database icon"
                className={classNames('align-center text-muted', styles.icon)}
                svgPath={mdiDatabaseOutline}
            />
        )
    }

    if (repos.length === 1) {
        const { tooltip, icon, className } = getEmbeddingStatus(repos[0])

        return (
            <Tooltip content={tooltip}>
                <Icon
                    aria-label="Database icon"
                    className={classNames('align-center', styles.icon, className)}
                    svgPath={icon}
                />
            </Tooltip>
        )
    }

    if (repoStatusCounts[RepoEmbeddingStatus.FAILED]) {
        return (
            <Tooltip content="Indexing failed for some repositories">
                <Icon
                    aria-label="Database icon"
                    className={classNames('align-center text-danger', styles.icon)}
                    svgPath={mdiDatabaseRemoveOutline}
                />
            </Tooltip>
        )
    }

    if (repoStatusCounts[RepoEmbeddingStatus.INDEXING]) {
        return (
            <Tooltip content="Some repositories are being indexed">
                <Icon
                    aria-label="Database icon"
                    className={classNames('align-center text-warning', styles.icon)}
                    svgPath={mdiDatabaseRefreshOutline}
                />
            </Tooltip>
        )
    }

    if (repoStatusCounts[RepoEmbeddingStatus.QUEUED]) {
        return (
            <Tooltip content="Some repositories are queued for indexing">
                <Icon
                    aria-label="Database icon"
                    className={classNames('align-center text-warning', styles.icon)}
                    svgPath={mdiDatabaseClockOutline}
                />
            </Tooltip>
        )
    }

    if (repoStatusCounts[RepoEmbeddingStatus.NOT_INDEXED]) {
        return (
            <Tooltip content="Some repositories are not indexed">
                <Icon
                    aria-label="Database icon"
                    className={classNames('align-center text-warning', styles.icon)}
                    svgPath={mdiDatabaseRemoveOutline}
                />
            </Tooltip>
        )
    }

    return (
        <Tooltip content="All repositories are indexed">
            <Icon
                aria-label="Database icon with a check mark"
                className={classNames('align-center text-success', styles.icon)}
                svgPath={mdiDatabaseCheckOutline}
            />
        </Tooltip>
    )
})
