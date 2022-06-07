import React from 'react'

import { CodeHostIcon, formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/search-ui'
import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { PathMatch } from '@sourcegraph/shared/src/search/stream'
import { Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import { SelectableSearchResult } from './SelectableSearchResult'

import styles from './PathSearchResult.module.scss'

interface Props {
    match: PathMatch
    selectedResult: null | string
    selectResult: (id: string) => void
}

export const PathSearchResult: React.FunctionComponent<Props> = ({ match, selectedResult, selectResult }: Props) => {
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)

    const [fileBase, fileName] = splitPath(match.path)

    return (
        <SelectableSearchResult match={match} selectResult={selectResult} selectedResult={selectedResult}>
            <CodeHostIcon repoName={match.repository} className="text-muted flex-shrink-0" />
            <Tooltip content={truncated ? (fileBase ? `${fileBase}/${fileName}` : fileName) : null}>
                <div ref={titleReference} onMouseEnter={checkTruncation}>
                    {displayRepoName(match.repository)} › {fileBase ? `${fileBase}/` : null}
                    <strong>{fileName}</strong>
                </div>
            </Tooltip>
            <span className={styles.spacer} />
            {formattedRepositoryStarCount && (
                <>
                    <div className={styles.divider} />
                    <SearchResultStar />
                    {formattedRepositoryStarCount}
                </>
            )}
        </SelectableSearchResult>
    )
}
