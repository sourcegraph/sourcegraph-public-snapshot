import React, { useState, useCallback, useEffect, useMemo } from 'react'

import {
    mdiChevronUp,
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

import { useLazyQuery } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import {
    Icon,
    Popover,
    PopoverTrigger,
    PopoverContent,
    Position,
    Button,
    Card,
    Text,
    Input,
    Tooltip,
    Link,
    useDebounce,
} from '@sourcegraph/wildcard'

import { ReposSelectorSearchResult, ReposSelectorSearchVariables } from '../../../graphql-operations'
import { ExternalRepositoryIcon } from '../../../site-admin/components/ExternalRepositoryIcon'

import { ReposSelectorSearchQuery } from './backend'
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

export const RepositoriesSelectorPopover: React.FC<{
    includeInferredRepository: boolean
    includeInferredFile: boolean
    inferredRepository: IRepo | null
    inferredFilePath: string | null
    additionalRepositories: IRepo[]
    resetScope: () => Promise<void>
    addRepository: (repoName: string) => void
    removeRepository: (repoName: string) => void
    toggleIncludeInferredRepository: () => void
    toggleIncludeInferredFile: () => void
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
}) {
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)
    const [searchText, setSearchText] = useState('')
    const searchTextDebounced = useDebounce(searchText, 300)

    const [searchRepositories, { data: searchResultsData, loading: loadingSearchResults, stopPolling }] = useLazyQuery<
        ReposSelectorSearchResult,
        ReposSelectorSearchVariables
    >(ReposSelectorSearchQuery, {})

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
                variables: { query: searchTextDebounced, includeJobs: !!window.context.currentUser?.siteAdmin },
                pollInterval: 5000,
            }).catch(console.error)
            /* eslint-enable no-console */
        }
    }, [searchTextDebounced, searchRepositories])

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
                    outline={false}
                    className={classNames(
                        'd-flex justify-content-between p-0 pr-1 align-items-center w-100',
                        styles.repositoryNamesText
                    )}
                >
                    <div className="mr-1">
                        <EmbeddingStatusIndicator repos={netRepositories} />
                    </div>
                    <div
                        className={classNames('text-truncate mr-1', {
                            'text-muted': !netRepositories.length,
                        })}
                    >
                        {netRepositories.length > 1 ? (
                            <>
                                {netRepositories.length} Repositories (
                                {netRepositories.map(({ name }) => getFileName(name)).join(', ')})
                            </>
                        ) : netRepositories.length ? (
                            getFileName(netRepositories[0].name)
                        ) : (
                            'Add repositories...'
                        )}
                    </div>
                    <Icon aria-hidden={true} svgPath={mdiChevronUp} />
                </PopoverTrigger>

                <PopoverContent position={Position.topStart}>
                    <Card
                        className={classNames(
                            'd-flex flex-column justify-content-between',
                            styles.repositorySelectorContent
                        )}
                    >
                        <div>
                            {!searchText && (
                                <>
                                    <div className="d-flex justify-content-between p-2 border-bottom mb-1">
                                        <Text className={classNames('m-0', styles.header)}>Chat Context</Text>
                                        {scopeChanged && (
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
                                                        <EmbeddingExistsIcon repo={inferredRepository} />
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
                                                    <EmbeddingExistsIcon repo={inferredRepository} />
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
                                                        {inferredRepository
                                                            ? 'Additional repositories'
                                                            : 'Repositories'}
                                                    </span>
                                                    <span className="small">{additionalRepositories.length}/10</span>
                                                </Text>
                                                {additionalRepositories.map(repository => (
                                                    <AdditionalRepositoriesListItem
                                                        key={repository.id}
                                                        repository={repository}
                                                        removeRepository={removeRepository}
                                                    />
                                                ))}
                                            </div>
                                        )}

                                        {!inferredRepository && !inferredFilePath && !additionalRepositories.length && (
                                            <div className="d-flex align-items-center justify-content-center flex-column p-4 mt-4">
                                                <Text size="small" className="m-0 text-center text-muted">
                                                    Add up to 10 repositories for Cody to reference when providing
                                                    answers.
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
                        </div>
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
}> = React.memo(function RepositoryListItemContent({ repository, removeRepository }) {
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
            <EmbeddingExistsIcon repo={repository} />
        </button>
    )
})

const SearchResultsListItem: React.FC<{
    additionalRepositories: IRepo[]
    repository: IRepo
    searchText: string
    addRepository: (repoName: string) => void
    removeRepository: (repoName: string) => void
}> = React.memo(function RepositoryListItemContent({
    additionalRepositories,
    repository,
    searchText,
    addRepository,
    removeRepository,
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
            <EmbeddingExistsIcon repo={repository} />
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
            className: '',
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
                className: '',
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

const EmbeddingExistsIcon: React.FC<{
    repo: IRepo
}> = React.memo(function EmbeddingExistsIconContent({ repo }) {
    const { tooltip, icon, className } = getEmbeddingStatus(repo)

    return (
        <Tooltip content={tooltip}>
            {window.context.currentUser?.siteAdmin ? (
                <Link to="/site-admin/embeddings" className="text-body" onClick={event => event.stopPropagation()}>
                    <Icon aria-hidden={true} className={className} svgPath={icon} />
                </Link>
            ) : (
                <Icon aria-hidden={true} className={className} svgPath={icon} />
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
        return <Icon aria-label="Database icon" className="align-center text-muted" svgPath={mdiDatabaseOutline} />
    }

    if (repos.length === 1) {
        const { tooltip, icon, className } = getEmbeddingStatus(repos[0])

        return (
            <Tooltip content={tooltip}>
                <Icon aria-label="Database icon" className={classNames('align-center', className)} svgPath={icon} />
            </Tooltip>
        )
    }

    if (repoStatusCounts[RepoEmbeddingStatus.FAILED]) {
        return (
            <Tooltip content="Indexing failed for some repositories">
                <Icon
                    aria-label="Database icon"
                    className="align-center text-danger"
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
                    className="align-center text-warning"
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
                    className="align-center text-warning"
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
                    className="align-center text-warning"
                    svgPath={mdiDatabaseRemoveOutline}
                />
            </Tooltip>
        )
    }

    return (
        <Tooltip content="All repositories are indexed">
            <Icon
                aria-label="Database icon with a check mark"
                className="align-center"
                svgPath={mdiDatabaseCheckOutline}
            />
        </Tooltip>
    )
})
