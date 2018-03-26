import Loader from '@sourcegraph/icons/lib/Loader'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { combineLatest } from 'rxjs/observable/combineLatest'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { delay } from 'rxjs/operators/delay'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { isError } from 'util'
import { Location } from 'vscode-languageserver-types'
import { VirtualList } from '../../components/VirtualList'
import { FileMatch, IFileMatch, ILineMatch } from '../../search/FileMatch'
import { asError } from '../../util/errors'
import { ErrorLike, isErrorLike } from '../../util/errors'

interface Props {
    /**
     * The function called to query for file locations.
     */
    query: () => Observable<{ loading: boolean; locations: Location[] }>

    /** An observable that upon emission causes the connection to refresh the data (by calling queryConnection). */
    updates?: Observable<void>

    /** If given, use this revision in the link URLs to the files (instead of empty). */
    inputRevision?: string

    /** The icon to use for each location. */
    icon: React.ComponentType<{ className: string }>

    /** Called when a location is selected. */
    onSelect: () => void

    className: string

    isLightTheme: boolean
}

interface State {
    /**
     * Locations (inside files identified by LSP-style git:// URIs) to display, or an error if they failed to load.
     * Undefined while loading.
     */
    locationsOrError?: Location[] | ErrorLike

    /** Whether to show a loading indicator. */
    loading: boolean

    itemsToShow: number
}

/**
 * Displays a list of file excerpts in a panel.
 */
export class FileLocationsPanelContent extends React.PureComponent<Props, State> {
    public state: State = {
        itemsToShow: 3,
        loading: false,
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Manually requested refreshes.
        const refreshRequests = new Subject<void>()

        // Changes to the query callback function.
        const queryFuncChanges = this.componentUpdates.pipe(map(({ query }) => query), distinctUntilChanged())

        // Force updates from parent component.
        if (this.props.updates) {
            this.subscriptions.add(this.props.updates.subscribe(c => refreshRequests.next()))
        }

        this.subscriptions.add(
            combineLatest(queryFuncChanges, refreshRequests.pipe(startWith<void>(void 0)))
                .pipe(
                    switchMap(([query]) => {
                        type PartialStateUpdate = Pick<State, 'locationsOrError' | 'loading'>
                        const result = query().pipe(
                            catchError(error => [asError(error)]),
                            map(
                                c =>
                                    ({
                                        locationsOrError: isError(c) ? c : c.locations,
                                        loading: isError(c) ? false : c.loading,
                                    } as PartialStateUpdate)
                            ),
                            publishReplay<PartialStateUpdate>(),
                            refCount()
                        )

                        return merge(
                            of({ locationsOrError: undefined, loading: false }), // clear old data immediately
                            result,
                            of({ loading: true }).pipe(delay(50), takeUntil(result)) // delay loading spinner to reduce jitter
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (isErrorLike(this.state.locationsOrError)) {
            return (
                <div className="alert alert-danger m-2">Error: {upperFirst(this.state.locationsOrError.message)}</div>
            )
        }
        if (!this.state.loading && this.state.locationsOrError && this.state.locationsOrError.length === 0) {
            return <div className="m-2">No results</div>
        }

        // Locations by fully qualified URI, like git://github.com/gorilla/mux?rev#mux.go
        const locationsByURI = new Map<string, Location[]>()

        // URIs with >0 locations, in order (to avoid jitter as more results stream in).
        const orderedURIs: string[] = []

        if (this.state.locationsOrError) {
            for (const loc of this.state.locationsOrError) {
                if (!locationsByURI.has(loc.uri)) {
                    locationsByURI.set(loc.uri, [])

                    orderedURIs.push(loc.uri)
                }
                locationsByURI.get(loc.uri)!.push(loc)
            }
        }

        return (
            <div className={`file-locations-panel ${this.props.className}`}>
                <VirtualList
                    itemsToShow={this.state.itemsToShow}
                    onShowMoreItems={this.onShowMoreItems}
                    items={orderedURIs.map((uri, i) => (
                        <FileMatch
                            key={i}
                            expanded={true}
                            result={refsToFileMatch(uri, this.props.inputRevision, locationsByURI.get(uri)!)}
                            icon={this.props.icon}
                            onSelect={this.props.onSelect}
                            showAllMatches={true}
                            isLightTheme={this.props.isLightTheme}
                        />
                    ))}
                />
                {this.state.loading && <Loader className="icon-inline p-2" />}
            </div>
        )
    }

    private onShowMoreItems = (): void => {
        this.setState(state => ({ itemsToShow: state.itemsToShow + 3 }))
    }
}

function refsToFileMatch(uri: string, rev: string | undefined, refs: Location[]): IFileMatch {
    const resource = new URL(uri)
    if (rev) {
        resource.search = rev
    }
    return {
        resource: resource.toString(),
        limitHit: false,
        lineMatches: refs.map((ref): ILineMatch => ({
            preview: '',
            limitHit: false,
            lineNumber: ref.range.start.line,
            offsetAndLengths: [[ref.range.start.character, ref.range.end.character - ref.range.start.character]],
        })),
    }
}
