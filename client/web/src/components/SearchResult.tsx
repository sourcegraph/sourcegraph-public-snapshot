import * as H from 'history'
import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { ResultContainer } from '@sourcegraph/shared/src/components/ResultContainer'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
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
        if (isRedesignEnabled) {
            return (
                <div className="search-result__title">
                    <RepoIcon repoName={repoName} className="icon-inline text-muted flex-shrink-0" />
                    <Markdown
                        className="test-search-result-label ml-1 flex-shrink-past-contents text-truncate"
                        dangerousInnerHTML={result.label.html ? result.label.html : renderMarkdown(result.label.text)}
                    />
                    {result.__typename !== 'Repository' && result.detail && (
                        <>
                            <span className="search-result__spacer" />
                            <Markdown
                                className="flex-shrink-0"
                                dangerousInnerHTML={
                                    result.detail.html ? result.detail.html : renderMarkdown(result.detail.text)
                                }
                            />
                        </>
                    )}
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
            // Don't allow collapsing in the redesign
            collapsible={!isRedesignEnabled && result && result.matches.length > 0}
            defaultExpanded={true}
            title={renderTitle()}
            expandedChildren={renderBody()}
        />
    )
}
