import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import * as H from 'history'
import { isEqual } from 'lodash'
import marked from 'marked'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { forkJoin, of, Subject, Subscription } from 'rxjs'
import { catchError, debounceTime, distinctUntilChanged, startWith, switchMap } from 'rxjs/operators'
import { MarkedString, MarkupContent, MarkupKind } from 'vscode-languageserver-types'
import { parseBrowserRepoURL } from '.'
import { getHover, HoverMerged } from '../backend/features'
import { EMODENOTFOUND } from '../backend/lsp'
import { displayRepoPath } from '../components/RepoFileLink'
import { CXPControllerProps } from '../cxp/CXPEnvironment'
import { getModeFromPath } from '../util'
import { asError, ErrorLike } from '../util/errors'
import { fetchHighlightedFileLines } from './backend'

const highlightCodeSafe = (code: string, language?: string): string => {
    try {
        if (language) {
            return highlight(language, code, true).value
        }
        return highlightAuto(code).value
    } catch (err) {
        console.warn('Error syntax-highlighting hover markdown code block', err)
        return escape(code)
    }
}

const hoverContentsToString = (contents: (MarkupContent | MarkedString)[]): string[] => {
    const contentList = []
    const hoverContents = contents
    for (let content of hoverContents) {
        let sig
        if (typeof content === 'string') {
            const hold = content
            content = { kind: MarkupKind.Markdown, value: hold }
        }
        if (MarkupContent.is(content)) {
            if (content.kind === MarkupKind.Markdown) {
                try {
                    const rendered = marked(content.value, {
                        gfm: true,
                        breaks: true,
                        sanitize: true,
                        highlight: (code, language) => '<code>' + highlightCodeSafe(code, language) + '</code>',
                    })
                    sig = rendered
                } catch (err) {
                    sig = 'errored'
                }
            } else {
                sig = content.value
            }
        } else {
            sig = highlightCodeSafe(content.value, content.language)
        }
        contentList.push(sig)
    }

    return contentList
}

const getSymbolSignature = (contents: (MarkupContent | MarkedString)[]): string => {
    const symbolSignature = [contents[0]]
    return hoverContentsToString(symbolSignature)[0]
}

interface HistoryProps extends CXPControllerProps {
    location: H.Location
    history: H.History
    commitID: string
}

/** The data stored for each history entry. */
interface SymbolHistoryEntry {
    /** The symbol name */
    name: string
    repoPath: string
    filePath: string
    url: string
    lineNumber?: number
    /** Hover contents, excluding the symbol name */
    hoverContents: string[]
    /** Combination of name, hover contents, lines of code. */
    rawString: string
}

interface HistoryState {
    hoverOrError?: HoverMerged | ErrorLike | null
    fileLinesOrError?: string[]
    historyListToRender: SymbolHistoryEntry[]
    filter?: string
}

export class RepoRevSidebarHistory extends React.Component<HistoryProps, HistoryState> {
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<HistoryProps>()
    private filterQueryUpdates = new Subject<string>()
    public state: HistoryState = { historyListToRender: JSON.parse(localStorage.getItem('historyList') || '{}') }

    /**
     * Update history list adds a new history entry into the existing list of entries in historyListToRender and in localStorage.
     * @param obj the new history entry to be added
     */
    private updateHistoryList = (obj: SymbolHistoryEntry): void => {
        const currentList = localStorage.getItem('historyList')
        let newHistoryList = new Array()
        if (currentList) {
            newHistoryList = JSON.parse(currentList)
        }
        if (!isEqual(newHistoryList[0], obj)) {
            newHistoryList = [obj, ...newHistoryList]
        }

        const newHistoryJsonString = JSON.stringify(newHistoryList)
        // Store updated JSON string in localStorage.
        localStorage.setItem('historyList', newHistoryJsonString)
        // Update the list of items to render.
        this.setState({ historyListToRender: newHistoryList })
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged((a, b) => isEqual(a.location, b.location)),
                    switchMap(props => {
                        const loc = parseBrowserRepoURL(
                            props.location.pathname + props.location.search + props.location.hash
                        )
                        if (loc.position && loc.filePath) {
                            const mode = getModeFromPath(loc.filePath)
                            // Get the hover for a given symbol to get title, hover contents, etc.
                            // HighlightedFileLines gets the lines for the file. We use this to get
                            // the lines of code surrounding the symbol.
                            const hoverOrErrorAndFileLinesOrError = forkJoin(
                                getHover(
                                    {
                                        repoPath: loc.repoPath,
                                        commitID: props.commitID,
                                        filePath: loc.filePath,
                                        rev: loc.rev || props.commitID,
                                        position: loc.position,
                                        mode,
                                    },
                                    { extensions: [], cxpController: this.props.cxpController }
                                ).pipe(
                                    catchError(error => {
                                        if (error && error.code === EMODENOTFOUND) {
                                            return [undefined]
                                        }
                                        return [asError(error)]
                                    })
                                ),
                                fetchHighlightedFileLines({
                                    repoPath: loc.repoPath,
                                    filePath: loc.filePath,
                                    commitID: props.commitID,
                                    isLightTheme: true,
                                    disableTimeout: true,
                                })
                            )

                            return hoverOrErrorAndFileLinesOrError
                        }
                        return of([undefined, undefined])
                    })
                )
                .subscribe(
                    ([hoverOrError, fileLinesOrError]) => {
                        this.setState({ hoverOrError, fileLinesOrError })
                        const parsedRepoURI = parseBrowserRepoURL(
                            this.props.location.pathname + this.props.location.search + this.props.location.hash
                        )
                        if (HoverMerged.is(hoverOrError) && parsedRepoURI.filePath) {
                            const name = getSymbolSignature(hoverOrError.contents)
                            const hoverContents = hoverContentsToString(hoverOrError.contents).slice(1)
                            const position = parsedRepoURI.position
                            let lineNumber = 0
                            let surroundingLinesOfCode = fileLinesOrError
                            if (position) {
                                lineNumber = position.line
                                let startLine = lineNumber - 2
                                if (startLine < 0) {
                                    startLine = 0
                                }
                                const endLine = lineNumber + 2
                                // We store the 5 lines of code around the line in question. This is used as part of the dataset for filtering
                                // because want the 5 surrounding lines to inform/influence the filter matches.
                                surroundingLinesOfCode = fileLinesOrError.slice(startLine, endLine)
                            }
                            let rawString = name
                            rawString = rawString.concat(...surroundingLinesOfCode)
                            rawString = rawString.concat(...hoverContents)
                            const obj: SymbolHistoryEntry = {
                                name,
                                url: this.props.location.pathname + this.props.location.hash,
                                repoPath: displayRepoPath(parsedRepoURI.repoPath),
                                filePath: parsedRepoURI.filePath,
                                hoverContents,
                                lineNumber,
                                rawString,
                            }
                            this.updateHistoryList(obj)
                        }
                    },
                    err => console.error(err)
                )
        )

        this.subscriptions.add(
            this.filterQueryUpdates
                .pipe(
                    distinctUntilChanged(),
                    debounceTime(200),
                    startWith(this.state.filter)
                )
                .subscribe(query => this.setState({ filter: query }))
        )
        this.componentUpdates.next(this.props)
    }

    public shouldComponentUpdate(nextProps: HistoryProps, nextState: HistoryState): boolean {
        if (
            this.props.location !== nextProps.location ||
            this.state.historyListToRender !== nextState.historyListToRender ||
            this.state.filter !== nextState.filter
        ) {
            return true
        }
        return false
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private onFilterChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        this.filterQueryUpdates.next(e.currentTarget.value)
    }

    public render(): JSX.Element {
        let prevFile = ''
        return (
            <>
                <input
                    className="form-control filtered-connection__filter"
                    type="search"
                    placeholder={`Filter history...`}
                    name="query"
                    onChange={this.onFilterChange}
                />
                {this.state.historyListToRender && this.state.historyListToRender.length > 0 ? (
                    this.state.historyListToRender
                        .filter((item, i) => {
                            if (
                                !this.state.filter ||
                                item.rawString.toLowerCase().includes(this.state.filter.toLowerCase())
                            ) {
                                return true
                            }
                            return false
                        })
                        .map((item, i) => {
                            const element = (
                                <div key={item.url + i + ' separator'}>
                                    {item.filePath !== prevFile && (
                                        <div className="repo-rev-sidebar-history__header">
                                            {item.repoPath} - {item.filePath}
                                        </div>
                                    )}
                                    <Link to={item.url} key={item.url + i}>
                                        <li className="repo-rev-sidebar-history_list-item list-group-item">
                                            <span
                                                className="repo-rev-sidebar-history__symbol-title"
                                                dangerouslySetInnerHTML={{ __html: item.name }}
                                            />
                                            <span>
                                                <small className="repo-rev-sidebar-history__item-info text-muted">{`
                                            ${item.repoPath} - ${item.filePath} ${item.lineNumber &&
                                                    `- L${item.lineNumber}`}`}</small>
                                            </span>
                                            {item.hoverContents &&
                                                item.hoverContents.map((item, i) => (
                                                    <div
                                                        key={item + i}
                                                        className="repo-rev-sidebar-history__contents"
                                                        dangerouslySetInnerHTML={{ __html: item }}
                                                    />
                                                ))}
                                        </li>
                                    </Link>
                                </div>
                            )
                            prevFile = item.filePath
                            return element
                        })
                ) : (
                    <p className="repo-rev-sidebar-history__empty">
                        <small>No navigation history yet</small>
                    </p>
                )}
            </>
        )
    }
}
