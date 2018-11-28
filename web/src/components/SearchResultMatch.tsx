import marked from 'marked'
import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { MatchExcerpt } from './MatchExcerpt'
import { ResultContainer } from './ResultContainer'

export interface HighlightRange {
    /**
     * The 0-based line number that this highlight appears in
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

interface Props {
    result: GQL.GenericSearchResult
    isLightTheme: boolean
}

export class SearchResultMatch extends React.Component<Props> {
    constructor(props: Props) {
        super(props)
    }

    private renderTitle = () => (
        <div className="repository-search-result__title">
            <span
                dangerouslySetInnerHTML={{
                    __html: marked(this.props.result.label, { gfm: true, breaks: true, sanitize: true }),
                }}
            />
            {this.props.result.detail && (
                <>
                    <span className="repository-search-result__spacer" />
                    <small
                        dangerouslySetInnerHTML={{
                            __html: marked(this.props.result.detail, { gfm: true, breaks: true, sanitize: true }),
                        }}
                    />
                </>
            )}
        </div>
    )

    private renderBody = () => (
        <>
            {this.props.result.results.map(item => {
                const highlightRanges: HighlightRange[] = []
                item.highlights.map(i =>
                    highlightRanges.push({ line: i.line, character: i.character, length: i.length })
                )

                return (
                    <MatchExcerpt
                        key={item.url}
                        item={item}
                        body={item.body}
                        url={item.url}
                        highlightRanges={highlightRanges}
                        isLightTheme={this.props.isLightTheme}
                    />
                )
            })}
        </>
    )

    public render(): JSX.Element {
        return (
            <ResultContainer
                stringIcon={this.props.result.icon}
                icon={FileIcon}
                title={this.renderTitle()}
                expandedChildren={this.renderBody()}
                collapsedChildren={this.renderBody()}
            />
        )
    }
}
