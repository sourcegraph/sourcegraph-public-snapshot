import React, { useCallback } from 'react'

import classNames from 'classnames'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import LockIcon from 'mdi-react/LockIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'

import { CodeHostIcon, formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/search-ui'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { getRepoMatchLabel, RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { Icon, Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import { getResultId } from './utils'

import styles from './SearchResult.module.scss'

export interface RepoSearchResultProps {
    match: RepositoryMatch
    selectedResult: null | string
    selectResult: (id: string) => void
}

export const RepoSearchResult: React.FunctionComponent<RepoSearchResultProps> = ({
    match,
    selectedResult,
    selectResult,
}) => {
    /**
     * Use the custom hook useIsTruncated to check if the content overflows: ellipsis is activated for the element
     * We want to do it on mouse enter as browser window size might change after the element has been
     * loaded initially
     */
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    const resultId = getResultId(match)
    const onClick = useCallback((): void => selectResult(resultId), [selectResult, resultId])

    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)
    return (
        // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions -- this is a clickable element
        <div
            id={`search-result-list-item-${resultId}`}
            className={classNames(styles.line, {
                [styles.lineActive]: resultId === selectedResult,
            })}
            onClick={onClick}
            key={resultId}
        >
            <CodeHostIcon repoName={match.repository} className="text-muted flex-shrink-0" />
            <Tooltip content={(truncated && displayRepoName(getRepoMatchLabel(match))) || null}>
                <span
                    onMouseEnter={checkTruncation}
                    className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate"
                    ref={titleReference}
                >
                    {displayRepoName(getRepoMatchLabel(match))} (Repository match)
                </span>
            </Tooltip>
            {match.fork && (
                <>
                    <div className={styles.divider} />
                    <div>
                        <Icon
                            role="img"
                            aria-label="Forked repository"
                            className={classNames('flex-shrink-0 text-muted', styles.icon)}
                            as={SourceForkIcon}
                        />
                    </div>
                    <div>
                        <small>Fork</small>
                    </div>
                </>
            )}
            {match.archived && (
                <>
                    <div className={styles.divider} />
                    <div>
                        <Icon
                            role="img"
                            aria-label="Archived repository"
                            className={classNames('flex-shrink-0 text-muted', styles.icon)}
                            as={ArchiveIcon}
                        />
                    </div>
                    <div>
                        <small>Archived</small>
                    </div>
                </>
            )}
            {match.private && (
                <>
                    <div className={styles.divider} />
                    <div>
                        <Icon
                            role="img"
                            aria-label="Private repository"
                            className={classNames('flex-shrink-0 text-muted', styles.icon)}
                            as={LockIcon}
                        />
                    </div>
                    <div>
                        <small>Private</small>
                    </div>
                </>
            )}
            <span className={styles.spacer} />
            {formattedRepositoryStarCount && (
                <>
                    <SearchResultStar />
                    {formattedRepositoryStarCount}
                </>
            )}
        </div>
    )
}
