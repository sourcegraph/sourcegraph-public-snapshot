import React from 'react'

import classNames from 'classnames'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import LockIcon from 'mdi-react/LockIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

// eslint-disable-next-line no-restricted-imports
import styles from '@sourcegraph/search-ui/src/components/SearchResult.module.scss'
import { LastSyncedIcon } from '@sourcegraph/shared/src/components/LastSyncedIcon'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { ResultContainer } from '@sourcegraph/shared/src/components/ResultContainer'
import { SearchResultStar } from '@sourcegraph/shared/src/components/SearchResultStar'
import { getRepoMatchLabel, RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { formatRepositoryStarCount } from '@sourcegraph/shared/src/util/stars'
import { Icon } from '@sourcegraph/wildcard'

import { useOpenSearchResultsContext } from '../MatchHandlersContext'
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
    const { openRepo } = useOpenSearchResultsContext()

    const renderTitle = (): JSX.Element => {
        const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)
        return (
            <div className={styles.title}>
                <RepoIcon repoName={repoName} className="text-muted flex-shrink-0" />
                <span className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate">
                    <button type="button" className="btn btn-text-link" onClick={() => openRepo(result)}>
                        {displayRepoName(getRepoMatchLabel(result))}
                    </button>
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
                                <Icon className={classNames('flex-shrink-0 text-muted', styles.icon)} as={LockIcon} />
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
