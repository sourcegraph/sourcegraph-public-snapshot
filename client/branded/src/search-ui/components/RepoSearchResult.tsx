import React, { useEffect, useRef } from 'react'

import { mdiArchive, mdiLock, mdiSourceFork } from '@mdi/js'
import classNames from 'classnames'

import { highlightNode } from '@sourcegraph/common'
import { codeHostSubstrLength, displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import type { BuildSearchQueryURLParameters, QueryState } from '@sourcegraph/shared/src/search'
import { getRepoMatchLabel, getRepoMatchUrl, type RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { Icon, Link, Text } from '@sourcegraph/wildcard'

import { metadataToTag, TagList, topicToTag } from './RepoMetadata'
import { ResultContainer } from './ResultContainer'

import styles from './ResultContainer.module.scss'

const REPO_DESCRIPTION_CHAR_LIMIT = 500

export interface RepoSearchResultProps {
    result: RepositoryMatch
    onSelect: () => void
    buildSearchURLQueryFromQueryState?: (queryParameters: BuildSearchQueryURLParameters) => string
    queryState?: QueryState
    containerClassName?: string
    as?: React.ElementType
    index: number
    enableRepositoryMetadata?: boolean
}

export const RepoSearchResult: React.FunctionComponent<RepoSearchResultProps> = ({
    result,
    onSelect,
    containerClassName,
    as,
    index,
    enableRepositoryMetadata,
    buildSearchURLQueryFromQueryState,
    queryState,
}) => {
    const repoDescriptionElement = useRef<HTMLDivElement>(null)
    const repoNameElement = useRef<HTMLAnchorElement>(null)

    const title = (
        <div className={styles.title}>
            <span className={classNames('test-search-result-label', styles.titleInner)}>
                <Link to={getRepoMatchUrl(result)} ref={repoNameElement} data-selectable-search-result="true">
                    {displayRepoName(getRepoMatchLabel(result))}
                </Link>
            </span>
        </div>
    )
    const { description, topics, metadata, repository: repoName, descriptionMatches, repositoryMatches } = result

    useEffect((): void => {
        if (repoNameElement.current && repoName && repositoryMatches) {
            for (const range of repositoryMatches) {
                highlightNode(
                    repoNameElement.current as HTMLElement,
                    range.start.column - codeHostSubstrLength(repoName),
                    range.end.column - range.start.column
                )
            }
        }

        if (repoDescriptionElement.current && descriptionMatches) {
            for (const range of descriptionMatches) {
                highlightNode(
                    repoDescriptionElement.current as HTMLElement,
                    range.start.column,
                    range.end.column - range.start.column
                )
            }
        }
    }, [result, repositoryMatches, repoNameElement, description, descriptionMatches, repoDescriptionElement, repoName])

    const showExtraInfo = result.archived || result.fork || result.private

    const tags = [
        ...(metadata
            ? Object.entries(metadata).map(([key, value]) =>
                  metadataToTag({ key, value }, queryState, false, buildSearchURLQueryFromQueryState)
              )
            : []),
        ...(topics ? topics.map(topic => topicToTag(topic, queryState, false, buildSearchURLQueryFromQueryState)) : []),
    ]

    const showRepoMetadata = enableRepositoryMetadata && tags.length > 0

    return (
        <ResultContainer
            index={index}
            title={title}
            resultType={result.type}
            onResultClicked={onSelect}
            repoName={repoName}
            repoStars={result.repoStars}
            className={containerClassName}
            repoLastFetched={result.repoLastFetched}
            as={as}
        >
            {(showExtraInfo || description || showRepoMetadata) && (
                <div
                    data-testid="search-repo-result"
                    className={classNames(styles.searchResultMatch, styles.gap1, 'p-3 flex-column')}
                >
                    {showExtraInfo && (
                        <div className={classNames('d-flex', styles.dividerBetween)}>
                            {result.fork && (
                                <div className="d-flex align-items-center">
                                    <Icon
                                        aria-label="Forked repository"
                                        className={classNames('flex-shrink-0 text-muted mr-1')}
                                        svgPath={mdiSourceFork}
                                    />
                                    <small>Fork</small>
                                </div>
                            )}
                            {result.archived && (
                                <div className="d-flex align-items-center">
                                    <Icon
                                        aria-label="Archived repository"
                                        className={classNames('flex-shrink-0 text-muted mr-1')}
                                        svgPath={mdiArchive}
                                    />
                                    <small>Archived</small>
                                </div>
                            )}
                            {result.private && (
                                <div className="d-flex align-items-center">
                                    <Icon
                                        aria-label="Private repository"
                                        className={classNames('flex-shrink-0 text-muted mr-1')}
                                        svgPath={mdiLock}
                                    />
                                    <small>Private</small>
                                </div>
                            )}
                        </div>
                    )}
                    {description && (
                        <Text as="em" ref={repoDescriptionElement}>
                            {description.length > REPO_DESCRIPTION_CHAR_LIMIT
                                ? description.slice(0, REPO_DESCRIPTION_CHAR_LIMIT) + ' ...'
                                : description}
                        </Text>
                    )}
                    {showRepoMetadata && (
                        <div className="d-flex">
                            <TagList tags={tags} />
                        </div>
                    )}
                </div>
            )}
        </ResultContainer>
    )
}
