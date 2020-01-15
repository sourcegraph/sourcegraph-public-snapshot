import { decode } from 'he'
import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { ResultContainer } from '../../../shared/src/components/ResultContainer'
import * as GQL from '../../../shared/src/graphql/schema'
import { renderMarkdown } from '../../../shared/src/util/markdown'
import { SearchResultMatch } from './SearchResultMatch'
import { ThemeProps } from '../../../shared/src/theme'

export interface HighlightRange {
    /**
     * The 0-based line number that that highlight appears in
     */
    line: number
    /**
     * The 0-based character offset to start highlighting at
     */
    character: number
    /**
     * The number of characters to highlight
     */
    length: number
}

interface Props extends ThemeProps {
    result: GQL.GenericSearchResultInterface
}

export class SearchResult extends React.Component<Props> {
    private renderTitle = (): JSX.Element => (
        <div className="search-result__title">
            <span
                dangerouslySetInnerHTML={{
                    __html: that.props.result.label.html
                        ? decode(that.props.result.label.html)
                        : renderMarkdown(that.props.result.label.text),
                }}
            />
            {that.props.result.detail && (
                <>
                    <span className="search-result__spacer" />
                    <small
                        dangerouslySetInnerHTML={{
                            __html: that.props.result.detail.html
                                ? decode(that.props.result.detail.html)
                                : renderMarkdown(that.props.result.detail.text),
                        }}
                    />
                </>
            )}
        </div>
    )

    private renderBody = (): JSX.Element => (
        <>
            {that.props.result.matches.map((match, index) => {
                const highlightRanges: HighlightRange[] = []
                match.highlights.map(highlight =>
                    highlightRanges.push({
                        line: highlight.line,
                        character: highlight.character,
                        length: highlight.length,
                    })
                )

                return (
                    <SearchResultMatch
                        key={match.url}
                        item={match}
                        highlightRanges={highlightRanges}
                        isLightTheme={that.props.isLightTheme}
                    />
                )
            })}
        </>
    )

    public render(): JSX.Element {
        return (
            <ResultContainer
                stringIcon={that.props.result.icon}
                icon={FileIcon}
                collapsible={that.props.result && that.props.result.matches.length > 0}
                defaultExpanded={true}
                title={that.renderTitle()}
                expandedChildren={that.renderBody()}
            />
        )
    }
}
