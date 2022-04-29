import React from 'react'

import classNames from 'classnames'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'

import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { SearchResultStar } from '@sourcegraph/shared/src/components/SearchResultStar'
import { ContentMatch, getFileMatchUrl } from '@sourcegraph/shared/src/search/stream'
import { formatRepositoryStarCount } from '@sourcegraph/shared/src/util/stars'
import { Icon } from '@sourcegraph/wildcard'

import { getIdForLine } from './utils'

interface Props {
    selectResultFromId: (id: string) => void
    selectedResult: null | string
    result: ContentMatch
}

import styles from './FileSearchResult.module.scss'

export const FileSearchResult: React.FunctionComponent<Props> = ({
    result,
    selectedResult,
    selectResultFromId,
}: Props) => {
    const lines = result.lineMatches.map(line => {
        const key = getIdForLine(result, line)
        const onClick = (): void => selectResultFromId(key)

        if (key === selectedResult) {
            console.log(result, line)
        }

        return (
            // The below element's accessibility is handled via a document level event listener.
            //
            // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
            <div
                id={`search-result-list-item-${key}`}
                className={classNames(styles.item, {
                    [styles.itemActive]: key === selectedResult,
                })}
                onMouseDown={preventAll}
                onClick={onClick}
                key={key}
            >
                {line.line} <small>{result.path}</small>
            </div>
        )
    })

    // const renderTitle = (): JSX.Element => {
    //     const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)
    //     return (
    //         <div className={styles.title}>
    //             <RepoIcon repoName={repoName} className="text-muted flex-shrink-0" />
    //             <span
    //                 onMouseEnter={checkTruncation}
    //                 className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate"
    //                 ref={titleReference}
    //                 data-tooltip={
    //                     (truncated && result.type === 'repo' && displayRepoName(getRepoMatchLabel(result))) ||
    //                     (truncated &&
    //                         result.type === 'commit' &&
    //                         `${result.authorName}: ${result.message.split('\n', 1)[0]}`) ||
    //                     null
    //                 }
    //             >
    //                 {result.type === 'commit' && (
    //                     <>
    //                         <Link to={getRepositoryUrl(result.repository)}>{displayRepoName(result.repository)}</Link>
    //                         {' › '}
    //                         <Link to={getCommitMatchUrl(result)}>{result.authorName}</Link>
    //                         {': '}
    //                         <Link to={getCommitMatchUrl(result)}>{result.message.split('\n', 1)[0]}</Link>
    //                     </>
    //                 )}
    //                 {result.type === 'repo' && (
    //                     <Link to={getRepoMatchUrl(result)}>{displayRepoName(getRepoMatchLabel(result))}</Link>
    //                 )}
    //             </span>
    //             <span className={styles.spacer} />
    //             {result.type === 'commit' && (
    //                 <Link to={getCommitMatchUrl(result)}>
    //                     <code className={styles.commitOid}>{result.oid.slice(0, 7)}</code>{' '}
    //                     <Timestamp date={result.authorDate} noAbout={true} strict={true} />
    //                 </Link>
    //             )}
    //             {result.type === 'commit' && formattedRepositoryStarCount && <div className={styles.divider} />}
    //             {formattedRepositoryStarCount && (
    //                 <>
    //                     <SearchResultStar />
    //                     {formattedRepositoryStarCount}
    //                 </>
    //             )}
    //         </div>
    //     )
    // }

    const repoDisplayName = result.repository
    const repoAtRevisionURL = '#'
    const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)

    const title = (
        // eslint-disable-next-line jsx-a11y/no-static-element-interactions
        <div className={styles.header} onMouseDown={preventAll}>
            <Icon className="flex-shrink-0" as={FileDocumentIcon} />
            <div className={classNames('mx-1', styles.headerDivider)} />
            <div className={classNames(styles.headerTitle)} data-testid="result-container-header">
                <RepoIcon repoName={result.repository} className="text-muted flex-shrink-0" />
                <RepoFileLink
                    repoName={result.repository}
                    repoURL={repoAtRevisionURL}
                    filePath={result.path}
                    fileURL={getFileMatchUrl(result)}
                    repoDisplayName={repoDisplayName}
                    className="ml-1 flex-shrink-past-contents text-truncate"
                />
            </div>
            {formattedRepositoryStarCount && (
                <>
                    <SearchResultStar />
                    {formattedRepositoryStarCount}
                </>
            )}
        </div>
    )

    return (
        <>
            {title}
            {lines}
        </>
    )
}

function preventAll(event: React.MouseEvent): void {
    event.stopPropagation()
    event.preventDefault()
}
