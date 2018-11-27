import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import marked from 'marked'
import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import VisibilitySensor from 'react-visibility-sensor'
import { Subject } from 'rxjs'
import sanitizeHtml from 'sanitize-html'
import * as GQL from '../../../shared/src/graphql/schema'
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
    length: number
}

export class SearchResultMatch extends React.Component<Props> {
    constructor(props: Props) {
        super(props)
    }

    private renderTitle = () => <div dangerouslySetInnerHTML={{ __html: marked(this.props.result.label) }} />

    private renderBody = () => (
        <>
            {this.props.result.results!.map(item => {
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

interface CodeExcerptProps {
    item: GQL.ISearchMatch
    body: string
    url: string
    highlightRanges: HighlightRange[]
    isLightTheme: boolean
}

class MatchExcerpt extends React.Component<CodeExcerptProps> {
    private visibilitySensorOffset = { bottom: -500 }
    private visibilityChanges = new Subject<boolean>()

    public constructor(props: CodeExcerptProps) {
        super(props)
        this.state = { HTMLBody: '' }
    }

    private tableContainerElement: HTMLElement | null = null

    public componentDidMount(): void {
        this.highlightNodes()
    }

    public componentDidUpdate(prevProps: CodeExcerptProps): void {
        this.highlightNodes()
    }

    private highlightNodes(): void {
        this.splitText()
        if (this.tableContainerElement) {
            // Our results are always wrapped in a code element.
            const visibleRows = this.tableContainerElement.querySelectorAll('code')
            if (visibleRows.length > 0) {
                for (const h of this.props.highlightRanges) {
                    const visRows = Array.from(visibleRows[0].childNodes).filter(
                        (node: ChildNode) => node.nodeValue !== '\n'
                    )
                    const code = visRows[h.line - 1]
                    if (code) {
                        highlightNode(code as HTMLElement, h.character, h.length)
                    }
                }
            }
        }
    }

    // Split text splits text nodes. Marked will combine text lines into a single node,
    // which causes our line counts to be off, so we run this to ensure our line counts match.
    private splitText(): void {
        if (this.tableContainerElement) {
            const visibleRows = this.tableContainerElement.querySelectorAll('code')
            if (visibleRows.length > 0) {
                const visRows = Array.from(visibleRows[0].childNodes).filter(
                    (node: ChildNode) => node.nodeValue !== '\n'
                )

                for (const n of visRows) {
                    if (n.nodeName === '#text') {
                        if (n.textContent && n.textContent.indexOf('\n') >= 0) {
                            const node = n as Text
                            const newLineRegex = /\n/g
                            const indices = []
                            let res = newLineRegex.exec(n.textContent.trim())
                            while (res) {
                                indices.push(res.index + 1)
                                res = newLineRegex.exec(n.textContent.trim())
                            }
                            indices.map(i => {
                                try {
                                    node.splitText(i)
                                } catch {
                                    console.error('Index for split text invalid ' + i)
                                }
                            })
                        }
                    }
                }
            }
        }
    }

    private bodyIsCode(): boolean {
        return this.props.body.startsWith('```') && this.props.body.endsWith('```')
    }

    private onChangeVisibility = (isVisible: boolean): void => {
        this.visibilityChanges.next(isVisible)
    }

    private getLanguage(): string | undefined {
        const matches = /(?:```)([^\s]+)\s/.exec(this.props.body)
        if (!matches) {
            return undefined
        }
        return matches[1]
    }

    public render(): JSX.Element {
        return (
            <VisibilitySensor
                onChange={this.onChangeVisibility}
                partialVisibility={true}
                offset={this.visibilitySensorOffset}
            >
                <Link key={this.props.url} to={this.props.url} className="file-match__item">
                    {this.bodyIsCode() ? (
                        <div
                            ref={this.setTableContainerElement}
                            className="search-result-match"
                            dangerouslySetInnerHTML={{
                                __html: '<code>' + this.highlightCodeBlock() + '</code>',
                            }}
                        />
                    ) : (
                        // TODO @attfarhan: Render markdown in the future using marked.
                        <div
                            ref={this.setTableContainerElement}
                            className="search-result-match"
                            dangerouslySetInnerHTML={{
                                __html: '<code>' + splitLines(sanitizeHtml(this.props.body)) + '</code>',
                            }}
                        />
                    )}
                </Link>
            </VisibilitySensor>
        )
    }

    private setTableContainerElement = (ref: HTMLElement | null) => {
        this.tableContainerElement = ref
    }

    private highlightCodeBlock(): string {
        const lang = this.getLanguage()
        if (lang) {
            return highlight(lang!, stripCodeFence(sanitizeHtml(this.props.body)), true).value
        }
        return highlightAuto(stripCodeFence(sanitizeHtml(this.props.body))).value
    }
}

// Strips the code fence from a markdown code block.
function stripCodeFence(code: string): string {
    if (code.startsWith('```') && code.endsWith('```')) {
        const c = code.split('\n')
        return c.slice(1, c.length - 1).join('\n')
    }
    return code
}

// Split lines separates markdown text lines into individual elements so that we can treat each
// line individually for match highlighting.
function splitLines(body: string): string {
    const split = body.split('\n')
    let htmlAsString = ''
    for (const s of split) {
        const sp = `<span>${s}\n</span>`
        htmlAsString += sp
    }
    return htmlAsString
}
