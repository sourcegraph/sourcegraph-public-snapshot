import H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { forkJoin, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, switchMap } from 'rxjs/operators'
import { AbsoluteRepoFile, parseBrowserRepoURL } from '.'
import { getHover, HoverMerged } from '../backend/features'
import * as GQL from '../backend/graphqlschema'
import { EMODENOTFOUND } from '../backend/lsp'
import { CXPControllerProps } from '../cxp/CXPEnvironment'
import { getModeFromPath } from '../util'
import { asError, isErrorLike } from '../util/errors'
import { fetchHighlightedFileLines } from './backend'
import { createSymbolHistoryEntry, SymbolHistoryEntry } from './history/utils'
import { RepoRevSidebar } from './RepoRevSidebar'

interface Props extends AbsoluteRepoFile, CXPControllerProps {
    repoID: GQL.ID
    isDir: boolean
    defaultBranch: string
    className: string
    history: H.History
    location: H.Location
}

interface SymbolHistoryState {
    /** The list of history entries to render for the History sidebar */
    historyListToRender: SymbolHistoryEntry[]
}

export class SymbolHistoryContainer extends React.Component<Props, SymbolHistoryState> {
    public state: SymbolHistoryState = {
        historyListToRender: JSON.parse(localStorage.getItem('historyList') || '{}'),
    }
    private subscriptions = new Subscription()
    private locationUpdates = new Subject<Pick<Props, 'location'>>()
    /**
     * Update history list adds a new symbol history entry to historyListToRender and localStorage.
     * It trims the list to 500 entries, to avoid rendering problems and slowness.
     * We also don't want to crash the browser by overflowing localStorage, which could happen after a lot of use.
     * @param obj the new history entry to be added
     */
    private updateHistoryList = (obj: SymbolHistoryEntry): void => {
        const currentList = localStorage.getItem('historyList')
        let newHistoryList: SymbolHistoryEntry[] = []
        if (currentList) {
            newHistoryList = JSON.parse(currentList).slice(0, 500)
        }
        // Don't add to list if it's the same entry as the previous one.
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
        // Update symbol history when location changes.
        this.subscriptions.add(
            this.locationUpdates
                .pipe(
                    distinctUntilChanged((a, b) => isEqual(a.location, b.location)),
                    switchMap(props => {
                        const loc = parseBrowserRepoURL(
                            props.location.pathname + props.location.search + props.location.hash
                        )
                        if (loc.position && loc.filePath && this.props.commitID) {
                            const mode = getModeFromPath(loc.filePath)
                            // Get the hover for a given symbol to get title, hover contents, etc.
                            // HighlightedFileLines gets the lines for the file. We use this to get
                            // the lines of code surrounding the symbol.
                            const hoverOrErrorAndFileLines = forkJoin(
                                getHover(
                                    {
                                        repoPath: loc.repoPath,
                                        commitID: this.props.commitID,
                                        filePath: loc.filePath,
                                        rev: this.props.rev,
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
                                    commitID: this.props.commitID,
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
                        const parsedRepoURI = parseBrowserRepoURL(
                            this.props.location.pathname + this.props.location.search + this.props.location.hash
                        )

                        if (
                            HoverMerged.is(hoverOrError) &&
                            parsedRepoURI.filePath &&
                            fileLinesOrError &&
                            !isErrorLike(fileLinesOrError)
                        ) {
                            const obj: SymbolHistoryEntry = createSymbolHistoryEntry(
                                parsedRepoURI,
                                hoverOrError,
                                fileLinesOrError as string[],
                                this.props.location.pathname + this.props.location.hash
                            )
                            this.updateHistoryList(obj)
                        }
                    },
                    err => console.error(err)
                )
        )

        this.locationUpdates.next(this.props)
    }

    public componentDidUpdate(prevProps: Props): void {
        if (this.props.location !== prevProps.location) {
            this.locationUpdates.next(this.props)
        }
    }

    public render(): JSX.Element {
        return (
            <RepoRevSidebar
                className="repo-rev-container__sidebar"
                repoID={this.props.repoID}
                repoPath={this.props.repoPath}
                rev={this.props.rev}
                commitID={this.props.commitID}
                filePath={this.props.filePath || ''}
                isDir={this.props.isDir}
                defaultBranch={this.props.defaultBranch}
                history={this.props.history}
                location={this.props.location}
                cxpController={this.props.cxpController}
                historyListToRender={this.state.historyListToRender}
            />
        )
    }
}
