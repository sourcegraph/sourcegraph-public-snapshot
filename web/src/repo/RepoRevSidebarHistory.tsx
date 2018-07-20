import copy from 'copy-to-clipboard'
import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { debounceTime, distinctUntilChanged, startWith } from 'rxjs/operators'
import { CXPControllerProps } from '../cxp/CXPEnvironment'
import { eventLogger } from '../tracking/eventLogger'
import { SymbolHistoryEntry } from './history/utils'
import { RepoRevSidebarHistoryEntry } from './RepoRevSidebarHistoryEntry'

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
    /** The list of history entries to render */
    historyListToRender: SymbolHistoryEntry[]
}

const LAST_MODE_KEY = 'repo-rev-sidebar-history-mode'
export type Mode = 'CODE' | 'DOC'

interface HistoryState {
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
                        selected[i] = this.props.historyListToRender[i].url
                    }
                } else {
                    for (i = index; i <= this.state.lastSelected; i++) {
                        selected[i] = this.props.historyListToRender[i].url
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
            this.props.historyListToRender !== nextProps.historyListToRender ||
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
                    {this.props.historyListToRender && this.props.historyListToRender.length > 0 ? (
                        this.props.historyListToRender
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
