import { decode } from 'he'
import { range } from 'lodash'
import _ from 'lodash'
import React from 'react'
import { Link } from 'react-router-dom'
import VisibilitySensor from 'react-visibility-sensor'
import { combineLatest, of, Subject, Subscription } from 'rxjs'
import { filter, switchMap } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { renderMarkdown } from '../discussions/backend'
import { highlightCode } from '../search/backend'
import { highlightNode } from '../util/dom'
import { Markdown } from './Markdown'
import { HighlightRange } from './SearchResult'

interface SearchResultMatchProps {
    item: GQL.ISearchMatch
    body: string
    url: string
    highlightRanges: HighlightRange[]
    isLightTheme: boolean
}

interface SearchResultMatchState {
    HTML?: string
}

export class SearchResultMatch extends React.Component<SearchResultMatchProps, SearchResultMatchState> {
    public state: SearchResultMatchState = {}
    private visibilitySensorOffset = { bottom: -500 }
    private visibilityChanges = new Subject<boolean>()
    private subscriptions = new Subscription()
    private propsChanges = new Subject<SearchResultMatchProps>()

    public constructor(props: SearchResultMatchProps) {
        super(props)
        // Render the match body as markdown, and syntax highlight the response if it's a code block.
        this.subscriptions.add(
            combineLatest(this.propsChanges, this.visibilityChanges)
                .pipe(
                    filter(([, isVisible]) => isVisible),
                    switchMap(([props]) => renderMarkdown({ markdown: props.body }))
                )
                .pipe(
                    switchMap(markdownHTML => {
                        if (this.bodyIsCode() && markdownHTML.includes('<code') && markdownHTML.includes('</code>')) {
                            const lang = this.getLanguage()
                            const codeContent = /<code(?:.*)>([\s\S]*?)<\/code>/.exec(markdownHTML)
                            if (lang && codeContent && codeContent[1]) {
                                return highlightCode({
                                    code: decode(codeContent[1]),
                                    path: 'file.' + lang,
                                    disableTimeout: false,
                                    isLightTheme: this.props.isLightTheme,
                                }).pipe(
                                    switchMap(highlightedStr => {
                                        const highlightedMarkdown = markdownHTML.replace(codeContent[1], highlightedStr)
                                        return of(highlightedMarkdown)
                                    })
                                )
                            }
                        }
                        return of(markdownHTML)
                    })
                )
                .subscribe(str => this.setState({ HTML: str }))
        )
    }

    private tableContainerElement: HTMLElement | null = null

    public componentDidMount(): void {
        this.propsChanges.next(this.props)
        this.highlightNodes()
    }

    public componentDidUpdate(): void {
        this.highlightNodes()
    }

    private highlightNodes(): void {
        // this.splitText()
        if (this.tableContainerElement) {
            const visibleRows = this.tableContainerElement.querySelectorAll('table tr')
            if (visibleRows.length > 0) {
                for (const h of this.props.highlightRanges) {
                    const code = visibleRows[h.line - 1]
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

    private getFirstLine(): number {
        return Math.max(0, Math.min(...this.props.highlightRanges.map(r => r.line)) - 1)
    }

    private getLastLine(): number {
        const lastLine = Math.max(...this.props.highlightRanges.map(r => r.line)) + 1
        return this.props.highlightRanges ? Math.min(lastLine, this.props.highlightRanges.length) : lastLine
    }

    public render(): JSX.Element {
        return (
            <VisibilitySensor
                onChange={this.onChangeVisibility}
                partialVisibility={true}
                offset={this.visibilitySensorOffset}
            >
                <>
                    {this.state.HTML && (
                        <Link key={this.props.url} to={this.props.url} className="file-match__item">
                            <Markdown
                                refFn={this.setTableContainerElement}
                                className="search-result-match code-excerpt"
                                dangerousInnerHTML={this.state.HTML}
                            />
                        </Link>
                    )}
                    {!this.state.HTML && (
                        <table>
                            <tbody>
                                {range(this.getFirstLine(), this.getLastLine()).map(i => (
                                    <tr key={i}>
                                        <td className="line">
                                            <code>{i}</code>
                                        </td>
                                        {/* create empty space to fill viewport (as if the blob content were already fetched, otherwise we'll overfetch) */}
                                        <td className="code"> </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                </>
            </VisibilitySensor>
        )
    }

    private setTableContainerElement = (ref: HTMLElement | null) => {
        this.tableContainerElement = ref
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
