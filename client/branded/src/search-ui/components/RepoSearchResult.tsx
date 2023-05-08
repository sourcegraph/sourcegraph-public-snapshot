import React, { useEffect, useRef } from 'react'

import { mdiSourceFork, mdiArchive, mdiLock } from '@mdi/js'
import classNames from 'classnames'

import { highlightNode } from '@sourcegraph/common'
import { codeHostSubstrLength, displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { getRepoMatchLabel, getRepoMatchUrl, RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { Icon, Link } from '@sourcegraph/wildcard'

import { LastSyncedIcon } from './LastSyncedIcon'
import { RepoMetadata } from './RepoMetadata'
import { ResultContainer } from './ResultContainer'

import styles from './SearchResult.module.scss'

const REPO_DESCRIPTION_CHAR_LIMIT = 500

export interface RepoSearchResultProps {
    result: RepositoryMatch
    onSelect: () => void
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
}) => {
    const repoDescriptionElement = useRef<HTMLDivElement>(null)
    const repoNameElement = useRef<HTMLAnchorElement>(null)

    const title = (
        <div className={styles.title}>
            <span className={classNames('test-search-result-label', styles.titleInner, styles.mutedRepoFileLink)}>
                <Link to={getRepoMatchUrl(result)} ref={repoNameElement} data-selectable-search-result="true">
                    {displayRepoName(getRepoMatchLabel(result))}
                </Link>
            </span>
        </div>
    )

    useEffect((): void => {
        if (repoNameElement.current && result.repository && result.repositoryMatches) {
            for (const range of result.repositoryMatches) {
                highlightNode(
                    repoNameElement.current as HTMLElement,
                    range.start.column - codeHostSubstrLength(result.repository),
                    range.end.column - range.start.column
                )
            }
        }

        if (repoDescriptionElement.current && result.descriptionMatches) {
            for (const range of result.descriptionMatches) {
                highlightNode(
                    repoDescriptionElement.current as HTMLElement,
                    range.start.column,
                    range.end.column - range.start.column
                )
            }
        }
    }, [
        result,
        result.repositoryMatches,
        repoNameElement,
        result.description,
        result.descriptionMatches,
        repoDescriptionElement,
    ])

    return (
        <ResultContainer
            index={index}
            title={title}
            resultType={result.type}
            onResultClicked={onSelect}
            repoName={result.repository}
            repoStars={result.repoStars}
            className={containerClassName}
            as={as}
        >
            <div data-testid="search-repo-result">
                <div className={classNames(styles.searchResultMatch, 'p-2 flex-column')}>
                    {result.repoLastFetched && <LastSyncedIcon lastSyncedTime={result.repoLastFetched} />}
                    <div className="d-flex align-items-center flex-row">
                        <div className={styles.matchType}>
                            <small>Repository match</small>
                            {enableRepositoryMetadata && !!result.metadata && (
                                <RepoMetadata
                                    small={true}
                                    className="justify-content-end mt-1"
                                    items={Object.entries(result.metadata).map(([key, value]) => ({ key, value }))}
                                />
                            )}
                        </div>
                        {result.fork && (
                            <>
                                <div className={styles.divider} />
                                <div>
                                    <Icon
                                        aria-label="Forked repository"
                                        className={classNames('flex-shrink-0 text-muted', styles.icon)}
                                        svgPath={mdiSourceFork}
                                    />
                                </div>
                                <div>
                                    <small>Fork</small>
                                </div>
                            </>
                        )}
                        {result.archived && (
                            <>
                                <div className={styles.divider} />
                                <div>
                                    <Icon
                                        aria-label="Archived repository"
                                        className={classNames('flex-shrink-0 text-muted', styles.icon)}
                                        svgPath={mdiArchive}
                                    />
                                </div>
                                <div>
                                    <small>Archived</small>
                                </div>
                            </>
                        )}
                        {result.private && (
                            <>
                                <div className={styles.divider} />
                                <div>
                                    <Icon
                                        aria-label="Private repository"
                                        className={classNames('flex-shrink-0 text-muted', styles.icon)}
                                        svgPath={mdiLock}
                                    />
                                </div>
                                <div>
                                    <small>Private</small>
                                </div>
                            </>
                        )}
                    </div>
                    {result.description && (
                        <>
                            <div className={styles.dividerVertical} />
                            <div>
                                <small>
                                    <em ref={repoDescriptionElement}>
                                        {result.description.length > REPO_DESCRIPTION_CHAR_LIMIT
                                            ? result.description.slice(0, REPO_DESCRIPTION_CHAR_LIMIT) + ' ...'
                                            : result.description}
                                    </em>
                                </small>
                            </div>
                        </>
                    )}
                </div>
            </div>
        </ResultContainer>
    )
}
