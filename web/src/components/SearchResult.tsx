import marked from 'marked'
import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { Subject, Subscription, of } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { renderMarkdowns } from '../discussions/backend'
import { ResultContainer } from './ResultContainer'
import { SearchResultMatch } from './SearchResultMatch'

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

interface State {
    matchesRenderedMarkdown?: string[]
}

export class SearchResult extends React.Component<Props, State> {
    public state: State = {}
    private subscriptions = new Subscription()
    private propsChanges = new Subject<Props>()

    constructor(props: Props) {
        super(props)
        this.subscriptions.add(
            this.propsChanges
                .pipe(
                    switchMap(props => {
                        const markdownsToRender: string[] = []
                        props.result.matches.map(matches => markdownsToRender.push(matches.body))
                        return renderMarkdowns(markdownsToRender).pipe(
                            switchMap(renderedMarkdowns => of(renderedMarkdowns))
                        )
                    })
                )
                .subscribe(htmlList => this.setState({ matchesRenderedMarkdown: htmlList }))
        )
    }

    public componentDidMount(): void {
        this.propsChanges.next(this.props)
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
            {!!this.state.matchesRenderedMarkdown &&
                this.props.result.matches.map((item, index) => {
                    const highlightRanges: HighlightRange[] = []
                    item.highlights.map(highlight =>
                        highlightRanges.push({
                            line: highlight.line,
                            character: highlight.character,
                            length: highlight.length,
                        })
                    )

                    return (
                        <SearchResultMatch
                            key={item.url}
                            item={item}
                            body={this.state.matchesRenderedMarkdown[index]}
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
                collapsible={this.props.result && this.props.result.matches.length > 0}
                defaultExpanded={true}
                title={this.renderTitle()}
                expandedChildren={this.renderBody()}
            />
        )
    }
}
