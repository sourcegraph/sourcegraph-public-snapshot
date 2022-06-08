import React from 'react'

import SourceCommitIcon from 'mdi-react/SourceCommitIcon'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { CommitMatch, getCommitMatchUrl, getRepositoryUrl } from '@sourcegraph/shared/src/search/stream'
// eslint-disable-next-line no-restricted-imports
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Link, Code, useIsTruncated, Tooltip } from '@sourcegraph/wildcard'

import { CommitSearchResultMatch } from './CommitSearchResultMatch'
import { ResultContainer } from './ResultContainer'

import styles from './SearchResult.module.scss'

interface Props extends PlatformContextProps<'requestGraphQL'> {
    result: CommitMatch
    onSelect: () => void
    openInNewTab?: boolean
    containerClassName?: string
}

// This is a search result for types diff or commit.
export const CommitSearchResult: React.FunctionComponent<Props> = ({
    result,
    platformContext,
    onSelect,
    openInNewTab,
    containerClassName,
}) => {
    /**
     * Use the custom hook useIsTruncated to check if overflow: ellipsis is activated for the element
     * We want to do it on mouse enter as browser window size might change after the element has been
     * loaded initially
     */
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    const renderTitle = (): JSX.Element => (
        <div className={styles.title}>
            <Tooltip content={(truncated && `${result.authorName}: ${result.message.split('\n', 1)[0]}`) || null}>
                <span
                    onMouseEnter={checkTruncation}
                    className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate"
                    ref={titleReference}
                >
                    <>
                        <Link to={getRepositoryUrl(result.repository)}>{displayRepoName(result.repository)}</Link>
                        {' › '}
                        <Link to={getCommitMatchUrl(result)}>{result.authorName}</Link>
                        {': '}
                        <Link to={getCommitMatchUrl(result)}>{result.message.split('\n', 1)[0]}</Link>
                    </>
                </span>
            </Tooltip>
            <span className={styles.spacer} />
            <Link to={getCommitMatchUrl(result)}>
                <Code className={styles.commitOid}>{result.oid.slice(0, 7)}</Code>{' '}
                <Timestamp date={result.authorDate} noAbout={true} strict={true} />
            </Link>
            {result.repoStars && <div className={styles.divider} />}
        </div>
    )

    const renderBody = (): JSX.Element => (
        <CommitSearchResultMatch
            key={result.url}
            item={result}
            platformContext={platformContext}
            openInNewTab={openInNewTab}
        />
    )

    return (
        <ResultContainer
            icon={SourceCommitIcon}
            collapsible={false}
            defaultExpanded={true}
            title={renderTitle()}
            resultType={result.type}
            onResultClicked={onSelect}
            expandedChildren={renderBody()}
            repoName={result.repository}
            repoStars={result.repoStars}
            className={containerClassName}
        />
    )
}
