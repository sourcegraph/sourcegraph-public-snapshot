import React from 'react'

import classNames from 'classnames'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import LockIcon from 'mdi-react/LockIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'

import { LastSyncedIcon } from '@sourcegraph/shared/src/components/LastSyncedIcon'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { ResultContainer } from '@sourcegraph/shared/src/components/ResultContainer'
import { SearchResultStar } from '@sourcegraph/shared/src/components/SearchResultStar'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import {
    CommitMatch,
    getCommitMatchUrl,
    getRepoMatchLabel,
    getRepoMatchUrl,
    getRepositoryUrl,
    RepositoryMatch,
} from '@sourcegraph/shared/src/search/stream'
import { formatRepositoryStarCount } from '@sourcegraph/shared/src/util/stars'
// eslint-disable-next-line no-restricted-imports
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Link, Icon } from '@sourcegraph/wildcard'

import { CommitSearchResultMatch } from './CommitSearchResultMatch'

import styles from './SearchResult.module.scss'

interface Props extends PlatformContextProps<'requestGraphQL'> {
    result: CommitMatch | RepositoryMatch
    repoName: string
    icon: React.ComponentType<{ className?: string }>
    onSelect: () => void
    openInNewTab?: boolean
}

export const SearchResult: React.FunctionComponent<Props> = ({
    result,
    icon,
    repoName,
    platformContext,
    onSelect,
    openInNewTab,
}) => {
    const renderTitle = (): JSX.Element => {
        const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)
        return (
            <div className={styles.title}>
                <RepoIcon repoName={repoName} className="text-muted flex-shrink-0" />
                <span className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate">
                    {result.type === 'commit' && (
                        <>
                            <Link to={getRepositoryUrl(result.repository)}>{displayRepoName(result.repository)}</Link>
                            {' › '}
                            <Link to={getCommitMatchUrl(result)}>{result.authorName}</Link>
                            {': '}
                            <Link to={getCommitMatchUrl(result)}>{result.message.split('\n', 1)[0]}</Link>
                        </>
                    )}
                    {result.type === 'repo' && (
                        <Link to={getRepoMatchUrl(result)}>{displayRepoName(getRepoMatchLabel(result))}</Link>
                    )}
                </span>
                <span className={styles.spacer} />
                {result.type === 'commit' && (
                    <Link to={getCommitMatchUrl(result)}>
                        <code className={styles.commitOid}>{result.oid.slice(0, 7)}</code>{' '}
                        <Timestamp date={result.authorDate} noAbout={true} strict={true} />
                    </Link>
                )}
                {result.type === 'commit' && formattedRepositoryStarCount && <div className={styles.divider} />}
                {formattedRepositoryStarCount && (
                    <>
                        <SearchResultStar />
                        {formattedRepositoryStarCount}
                    </>
                )}
            </div>
        )
    }

    const renderBody = (): JSX.Element => {
        if (result.type === 'repo') {
            return (
                <div data-testid="search-repo-result">
                    <div className={classNames(styles.searchResultMatch, 'p-2 flex-column')}>
                        {result.repoLastFetched && <LastSyncedIcon lastSyncedTime={result.repoLastFetched} />}
                        <div className="d-flex align-items-center flex-row">
                            <div className={styles.matchType}>
                                <small>Repository match</small>
                            </div>
                            {result.fork && (
                                <>
                                    <div className={styles.divider} />
                                    <div>
                                        <Icon
                                            className={classNames('flex-shrink-0 text-muted', styles.icon)}
                                            as={SourceForkIcon}
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
                                            className={classNames('flex-shrink-0 text-muted', styles.icon)}
                                            as={ArchiveIcon}
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
                                            className={classNames('flex-shrink-0 text-muted', styles.icon)}
                                            as={LockIcon}
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
                                        <em>{result.description}</em>
                                    </small>
                                </div>
                            </>
                        )}
                    </div>
                </div>
            )
        }

        return (
            <CommitSearchResultMatch
                key={result.url}
                item={result}
                platformContext={platformContext}
                openInNewTab={openInNewTab}
            />
        )
    }

    return (
        <ResultContainer
            icon={icon}
            collapsible={false}
            defaultExpanded={true}
            title={renderTitle()}
            resultType={result.type}
            onResultClicked={onSelect}
            expandedChildren={renderBody()}
        />
    )
}
