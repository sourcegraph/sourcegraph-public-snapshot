import copy from 'copy-to-clipboard'
import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import * as H from 'history'
import { isEqual } from 'lodash'
import { escape } from 'lodash'
import marked from 'marked'
import * as React from 'react'
import { forkJoin, of, Subject, Subscription } from 'rxjs'
import { catchError, debounceTime, distinctUntilChanged, startWith, switchMap } from 'rxjs/operators'
import { MarkedString, MarkupContent, MarkupKind } from 'vscode-languageserver-types'
import { parseBrowserRepoURL } from '.'
import { getHover, HoverMerged } from '../backend/features'
import { EMODENOTFOUND } from '../backend/lsp'
import { displayRepoPath } from '../components/RepoFileLink'
import { CXPControllerProps } from '../cxp/CXPEnvironment'
import { eventLogger } from '../tracking/eventLogger'
import { getModeFromPath } from '../util'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { fetchHighlightedFileLines } from './backend'
import { RepoRevSidebarHistoryEntry } from './RepoRevSidebarHistoryEntry'

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
        let signature: string
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
                    signature = rendered
                } catch (err) {
                    signature = 'errored'
                }
            } else {
                signature = content.value
            }
        } else {
            signature = highlightCodeSafe(content.value, content.language)
        }
        contentList.push(signature)
    }

    return contentList
}

const getSymbolSignature = (contents: (MarkupContent | MarkedString)[]): string => {
    const symbolSignature = [contents[0]]
    return hoverContentsToString(symbolSignature)[0]
}

const saveToLocalStorage = (storageKey: string, value: string): void => {
    localStorage.setItem(storageKey, value)
}

const getFromLocalStorage = (storageKey: string): Mode => {
    const mode = localStorage.getItem(storageKey)
    if (mode !== null && mode) {
        return mode as Mode
    }
    // Default to doc
    return 'DOC'
}

interface HistoryProps extends CXPControllerProps {
    location: H.Location
    history: H.History
    commitID: string
}

/** The data stored for each history entry. */
export interface SymbolHistoryEntry {
    /** The symbol name */
    name: string
    repoPath: string
    filePath: string
    url: string
    lineNumber?: number
    /** Hover contents, excluding the symbol name */
    hoverContents: string[]
    /** The actual line of code the symbol is in and 5 surrounding lines */
    linesOfCode: string[]
    /** Combination of name, hover contents, lines of code. */
    rawString: string
}

const LAST_MODE_KEY = 'repo-rev-sidebar-history-mode'
export type Mode = 'CODE' | 'DOC'

interface HistoryState {
    hoverOrError?: HoverMerged | ErrorLike | null
    /** */
    fileLinesOrError?: string[] | ErrorLike | null
    /** The list of history entries to render */
    historyListToRender: SymbolHistoryEntry[]
    /** Query in the filter bar */
    filter?: string
    mode: Mode
    /** List of elements selected for copying. The key is the index of the selected symbol in the history list. The value is the URL of the symbol. */
    selected: { [key: number]: string | undefined }
    /** The index of the last selected element so we can handle shift+click selecting ranges. */
    lastSelected?: number
}

export class RepoRevSidebarHistory extends React.Component<HistoryProps, HistoryState> {
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<HistoryProps>()
    private filterQueryUpdates = new Subject<string>()
    public state: HistoryState = {
        historyListToRender: JSON.parse(localStorage.getItem('historyList') || '{}'),
        mode: getFromLocalStorage(LAST_MODE_KEY),
        selected: {},
    }

    /**
     * Update the list of selected entries for copying. Handles shift+click to select ranges.
     */
    private updateSelected = (shiftKey: boolean, url: string, checked: boolean, index: number): void => {
        const selected: { [key: number]: string | undefined } = {}
        if (checked) {
            if (this.state.lastSelected !== undefined && shiftKey) {
                let i = 0
                if (this.state.lastSelected < index) {
                    for (i = this.state.lastSelected; i <= index; i++) {
                        selected[i] = this.state.historyListToRender[i].url
                    }
                } else {
                    for (i = index; i <= this.state.lastSelected; i++) {
                        selected[i] = this.state.historyListToRender[i].url
                    }
                }
            } else {
                selected[index] = url
            }

            this.setState({ selected: { ...this.state.selected, ...selected }, lastSelected: index })
        } else {
            const selected: { [key: number]: string | undefined } = {}
            if (this.state.lastSelected !== undefined && shiftKey) {
                let i = 0
                if (this.state.lastSelected < index) {
                    for (i = this.state.lastSelected; i <= index; i++) {
                        selected[i] = undefined
                    }
                } else {
                    for (i = index; i <= this.state.lastSelected; i++) {
                        selected[i] = undefined
                    }
                }
            } else {
                selected[index] = undefined
            }
            this.setState({ selected: { ...this.state.selected, ...selected }, lastSelected: index })
        }
    }

    /**
     * Update history list adds a new history entry into the existing list of entries in historyListToRender and in localStorage.
     * It trims the list to 2500 entries, which is the same number we use for the file tree to avoid rendering problems.
     * We also don't want to crash the browser by overflowing localStorage, which could happen after a lot of use.
     * @param obj the new history entry to be added
     */
    private updateHistoryList = (obj: SymbolHistoryEntry): void => {
        const currentList = localStorage.getItem('historyList')
        let newHistoryList: SymbolHistoryEntry[] = []
        if (currentList) {
            newHistoryList = JSON.parse(currentList).slice(0, 2500)
        }
        // Don't add to list if it's the same entry as the previous one.
        if (!isEqual(newHistoryList[0], obj)) {
            newHistoryList = [obj, ...newHistoryList]
        }

        const newHistoryJsonString = JSON.stringify(newHistoryList)
        // Store updated JSON string in localStorage.
        localStorage.setItem('historyList', newHistoryJsonString)
        // Update the list of items to render.
        this.setState({ historyListToRender: newHistoryList, mode: getFromLocalStorage(LAST_MODE_KEY) })
    }

    private copyToClipboard = () => {
        eventLogger.log('SymbolHistoryCopySelectedClicked')
        const strings = Object.values(this.state.selected)
            .filter(Boolean)
            .map(val => `${window.context.appURL}${val}`)
            .join('\n')

        copy(strings)
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
                            const hoverOrErrorAndFileLines = forkJoin(
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
                                }).pipe(catchError(error => [asError(error)]))
                            )

                            return hoverOrErrorAndFileLines
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

                        if (
                            HoverMerged.is(hoverOrError) &&
                            parsedRepoURI.filePath &&
                            fileLinesOrError &&
                            !isErrorLike(fileLinesOrError)
                        ) {
                            const name = getSymbolSignature(hoverOrError.contents)
                            let hoverContents = hoverContentsToString(hoverOrError.contents).slice(1)
                            const position = parsedRepoURI.position
                            let lineNumber = 0
                            let surroundingLinesOfCode = fileLinesOrError as string[]
                            if (position) {
                                lineNumber = position.line
                                let startLine = lineNumber - 3
                                if (startLine < 0) {
                                    startLine = 0
                                }
                                const endLine = lineNumber + 3
                                surroundingLinesOfCode = surroundingLinesOfCode.slice(startLine, endLine)
                            }

                            // Only show first 500 characers of hover documentation
                            if (hoverContents.length > 0 && hoverContents[0].length > 500) {
                                const div = document.createElement('div')
                                div.innerHTML = hoverContents[0].slice(0, 500)
                                if (div.lastChild && div.lastChild.textContent) {
                                    const span = document.createElement('span')
                                    span.textContent = '...'
                                    div.lastChild.appendChild(span)
                                }
                                hoverContents = [div.outerHTML]
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
                                linesOfCode: surroundingLinesOfCode,
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
            this.state.filter !== nextState.filter ||
            this.state.mode !== nextState.mode ||
            !isEqual(this.state.selected, nextState.selected)
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

    private setModeDoc = (): void => {
        eventLogger.log('SymbolHistoryDocsClicked')
        saveToLocalStorage(LAST_MODE_KEY, 'DOC')
        this.setState({ mode: 'DOC' })
    }

    private setModeCode = (): void => {
        eventLogger.log('SymbolHistoryCodeClicked')
        saveToLocalStorage(LAST_MODE_KEY, 'CODE')
        this.setState({ mode: 'CODE' })
    }

    public render(): JSX.Element {
        let prevFile = ''
        return (
            <>
                <div className="repo-rev-sidebar-history__list">
                    <input
                        className="form-control filtered-connection__filter"
                        type="search"
                        placeholder="Filter history..."
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
                                    <RepoRevSidebarHistoryEntry
                                        symbolHistoryEntry={item}
                                        index={i}
                                        prevFile={prevFile}
                                        mode={this.state.mode}
                                        onSelect={this.updateSelected}
                                        key={i + item.url}
                                        selected={!!this.state.selected[i]}
                                    />
                                )
                                prevFile = item.filePath
                                return element
                            })
                    ) : (
                        <p className="repo-rev-sidebar-history__empty">
                            <small>No navigation history yet</small>
                        </p>
                    )}
                </div>
                <div className="repo-rev-sidebar-history__utility-bar">
                    <div>
                        <button
                            className="btn btn-sm btn-primary"
                            onClick={this.setModeDoc}
                            disabled={this.state.mode === 'DOC'}
                        >
                            Doc
                        </button>
                        <button
                            className="btn btn-sm btn-primary"
                            onClick={this.setModeCode}
                            disabled={this.state.mode === 'CODE'}
                        >
                            Code
                        </button>
                    </div>
                    <div>
                        <button className="btn btn-sm btn-primary" onClick={this.copyToClipboard}>
                            Copy selected
                        </button>
                    </div>
                </div>
            </>
        )
    }
}
