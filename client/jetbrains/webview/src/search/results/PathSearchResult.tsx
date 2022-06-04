import React, { useCallback } from 'react'

import classNames from 'classnames'

import { CodeHostIcon, formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/search-ui'
import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { PathMatch } from '@sourcegraph/shared/src/search/stream'
import { Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import { getResultId } from './utils'

import styles from './SearchResult.module.scss'

interface Props {
    match: PathMatch
    selectedResult: null | string
    selectResult: (id: string) => void
}

export const PathSearchResult: React.FunctionComponent<Props> = ({ match, selectedResult, selectResult }: Props) => {
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)

    const resultId = getResultId(match)
    const onClick = useCallback((): void => selectResult(resultId), [selectResult, resultId])

    const [fileBase, fileName] = splitPath(match.path)

    return (
        // The below element's accessibility is handled via a document level event listener.
        //
        // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
        <div
            id={`search-result-list-item-${resultId}`}
            className={classNames(styles.line, {
                [styles.lineActive]: resultId === selectedResult,
            })}
            onClick={onClick}
            key={resultId}
        >
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
        </div>
    )
}
