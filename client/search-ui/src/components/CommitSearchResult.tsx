import React from 'react'

import VisuallyHidden from '@reach/visually-hidden'
import classNames from 'classnames'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { CommitMatch, getCommitMatchUrl, getRepositoryUrl } from '@sourcegraph/shared/src/search/stream'
// eslint-disable-next-line no-restricted-imports
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Link, Code } from '@sourcegraph/wildcard'

import { CommitSearchResultMatch } from './CommitSearchResultMatch'
import { ResultContainer } from './ResultContainer'

import styles from './SearchResult.module.scss'

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
    const renderTitle = (): JSX.Element => (
        <div className={styles.title}>
            <span
                className={classNames(
                    'test-search-result-label flex-grow-1',
                    styles.titleInner,
                    styles.mutedRepoFileLink
                )}
            >
                <Link to={getRepositoryUrl(result.repository)}>{displayRepoName(result.repository)}</Link>
                <span aria-hidden={true}> ›</span> <Link to={getCommitMatchUrl(result)}>{result.authorName}</Link>
                <span aria-hidden={true}>{': '}</span>
                <Link to={getCommitMatchUrl(result)}>{result.message.split('\n', 1)[0]}</Link>
            </span>
            {/*
                Relative positioning needed needed to avoid VisuallyHidden creating a scrollable overflow in Chrome.
                Related bug: https://bugs.chromium.org/p/chromium/issues/detail?id=1154640#c15
            */}
            <Link to={getCommitMatchUrl(result)} className={classNames('position-relative', styles.titleInner)}>
                <Code className={styles.commitOid}>
                    <VisuallyHidden>Commit hash:</VisuallyHidden>
                    {result.oid.slice(0, 7)}
                    <VisuallyHidden>,</VisuallyHidden>
                </Code>{' '}
                <VisuallyHidden>Commited</VisuallyHidden>
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
            index={index}
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
            as={as}
        />
    )
}
