import React from 'react'

import {
    formatRepositoryStarCount,
    SearchResultStyles as styles,
    CodeHostIcon,
    ResultContainer,
    SearchResultStar,
    CommitSearchResultMatch,
} from '@sourcegraph/search-ui'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { CommitMatch, getCommitMatchUrl } from '@sourcegraph/shared/src/search/stream'
// eslint-disable-next-line no-restricted-imports
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Typography } from '@sourcegraph/wildcard'

import { useOpenSearchResultsContext } from '../MatchHandlersContext'
interface Props extends PlatformContextProps<'requestGraphQL'> {
    result: CommitMatch
    repoName: string
    icon: React.ComponentType<{ className?: string }>
    onSelect: () => void
    openInNewTab?: boolean
    containerClassName?: string
}

export const CommitSearchResult: React.FunctionComponent<Props> = ({
    result,
    icon,
    repoName,
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
    const { openRepo, openCommit, instanceURL } = useOpenSearchResultsContext()

    const renderTitle = (): JSX.Element => {
        const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)
        return (
            <div className={styles.title}>
                <CodeHostIcon repoName={repoName} className="text-muted flex-shrink-0" />
                <span className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate">
                    <>
                        <button
                            type="button"
                            className="btn btn-text-link"
                            onClick={() =>
                                openRepo({
                                    repository: result.repository,
                                    branches: [result.oid],
                                })
                            }
                        >
                            {displayRepoName(result.repository)}
                        </button>
                        {' › '}
                        <button
                            type="button"
                            className="btn btn-text-link"
                            onClick={() => openCommit(getCommitMatchUrl(result))}
                        >
                            {result.authorName}
                        </button>
                        {': '}
                        <button
                            type="button"
                            className="btn btn-text-link"
                            onClick={() => openCommit(getCommitMatchUrl(result))}
                        >
                            {result.message.split('\n', 1)[0]}
                        </button>
                    </>
                </span>
                <span className={styles.spacer} />
                {result.type === 'commit' && (
                    <button
                        type="button"
                        className="btn btn-text-link"
                        onClick={() => openCommit(getCommitMatchUrl(result))}
                    >
                        <Typography.Code className={styles.commitOid}>{result.oid.slice(0, 7)}</Typography.Code>{' '}
                        <Timestamp date={result.authorDate} noAbout={true} strict={true} />
                    </button>
                )}
                {formattedRepositoryStarCount && (
                    <>
                        <div className={styles.divider} />
                        <SearchResultStar />
                        {formattedRepositoryStarCount}
                    </>
                )}
            </div>
        )
    }

    const renderBody = (): JSX.Element => (
        <CommitSearchResultMatch
            key={result.url}
            item={{
                ...result,
                // Make it an absolute URL to open in browser.
                url: new URL(result.url, instanceURL).href,
            }}
            platformContext={platformContext}
            openInNewTab={openInNewTab}
        />
    )

    return (
        <ResultContainer
            icon={icon}
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
