import React from 'react'

import classNames from 'classnames'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'

import { appendSubtreeQueryParameter } from '@sourcegraph/common'
import { CodeHostIcon, SearchResultStar, formatRepositoryStarCount } from '@sourcegraph/search-ui'
import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { ContentMatch, getFileMatchUrl } from '@sourcegraph/shared/src/search/stream'
import { useIsTruncated, Link, Icon } from '@sourcegraph/wildcard'

import { TrimmedCodeLineWithHighlights } from './TrimmedCodeLineWithHighlights'
import { getIdForLine } from './utils'

import styles from './FileSearchResult.module.scss'

interface Props {
    selectResultFromId: (id: string) => void
    selectedResult: null | string
    result: ContentMatch
}

export const FileSearchResult: React.FunctionComponent<Props> = ({
    result,
    selectedResult,
    selectResultFromId,
}: Props) => {
    const lines = result.lineMatches.map(line => {
        const key = getIdForLine(result, line)
        const onClick = (): void => selectResultFromId(key)

        return (
            // The below element's accessibility is handled via a document level event listener.
            //
            // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
            <div
                id={`search-result-list-item-${key}`}
                className={classNames(styles.line, {
                    [styles.lineActive]: key === selectedResult,
                })}
                onClick={onClick}
                key={key}
            >
                <div className={styles.lineCode}>
                    <TrimmedCodeLineWithHighlights line={line} />
                </div>
                <div className={classNames(styles.lineLineNumber, 'text-muted')}>{line.lineNumber + 1}</div>
            </div>
        )
    })

    const repoDisplayName = result.repository
    const repoAtRevisionURL = '#'
    const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)

    const title = (
        <div className={styles.header}>
            <div className={classNames(styles.headerTitle)} data-testid="result-container-header">
                <Icon role="img" aria-label="File" className="flex-shrink-0" as={FileDocumentIcon} />
                <div className={classNames('mx-1', styles.headerDivider)} />
                <CodeHostIcon repoName={result.repository} className="text-muted flex-shrink-0" />
                <UntabableRepoFileLink
                    repoName={result.repository}
                    repoURL={repoAtRevisionURL}
                    filePath={result.path}
                    fileURL={getFileMatchUrl(result)}
                    repoDisplayName={repoDisplayName}
                    className={classNames('ml-1', 'flex-shrink-past-contents', 'text-truncate', styles.headerLink)}
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

/**
 * This is a fork of RepoFileLink with an added tabIndex of -1 so that it's not possible to tab
 * navigate to the individual links (since we want to use manual arrow navigation instead)
 */
interface UntabableRepoFileLinkProps {
    repoName: string
    repoURL: string
    filePath: string
    fileURL: string
    repoDisplayName?: string
    className?: string
}
const UntabableRepoFileLink: React.FunctionComponent<React.PropsWithChildren<UntabableRepoFileLinkProps>> = ({
    repoDisplayName,
    repoName,
    repoURL,
    filePath,
    fileURL,
    className,
}) => {
    const [fileBase, fileName] = splitPath(filePath)
    /**
     * Use the custom hook useIsTruncated to check if overflow: ellipsis is activated for the element
     * We want to do it on mouse enter as browser window size might change after the element has been
     * loaded initially
     */
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    return (
        <div
            ref={titleReference}
            onMouseEnter={checkTruncation}
            className={classNames(className)}
            data-tooltip={truncated ? (fileBase ? `${fileBase}/${fileName}` : fileName) : null}
        >
            <Link tabIndex={-1} to={repoURL}>
                {repoDisplayName || displayRepoName(repoName)}
            </Link>{' '}
            ›{' '}
            <Link tabIndex={-1} to={appendSubtreeQueryParameter(fileURL)}>
                {fileBase ? `${fileBase}/` : null}
                <strong>{fileName}</strong>
            </Link>
        </div>
    )
}
