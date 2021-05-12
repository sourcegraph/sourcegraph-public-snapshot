import * as H from 'history'
import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ResultContainer } from '@sourcegraph/shared/src/components/ResultContainer'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { getRepoIcon } from '@sourcegraph/shared/src/util/getRepoIcon'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

import { SearchResultMatch } from './SearchResultMatch'

interface Props extends ThemeProps {
    result: GQL.GenericSearchResultInterface
    history: H.History
    repoName: string
    icon: React.ComponentType<{ className?: string }>
}

export const SearchResult: React.FunctionComponent<Props> = ({ result, history, icon, isLightTheme, repoName }) => {
    const [isRedesignEnabled] = useRedesignToggle()

    const renderTitle = (): JSX.Element => {
        if (isRedesignEnabled && result.__typename === 'Repository') {
            const RepoIcon = getRepoIcon(repoName)
            return (
                <div className="search-result__title">
                    {RepoIcon && <RepoIcon className="icon-inline text-muted" />}
                    <Markdown
                        className="test-search-result-label ml-1"
                        dangerousInnerHTML={result.label.html ? result.label.html : renderMarkdown(result.label.text)}
                    />
                </div>
            )
        }
        return (
            <div className="search-result__title">
                <Markdown
                    className="test-search-result-label"
                    dangerousInnerHTML={result.label.html ? result.label.html : renderMarkdown(result.label.text)}
                />
                {result.detail && (
                    <>
                        <span className="search-result__spacer" />
                        <Markdown
                            dangerousInnerHTML={
                                result.detail.html ? result.detail.html : renderMarkdown(result.detail.text)
                            }
                        />
                    </>
                )}
            </div>
        )
    }

    const renderBody = (): JSX.Element => {
        if (isRedesignEnabled && result.__typename === 'Repository') {
            return (
                <div className="search-result-match p-2">
                    <small>Repository name match</small>
                </div>
            )
        }

        return (
            <>
                {result.matches.map(match => (
                    <SearchResultMatch
                        key={match.url}
                        item={match}
                        highlightRanges={match.highlights}
                        isLightTheme={isLightTheme}
                        history={history}
                    />
                ))}
            </>
        )
    }

    return (
        <ResultContainer
            icon={icon}
            collapsible={result && result.matches.length > 0}
            defaultExpanded={true}
            title={renderTitle()}
            expandedChildren={renderBody()}
        />
    )
}
