import React from 'react'

import classNames from 'classnames'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import LockIcon from 'mdi-react/LockIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { getRepoMatchLabel, getRepoMatchUrl, RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { Icon, Link, useIsTruncated } from '@sourcegraph/wildcard'

import { formatRepositoryStarCount } from '../util/stars'

import { CodeHostIcon } from './CodeHostIcon'
import { LastSyncedIcon } from './LastSyncedIcon'
import { ResultContainer } from './ResultContainer'
import { SearchResultStar } from './SearchResultStar'

import styles from './SearchResult.module.scss'

export interface RepoSearchResultProps {
    result: RepositoryMatch
    repoName: string
    onSelect: () => void
    containerClassName?: string
}

export const RepoSearchResult: React.FunctionComponent<RepoSearchResultProps> = ({
    result,
    repoName,
    onSelect,
    containerClassName,
}) => {
    /**
     * Use the custom hook useIsTruncated to check if overflow: ellipsis is activated for the element
     * We want to do it on mouse enter as browser window size might change after the element has been
     * loaded initially
     */
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    const renderTitle = (): JSX.Element => {
        const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)
        return (
            <div className={styles.title}>
                <CodeHostIcon repoName={repoName} className="text-muted flex-shrink-0" />
                <span
                    onMouseEnter={checkTruncation}
                    className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate"
                    ref={titleReference}
                    data-tooltip={(truncated && displayRepoName(getRepoMatchLabel(result))) || null}
                >
                    <Link to={getRepoMatchUrl(result)}>{displayRepoName(getRepoMatchLabel(result))}</Link>
                </span>
                <span className={styles.spacer} />
                {formattedRepositoryStarCount && (
                    <>
                        <SearchResultStar />
                        {formattedRepositoryStarCount}
                    </>
                )}
            </div>
        )
    }

    const renderBody = (): JSX.Element => (
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
                                    role="img"
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
                    {result.archived && (
                        <>
                            <div className={styles.divider} />
                            <div>
                                <Icon
                                    role="img"
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
                    {result.private && (
                        <>
                            <div className={styles.divider} />
                            <div>
                                <Icon
                                    role="img"
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

    return (
        <ResultContainer
            icon={SourceRepositoryIcon}
            collapsible={false}
            defaultExpanded={true}
            title={renderTitle()}
            resultType={result.type}
            onResultClicked={onSelect}
            expandedChildren={renderBody()}
            className={containerClassName}
        />
    )
}
