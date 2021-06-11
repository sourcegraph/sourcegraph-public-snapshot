import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
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
    const renderTitle = (): JSX.Element => (
        <div className="search-result__title">
            <RepoIcon repoName={repoName} className="icon-inline text-muted flex-shrink-0" />
            <Markdown
                className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate"
                dangerousInnerHTML={renderMarkdown(getMatchTitle(result))}
            />
            {result.type === 'commit' && result.detail && (
                <>
                    <span className="search-result__spacer" />
                    <Markdown className="flex-shrink-0" dangerousInnerHTML={renderMarkdown(result.detail)} />
                </>
            )}
        </div>
    )

    const renderBody = (): JSX.Element => {
        if (result.type === 'repo') {
            return (
                <div className="search-result-match p-2">
                    <small>Repository match</small>
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
