import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { highlightNode } from '../util/dom'
import { ResultContainer } from './ResultContainer'

interface Props {
    result: GQL.IGenericSearchResult
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

    private renderTitle = () => <div dangerouslySetInnerHTML={{ __html: this.props.result.label }} />
    private renderBody = () => (
        <>
            {this.props.result.results!.map(item => {
                const highlightRanges: HighlightRange[] = []
                item.highlights.map(i =>
                    highlightRanges.push({ line: i.line, character: i.character, highlightLength: i.length })
                )
                return <GenCodeExcerpt body={item.body} url={item.url} highlightRanges={highlightRanges} />
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
    body: string
    url: string
    highlightRanges: HighlightRange[]
}

class GenCodeExcerpt extends React.Component<CodeExcerptProps> {
    constructor(props: CodeExcerptProps) {
        super(props)
    }

    private tableContainerElement: HTMLElement | null = null

    public componentDidUpdate(prevProps: CodeExcerptProps): void {
        if (this.tableContainerElement) {
            const visibleRows = this.tableContainerElement.querySelectorAll('table tr')
            for (const highlight of this.props.highlightRanges) {
                console.log(visibleRows)
                const code = visibleRows[0].lastChild as HTMLTableDataCellElement
                highlightNode(code, highlight.character, highlight.highlightLength)
            }
        }
    }

    private getFirstLine(): number {
        return Math.max(0, Math.min(...this.props.highlightRanges.map(r => r.line)) - 1)
    }

    public render(): JSX.Element {
        return (
            <Link to={this.props.url} className="file-match__item">
                <code>
                    <div ref={this.setTableContainerElement} dangerouslySetInnerHTML={{ __html: this.props.body }} />
                </code>
            </Link>
        )
    }
    private setTableContainerElement = (ref: HTMLElement | null) => {
        this.tableContainerElement = ref
    }
}
