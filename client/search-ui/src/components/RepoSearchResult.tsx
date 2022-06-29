import React from 'react'

import { mdiSourceFork, mdiArchive, mdiLock } from '@mdi/js'
import classNames from 'classnames'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { getRepoMatchLabel, getRepoMatchUrl, RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { Icon, Link, Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import { LastSyncedIcon } from './LastSyncedIcon'
import { ResultContainer } from './ResultContainer'

import styles from './SearchResult.module.scss'

export interface RepoSearchResultProps {
    result: RepositoryMatch
    onSelect: () => void
    containerClassName?: string
    as?: React.ElementType
    index: number
}

export const RepoSearchResult: React.FunctionComponent<RepoSearchResultProps> = ({
    result,
    onSelect,
    containerClassName,
    as,
    index,
}) => {
    /**
     * Use the custom hook useIsTruncated to check if overflow: ellipsis is activated for the element
     * We want to do it on mouse enter as browser window size might change after the element has been
     * loaded initially
     */
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    const renderTitle = (): JSX.Element => (
        <div className={styles.title}>
            <Tooltip content={(truncated && displayRepoName(getRepoMatchLabel(result))) || null} placement="bottom">
                <span
                    onMouseEnter={checkTruncation}
                    className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate"
                    ref={titleReference}
                >
                    <Link to={getRepoMatchUrl(result)}>{displayRepoName(getRepoMatchLabel(result))}</Link>
                </span>
            </Tooltip>
        </div>
    )

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
                                    aria-label="Forked repository"
                                    className={classNames('flex-shrink-0 text-muted', styles.icon)}
                                    svgPath={mdiSourceFork}
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
                                    aria-label="Archived repository"
                                    className={classNames('flex-shrink-0 text-muted', styles.icon)}
                                    svgPath={mdiArchive}
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
                                    aria-label="Private repository"
                                    className={classNames('flex-shrink-0 text-muted', styles.icon)}
                                    svgPath={mdiLock}
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
            index={index}
            icon={SourceRepositoryIcon}
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
