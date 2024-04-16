import React from 'react'

import VisuallyHidden from '@reach/visually-hidden'
import classNames from 'classnames'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { type CommitMatch, getCommitMatchUrl, getRepositoryUrl } from '@sourcegraph/shared/src/search/stream'
import { Link, Code } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/Timestamp'

import { CommitSearchResultMatch } from './CommitSearchResultMatch'
import { ResultContainer } from './ResultContainer'

import styles from './CommitSearchResult.module.scss'
import resultStyles from './ResultContainer.module.scss'

interface Props extends PlatformContextProps<'requestGraphQL'> {
    result: CommitMatch
    onSelect: () => void
    openInNewTab?: boolean
    containerClassName?: string
    as?: React.ElementType
    index: number
}

// This is a search result for types diff or commit.
export const CommitSearchResult: React.FunctionComponent<Props> = ({
    result,
    platformContext,
    onSelect,
    openInNewTab,
    containerClassName,
    as,
    index,
}) => {
    const title = (
        <div className={resultStyles.title}>
            <span className={classNames('test-search-result-label flex-grow-1', resultStyles.titleInner)}>
                <Link to={getRepositoryUrl(result.repository)}>{displayRepoName(result.repository)}</Link>
                <span aria-hidden={true}> ›</span> <Link to={getCommitMatchUrl(result)}>{result.authorName}</Link>
                <span aria-hidden={true}>{': '}</span>
                <Link to={getCommitMatchUrl(result)} data-selectable-search-result="true">
                    {result.message.split('\n', 1)[0]}
                </Link>
            </span>
            {/*
                Relative positioning needed needed to avoid VisuallyHidden creating a scrollable overflow in Chrome.
                Related bug: https://bugs.chromium.org/p/chromium/issues/detail?id=1154640#c15
            */}
            <Link to={getCommitMatchUrl(result)} className={classNames('position-relative', resultStyles.titleInner)}>
                <Code className={styles.commitOid}>
                    <VisuallyHidden>Commit hash:</VisuallyHidden>
                    {result.oid.slice(0, 7)}
                    <VisuallyHidden>,</VisuallyHidden>
                </Code>{' '}
                <VisuallyHidden>Committed</VisuallyHidden>
                {/* Display commit date in UTC to match behavior of before/after filters */}
                <Timestamp date={result.committerDate} noAbout={true} strict={true} utc={true} />
            </Link>
            {result.repoStars && <div className={resultStyles.divider} />}
        </div>
    )

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
            repoLastFetched={result.repoLastFetched}
        >
            <CommitSearchResultMatch
                key={result.url}
                item={result}
                platformContext={platformContext}
                openInNewTab={openInNewTab}
            />
        </ResultContainer>
    )
}
