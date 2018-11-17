import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import marked from 'marked'
import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import VisibilitySensor from 'react-visibility-sensor'
import { Subject } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { highlightNode } from '../util/dom'
import { ResultContainer } from './ResultContainer'

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

export class GenericMatch extends React.Component<Props> {
    constructor(props: Props) {
        super(props)
    }

    private renderTitle = () => <div dangerouslySetInnerHTML={{ __html: marked(this.props.result.label) }} />

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
                title={this.renderTitle()}
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

class GenCodeExcerpt extends React.Component<CodeExcerptProps> {
    private visibilitySensorOffset = { bottom: -500 }
    private visibilityChanges = new Subject<boolean>()

    public constructor(props: CodeExcerptProps) {
        super(props)
        this.state = { HTMLBody: '' }
    }

    private tableContainerElement: HTMLElement | null = null

    public componentDidUpdate(prevProps: CodeExcerptProps): void {
        if (this.tableContainerElement) {
            const visibleRows = this.tableContainerElement.querySelectorAll('table tr')
            if (visibleRows.length > 0) {
                for (const h of this.props.highlightRanges) {
                    // If we add context lines we must select the right line
                    const code = visibleRows[0].lastChild as HTMLTableDataCellElement
                    highlightNode(code, h.character, h.highlightLength)
                }
            }
        }
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
                for (const h of this.props.highlightRanges) {
                    // If we add context lines we must select the right line
                    const code = visibleRows[0].lastChild as HTMLTableDataCellElement
                    highlightNode(code, h.character, h.highlightLength)
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
                    <div
                        ref={this.setTableContainerElement}
                        dangerouslySetInnerHTML={{
                            __html: marked(this.props.body, highlightfn),
                        }}
                    />
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

const highlightfn = { highlight: (code: string) => highlightAuto(code).value }
