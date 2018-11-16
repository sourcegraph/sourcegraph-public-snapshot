import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import VisibilitySensor from 'react-visibility-sensor'
import { Subject, Subscription, from, of } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { highlightNode } from '../util/dom'
import { ResultContainer } from './ResultContainer'
import { renderMarkdown } from '../discussions/backend'
import { map, switchMap } from 'rxjs/operators'

interface Props {
    result: GQL.GenericSearchResult
    isLightTheme: boolean
}

interface HighlightRange {
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
    highlightLength: number
}

interface State {
    HTMLLabel: React.ReactFragment
}
export class GenericMatch extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = { HTMLLabel: '' }
    }

    private subscriptions = new Subscription()
    private propsChanges = new Subject<Props>()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.propsChanges
                .pipe(switchMap(props => renderMarkdown(props.result.label).pipe(str => str)))
                .subscribe(str => this.setState({ HTMLLabel: <div dangerouslySetInnerHTML={{ __html: str }} /> }))
        )

        this.propsChanges.next(this.props)
    }

    private renderBody = () => (
        <>
            {this.props.result.results!.map(item => {
                const highlightRanges: HighlightRange[] = []
                item.highlights.map(i =>
                    highlightRanges.push({ line: i.line, character: i.character, highlightLength: i.length })
                )

                return (
                    <GenCodeExcerpt
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
                title={this.state.HTMLLabel}
                expandedChildren={this.renderBody()}
                collapsedChildren={this.renderBody()}
            />
        )
    }
}

interface CodeExcerptProps {
    item: GQL.IGenericSearchMatch
    body: string
    url: string
    highlightRanges: HighlightRange[]
    isLightTheme: boolean
}

interface CodeExcerptState {
    HTMLBody: string
}

class GenCodeExcerpt extends React.Component<CodeExcerptProps, CodeExcerptState> {
    private visibilitySensorOffset = { bottom: -500 }
    private visibilityChanges = new Subject<boolean>()

    private subscriptions = new Subscription()
    private propsChanges = new Subject<CodeExcerptProps>()

    public constructor(props: CodeExcerptProps) {
        super(props)
        this.state = { HTMLBody: '' }
    }

    private tableContainerElement: HTMLElement | null = null

    public componentDidMount(): void {
        this.subscriptions.add(
            this.propsChanges
                .pipe(switchMap(props => renderMarkdown(props.body).pipe(str => str)))
                .subscribe(str => this.setState({ HTMLBody: str }))
        )

        this.propsChanges.next(this.props)
    }

    public componentDidUpdate(prevProps: CodeExcerptProps): void {
        if (this.tableContainerElement) {
            const visibleRows = this.tableContainerElement.querySelectorAll('table tr')
            if (visibleRows.length > 0) {
                for (const highlight of this.props.highlightRanges) {
                    // If we add context lines we must select the right line
                    const code = visibleRows[0].lastChild as HTMLTableDataCellElement
                    highlightNode(code, highlight.character, highlight.highlightLength)
                }
            }
        }

        // this.propsChanges.next(this.props)
    }

    private onChangeVisibility = (isVisible: boolean): void => {
        this.visibilityChanges.next(isVisible)
    }

    // private getFirstLine(): number {
    //     return Math.max(0, Math.min(...this.props.highlightRanges.map(r => r.line)) - 1)
    // }

    public render(): JSX.Element {
        if (this.tableContainerElement) {
            const visibleRows = this.tableContainerElement.querySelectorAll('table tr')
            if (visibleRows.length > 0) {
                for (const highlight of this.props.highlightRanges) {
                    // If we add context lines we must select the right line
                    const code = visibleRows[0].lastChild as HTMLTableDataCellElement
                    highlightNode(code, highlight.character, highlight.highlightLength)
                }
            }
        }
        return (
            <VisibilitySensor
                onChange={this.onChangeVisibility}
                partialVisibility={true}
                offset={this.visibilitySensorOffset}
            >
                <Link key={this.props.url} to={this.props.url} className="file-match__item">
                    <code>
                        <div
                            ref={this.setTableContainerElement}
                            dangerouslySetInnerHTML={{
                                __html: this.state.HTMLBody,
                            }}
                        />
                    </code>
                </Link>
            </VisibilitySensor>
        )
    }
    private setTableContainerElement = (ref: HTMLElement | null) => {
        this.tableContainerElement = ref
    }

    private makeTableHTML(): string {
        return '<table><tr><td>' + this.props.body + '</td></tr></table>'
    }
}
