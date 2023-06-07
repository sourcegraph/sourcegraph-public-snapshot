import React, { useState, useCallback, useEffect, useMemo } from 'react'

import {
    mdiChevronUp,
    mdiMinusCircleOutline,
    mdiClose,
    mdiCheck,
    mdiCloseCircle,
    mdiDatabaseCheckOutline,
    mdiDatabaseRemoveOutline,
} from '@mdi/js'
import classNames from 'classnames'

import { useLazyQuery } from '@sourcegraph/http-client'
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
    useDebounce,
} from '@sourcegraph/wildcard'

import { ReposSelectorSearchResult, ReposSelectorSearchVariables } from '../../../graphql-operations'
import { ExternalRepositoryIcon } from '../../../site-admin/components/ExternalRepositoryIcon'

import { ReposSelectorSearchQuery } from './backend'

import styles from './ScopeSelector.module.scss'

export interface IRepo {
    id: string
    name: string
    embeddingExists: boolean
    externalRepository: {
        serviceType: string
    }
}

export const MAX_ADDITIONAL_REPOSITORIES = 10

export const RepositoriesSelectorPopover: React.FC<{
    includeInferredRepository: boolean
    includeInferredFile: boolean
    inferredRepository: IRepo | null
    inferredFilePath: string | null
    additionalRepositories: IRepo[]
    addRepository: (repoName: string) => void
    removeRepository: (repoName: string) => void
    toggleIncludeInferredRepository: () => void
    toggleIncludeInferredFile: () => void
}> = React.memo(function RepositoriesSelectorPopoverContent({
    inferredRepository,
    inferredFilePath,
    additionalRepositories,
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

    const [searchRepositories, { data: searchResultsData, loading: loadingSearchResults }] = useLazyQuery<
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

    useEffect(() => {
        if (searchTextDebounced) {
            /* eslint-disable no-console */
            searchRepositories({ variables: { query: searchTextDebounced } }).catch(console.error)
            /* eslint-enable no-console */
        }
    }, [searchTextDebounced, searchRepositories])

    const clearSearchText = useCallback(() => setSearchText(''), [setSearchText])

    const onOpenChange = useCallback(
        (event: { isOpen: boolean }) => {
            setIsPopoverOpen(event.isOpen)
            if (!event.isOpen) {
                setSearchText('')
            }
        },
        [setIsPopoverOpen, setSearchText]
    )

    const repositoryNames = useMemo(() => {
        const names = []
        if (includeInferredRepository && inferredRepository) {
            names.push(inferredRepository.name)
        }
        return [...names, ...additionalRepositories.map(repo => repo.name)]
    }, [includeInferredRepository, inferredRepository, additionalRepositories])

    const additionalRepositoriesLeft = Math.max(MAX_ADDITIONAL_REPOSITORIES - additionalRepositories.length, 0)

    return (
        <Popover isOpen={isPopoverOpen} onOpenChange={onOpenChange}>
            <PopoverTrigger
                as={Button}
                outline={false}
                className={classNames(
                    'd-flex justify-content-between p-0 align-items-center w-100',
                    styles.repositoryNamesText
                )}
            >
                <div
                    className={classNames('text-truncate mr-1', {
                        'text-muted': !repositoryNames.length,
                    })}
                >
                    {repositoryNames.length > 1 ? (
                        <>
                            {repositoryNames.length} Repositories (
                            {repositoryNames?.map(repoName => getFileName(repoName)).join(', ')})
                        </>
                    ) : repositoryNames.length ? (
                        getFileName(repositoryNames[0])
                    ) : (
                        'Add repositories to the scope...'
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
                                    <Text className={classNames('m-0 text-uppercase', styles.header)}>
                                        Chat Context
                                    </Text>
                                    <Button
                                        onClick={() => setIsPopoverOpen(false)}
                                        variant="icon"
                                        aria-label="Close"
                                        className="text-muted"
                                    >
                                        <Icon aria-hidden={true} svgPath={mdiClose} />
                                    </Button>
                                </div>
                                <div className={classNames('d-flex flex-column', styles.contextItemsContainer)}>
                                    {inferredRepository && (
                                        <div className="d-flex flex-column">
                                            <button
                                                type="button"
                                                className={classNames(
                                                    'd-flex justify-content-between flex-row bg-transparent outline-none border-0 text-truncate w-100 px-2 py-1',
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
                                                        className={classNames('mr-2 text-muted', {
                                                            [styles.visibilityHidden]: !includeInferredRepository,
                                                        })}
                                                        svgPath={mdiCheck}
                                                    />
                                                    <ExternalRepositoryIcon
                                                        externalRepo={inferredRepository.externalRepository}
                                                        className="text-muted"
                                                    />
                                                    <span>{getRepoName(inferredRepository.name)}</span>
                                                </div>
                                                <Icon
                                                    aria-hidden={true}
                                                    className={classNames({
                                                        'text-muted': inferredRepository.embeddingExists,
                                                        'text-warning': !inferredRepository.embeddingExists,
                                                    })}
                                                    svgPath={
                                                        inferredRepository.embeddingExists
                                                            ? mdiDatabaseCheckOutline
                                                            : mdiDatabaseRemoveOutline
                                                    }
                                                />
                                            </button>
                                            {inferredFilePath && (
                                                <button
                                                    type="button"
                                                    className={classNames(
                                                        'd-flex justify-content-between flex-row bg-transparent outline-none border-0 text-truncate w-100 px-2 py-1 mt-1',
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
                                                            className={classNames('mr-2 text-muted', {
                                                                [styles.visibilityHidden]: !includeInferredFile,
                                                            })}
                                                            svgPath={mdiCheck}
                                                        />
                                                        <ExternalRepositoryIcon
                                                            externalRepo={inferredRepository.externalRepository}
                                                            className="text-muted"
                                                        />
                                                        <span>{getFileName(inferredFilePath)}</span>
                                                    </div>
                                                    <Icon
                                                        aria-hidden={true}
                                                        className={classNames({
                                                            'text-muted': inferredRepository.embeddingExists,
                                                            'text-warning': !inferredRepository.embeddingExists,
                                                        })}
                                                        svgPath={
                                                            inferredRepository.embeddingExists
                                                                ? mdiDatabaseCheckOutline
                                                                : mdiDatabaseRemoveOutline
                                                        }
                                                    />
                                                </button>
                                            )}
                                        </div>
                                    )}
                                    {!!additionalRepositories.length && (
                                        <div className="d-flex flex-column">
                                            <Text
                                                className={classNames(
                                                    'mb-1 px-2 py-1 text-muted d-flex justify-content-between',
                                                    styles.subHeader
                                                )}
                                            >
                                                <span>{inferredRepository ? 'Additional' : ''} Repositories</span>
                                                <span>({additionalRepositories.length}/10)</span>
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
                                                Add upto 10 repositories for Cody to reference when providing answers.
                                            </Text>
                                        </div>
                                    )}
                                </div>
                            </>
                        )}
                        {searchText && (
                            <>
                                <div className="d-flex justify-content-between p-2 border-bottom mb-1">
                                    <Text className={classNames('m-0 text-uppercase', styles.header)}>
                                        {additionalRepositoriesLeft
                                            ? `Add upto ${additionalRepositoriesLeft} Additional Repositories`
                                            : 'Maximum additional epositories added'}
                                    </Text>
                                    <Button
                                        onClick={() => setIsPopoverOpen(false)}
                                        variant="icon"
                                        aria-label="Close"
                                        className="text-muted"
                                    >
                                        <Icon aria-hidden={true} svgPath={mdiClose} />
                                    </Button>
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
                                    ? 'Add repositries...'
                                    : 'Maximum additional repositories added'
                            }
                            variant="small"
                            disabled={!additionalRepositoriesLeft}
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
                'd-flex justify-content-between flex-row bg-transparent outline-none border-0 text-truncate w-100 px-2 py-1 mb-1',
                styles.repositoryListItem
            )}
            onClick={onClick}
        >
            <div className="d-flex align-items-center text-truncate">
                <Icon
                    aria-hidden={true}
                    className={classNames('mr-2 text-muted', styles.removeRepoIcon)}
                    svgPath={mdiMinusCircleOutline}
                />
                <ExternalRepositoryIcon externalRepo={repository.externalRepository} className="text-muted" />
                <span>{getRepoName(repository.name)}</span>
            </div>
            <Icon
                aria-hidden={true}
                className={classNames({
                    'text-muted': repository.embeddingExists,
                    'text-warning': !repository.embeddingExists,
                })}
                svgPath={repository.embeddingExists ? mdiDatabaseCheckOutline : mdiDatabaseRemoveOutline}
            />
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
                'd-flex justify-content-between flex-row bg-transparent outline-none border-0 text-truncate w-100 px-2 py-1 mb-1',
                styles.repositoryListItem,
                { [styles.disabledSearchResult]: disabled }
            )}
            disabled={disabled}
            onClick={onClick}
        >
            <div className="text-truncate">
                <Icon
                    aria-hidden={true}
                    className={classNames('mr-2 text-success', { [styles.visibilityHidden]: !selected })}
                    svgPath={mdiCheck}
                />
                <ExternalRepositoryIcon externalRepo={repository.externalRepository} className="text-muted" />
                {getTintedText(getRepoName(repository.name), searchText)}
            </div>
            <Icon
                aria-hidden={true}
                className={classNames({
                    'text-muted': repository.embeddingExists,
                    'text-warning': !repository.embeddingExists,
                })}
                svgPath={repository.embeddingExists ? mdiDatabaseCheckOutline : mdiDatabaseRemoveOutline}
            />
        </button>
    )
})

const getTintedText = (item: string, searchText: string): React.ReactNode => {
    const searchRegex = new RegExp(`(${searchText})`, 'gi')

    const matches = item.match(searchRegex)
    return (
        <span>
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
