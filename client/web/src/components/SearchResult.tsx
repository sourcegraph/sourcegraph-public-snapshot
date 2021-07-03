import ArchiveIcon from 'mdi-react/ArchiveIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'
import StarIcon from 'mdi-react/StarIcon'
import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { ResultContainer } from '@sourcegraph/shared/src/components/ResultContainer'
import { CommitMatch, getMatchTitle, RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { formatRepositoryStarCount } from '@sourcegraph/shared/src/util/stars'

import { CommitSearchResultMatch } from './CommitSearchResultMatch'

interface Props extends ThemeProps {
    result: CommitMatch | RepositoryMatch
    repoName: string
    icon: React.ComponentType<{ className?: string }>
}

export const SearchResult: React.FunctionComponent<Props> = ({ result, icon, isLightTheme, repoName }) => {
    const renderTitle = (): JSX.Element => {
        const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)
        return (
            <div className="search-result__title">
                <RepoIcon repoName={repoName} className="icon-inline text-muted flex-shrink-0" />
                <Markdown
                    className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate"
                    dangerousInnerHTML={renderMarkdown(getMatchTitle(result))}
                />
                <span className="search-result__spacer" />
                {result.type === 'commit' && result.detail && (
                    <>
                        <Markdown className="flex-shrink-0" dangerousInnerHTML={renderMarkdown(result.detail)} />
                    </>
                )}
                {result.type === 'commit' && result.detail && formattedRepositoryStarCount && (
                    <div className="search-result__divider" />
                )}
                {result.type === 'repo' && (
                    <>
                        {result.fork && (
                            <SourceForkIcon
                                className="search-result__icon icon-inline flex-shrink-0 text-muted"
                                data-tooltip="Fork"
                            />
                        )}
                        {result.archived && (
                            <ArchiveIcon
                                className="search-result__icon icon-inline flex-shrink-0 text-muted"
                                data-tooltip="Archive"
                            />
                        )}
                    </>
                )}
                {formattedRepositoryStarCount && (
                    <>
                        <StarIcon className="search-result__star" />
                        {formattedRepositoryStarCount}
                    </>
                )}
            </div>
        )
    }

    const renderBody = (): JSX.Element | undefined => {
        if (result.type === 'repo') {
            return result.description ? (
                <div className="search-result-match p-2 flex-column">
                    <small>
                        <em>{result.description}</em>
                    </small>
                </div>
            ) : undefined
        }

        return <CommitSearchResultMatch key={result.url} item={result} isLightTheme={isLightTheme} />
    }

    return (
        <ResultContainer
            icon={icon}
            // Don't allow collapsing in the redesign
            collapsible={false}
            defaultExpanded={true}
            title={renderTitle()}
            expandedChildren={renderBody()}
        />
    )
}
