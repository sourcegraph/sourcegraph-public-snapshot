import React from 'react'

import classNames from 'classnames'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import LockIcon from 'mdi-react/LockIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'

import { CodeHostIcon, formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/search-ui'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { getRepoMatchLabel, RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { Icon, Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import { SelectableSearchResult } from './SelectableSearchResult'

import styles from './RepoSearchResult.module.scss'

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

    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)
    return (
        <SelectableSearchResult match={match} selectResult={selectResult} selectedResult={selectedResult}>
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
        </SelectableSearchResult>
    )
}
