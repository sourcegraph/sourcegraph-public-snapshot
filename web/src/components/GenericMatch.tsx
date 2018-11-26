import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import marked from 'marked'
import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import VisibilitySensor from 'react-visibility-sensor'
import { Subject } from 'rxjs'
import * as GQL from '../../../shared/src/graphql/schema'
import { highlightNode } from '../util/dom'
import { DecoratedTextLines } from './DecoratedTextLines'
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

export class GenericMatch extends React.Component<Props> {
    constructor(props: Props) {
        super(props)
    }

    private renderTitle = () => <div dangerouslySetInnerHTML={{ __html: marked(this.props.result.label) }} />

    private renderBody = () => (
        <>
            {this.props.result &&
                this.props.result.results &&
                this.props.result.results.map(item => {
                    const highlightRanges: HighlightRange[] = []
                    if (item.highlights) {
                        item.highlights.map(i =>
                            highlightRanges.push({ line: i.line, character: i.character, length: i.length })
                        )
                    }

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
                    highlightNode(code, h.character, h.length)
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

    public render(): JSX.Element {
        if (this.tableContainerElement) {
            // Our results are always wrapped in a code element.
            const visibleRows = this.tableContainerElement.querySelectorAll('code')
            // for commits
            // const visibleRows = this.tableContainerElement.querySelectorAll('p')
            // For code elements, content is wrapped in <pre> and <code> tags, if not, we wrap a content with a <div>.
            // const visibleRows = this.tableContainerElement.firstChild!.childNodes
            // console.log(visibleRows)
            if (visibleRows.length > 0) {
                // console.log(
                //     'vis',
                //     Array.from(visibleRows[0].childNodes).filter((node: ChildNode) => node.nodeValue !== '\n')
                // )
                for (const h of this.props.highlightRanges) {
                    // console.log(h)
                    // for diffs
                    const visRows = Array.from(visibleRows[0].childNodes).filter(
                        (node: ChildNode) => node.nodeValue !== '\n'
                    )
                    // for commits
                    // console.log('visibleROWS', visRows)
                    // const visRows = visibleRows[h.line - 1]
                    // console.log('FIRST ITEM IN VISIBLEROWS', visRows)
                    // console.log(visRows[h.line - 1])
                    const code = visRows[h.line - 1]
                    // const byLines = code.split('\n')
                    // const code = visibleRows[h.line - 1].lastChild as HTMLTableDataCellElement
                    if (code) {
                        highlightNode(code as HTMLElement, h.character, h.length)
                    }
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
                    {this.bodyIsCode() ? (
                        <div
                            ref={this.setTableContainerElement}
                            dangerouslySetInnerHTML={{
                                // Heuristic: replace 4 spaces with tabs, otherwise character counts get thrown off.
                                // Marked does not preserve tabs, so we get wrong spacing for results where white-space
                                // is actually spaces. TODO @attfarhan: we could read the language of the code block.
                                // or the file that the diff result comes from to optimize this.
                                __html: marked(this.props.body, this.highlightfn).replace(/\s{4}/g, '\t'),
                            }}
                        />
                    ) : (
                        <div
                            ref={this.setTableContainerElement}
                            dangerouslySetInnerHTML={{
                                __html: '<code>' + splitLines(this.props.body) + '</code>',
                            }}
                        />
                    )}
                </Link>
            </VisibilitySensor>
        )
    }

    private getLanguage = () => {
        const matches = /(?:```)([^\s]+)\s/.exec(this.props.body)
        if (!matches) {
            return null
        }
        return matches[1]
    }

    private setTableContainerElement = (ref: HTMLElement | null) => {
        this.tableContainerElement = ref
    }

    private highlightfn = {
        sanitize: true,
        highlight: (code: string) => {
            const lang = this.getLanguage()
            return lang ? highlight(lang, code, true).value : highlightAuto(code).value
        },
    }
}

// Split lines separates markdown text lines into individual elements so that we can treat each
// line individually for match highlighting.
function splitLines(body: string): string {
    console.log('body', body)
    const split = body.split('\n')
    let htmlAsString = ''
    for (const s of split) {
        const sp = `<span>${s}\n</span>`
        htmlAsString += sp
    }

    return htmlAsString
}
