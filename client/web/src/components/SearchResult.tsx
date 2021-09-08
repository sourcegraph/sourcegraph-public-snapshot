import classNames from 'classnames'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import LockIcon from 'mdi-react/LockIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'
import StarIcon from 'mdi-react/StarIcon'
import React from 'react'

import { LastSyncedIcon } from '@sourcegraph/shared/src/components/LastSyncedIcon'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { ResultContainer } from '@sourcegraph/shared/src/components/ResultContainer'
import { CommitMatch, getMatchTitle, RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { formatRepositoryStarCount } from '@sourcegraph/shared/src/util/stars'

import { CommitSearchResultMatch } from './CommitSearchResultMatch'
import styles from './SearchResult.module.scss'

interface Props {
    result: CommitMatch | RepositoryMatch
    repoName: string
    icon: React.ComponentType<{ className?: string }>
}

export const SearchResult: React.FunctionComponent<Props> = ({ result, icon, repoName }) => {
    const renderTitle = (): JSX.Element => {
        const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)
        return (
            <div className={styles.title}>
                <RepoIcon repoName={repoName} className="icon-inline text-muted flex-shrink-0" />
                <Markdown
                    className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate"
                    dangerousInnerHTML={renderMarkdown(getMatchTitle(result))}
                />
                <span className={styles.spacer} />
                {result.type === 'commit' && result.detail && (
                    <>
                        <Markdown className="flex-shrink-0" dangerousInnerHTML={renderMarkdown(result.detail)} />
                    </>
                )}
                {result.type === 'commit' && result.detail && formattedRepositoryStarCount && (
                    <div className={styles.divider} />
                )}
                {formattedRepositoryStarCount && (
                    <>
                        <StarIcon className={styles.star} />
                        {formattedRepositoryStarCount}
                    </>
                )}
            </div>
        )
    }

    const renderBody = (): JSX.Element => {
        if (result.type === 'repo') {
            return (
                <div>
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
                                        <SourceForkIcon
                                            className={classNames('icon-inline flex-shrink-0 text-muted', styles.icon)}
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
                                        <ArchiveIcon
                                            className={classNames('icon-inline flex-shrink-0 text-muted', styles.icon)}
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
                                        <LockIcon
                                            className={classNames('icon-inline flex-shrink-0 text-muted', styles.icon)}
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

        return <CommitSearchResultMatch key={result.url} item={result} />
    }

    return (
        <ResultContainer
            icon={icon}
            collapsible={false}
            defaultExpanded={true}
            title={renderTitle()}
            expandedChildren={renderBody()}
        />
    )
}
