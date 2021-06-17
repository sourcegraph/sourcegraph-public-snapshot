import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import SourceForkIcon from 'mdi-react/SourceForkIcon'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import StarIcon from 'mdi-react/StarIcon'
import { ResultContainer } from '@sourcegraph/shared/src/components/ResultContainer'
import { CommitMatch, getMatchTitle, RepositoryMatch } from '@sourcegraph/shared/src/search/stream'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { CommitSearchResultMatch } from './CommitSearchResultMatch'

interface Props extends ThemeProps {
    result: CommitMatch | RepositoryMatch
    repoName: string
    icon: React.ComponentType<{ className?: string }>
}

export const SearchResult: React.FunctionComponent<Props> = ({ result, icon, isLightTheme, repoName }) => {
    const renderTitle = (): JSX.Element => {
        const starDisplayString = starDisplay(result.repoStars)
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
                {result.type === 'commit' && result.detail && starDisplayString && (
                    <div className="search-result__divider"></div>
                )}
                {starDisplayString && (
                    <>
                        <StarIcon className="search-result__star" />
                        {starDisplayString}
                    </>
                )}
            </div>
        )
    }

    const renderBody = (): JSX.Element => {
        if (result.type === 'repo') {
            return (
                <div>
                    <div className="search-result-match p-2 flex-column">
                        <div className="flex-row">
                            <div className="search-result__match-type">
                                <small>Repository match</small>
                            </div>
                            {result.fork && (
                                <>
                                    <div className="search-result__divider"></div>
                                    <div>
                                        <SourceForkIcon className="icon-inline flex-shrink-0 text-muted" />
                                    </div>
                                    <div>
                                        <small>Fork</small>
                                    </div>
                                </>
                            )}
                            {result.archived && (
                                <>
                                    <div className="search-result__divider"></div>
                                    <div>
                                        <ArchiveIcon className="icon-inline flex-shrink-0 text-muted" />
                                    </div>
                                    <div>
                                        <small>Archived</small>
                                    </div>
                                </>
                            )}
                        </div>
                        {result.description && (
                            <>
                                <div className="search-result__divider-vertical"></div>
                                <div className="search-result__description">
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

/**
 * Converts the number of repo stars into a string, formatted nicely for large numbers
 */
const starDisplay = (repoStars?: number): string | undefined => {
    if (repoStars !== undefined) {
        if (repoStars > 1000) {
            return `${Math.floor(repoStars / 1000)}k`
        }
        return repoStars.toString()
    }
    return undefined
}
