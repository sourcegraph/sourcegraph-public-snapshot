import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, startWith, switchMap, takeUntil } from 'rxjs/operators'
import { Location } from 'sourcegraph/module/protocol/plainTypes'
import { isError } from 'util'
import { parseRepoURI } from '../..'
import { Resizable } from '../../../components/Resizable'
import { ErrorLike, isErrorLike } from '../../../util/errors'
import { asError } from '../../../util/errors'
import { RepositoryIcon } from '../../../util/icons' // TODO: Switch to mdi icon
import { RepoLink } from '../../RepoLink'
import { FileLocations, FileLocationsError, FileLocationsNotFound } from './FileLocations'

interface Props {
    /**
     * The function called to query for file locations.
     */
    query: () => Observable<{ loading: boolean; locations: Location[] }>

    /** An observable that upon emission causes the connection to refresh the data (by calling queryConnection). */
    updates?: Observable<void>

    /**
     * Used along with the "inputRevision" prop to preserve the original Git revision specifier for the current
     * repository.
     */
    inputRepo?: string

    /**
     * If given, use this revision in the link URLs to the files (instead of empty) for locations whose repository
     * matches the "inputRepo" prop.
     */
    inputRevision?: string

    /** The icon to use for each location. */
    icon: React.ComponentType<{ className?: string }>

    /** Called when an item in the tree is selected. */
    onSelectTree?: () => void

    /** Called when a location is selected. */
    onSelectLocation?: () => void

    /** The plural noun described by the locations, such as "references" or "implementations". */
    pluralNoun: string

    className: string

    isLightTheme: boolean

    location: H.Location
}

interface State {
    /**
     * Locations (inside files identified by LSP-style git:// URIs) to display, or an error if they failed to load.
     * Undefined while loading.
     */
    locationsOrError?: Location[] | ErrorLike

    /** Whether to show a loading indicator. */
    loading: boolean

    selectedRepo?: string
}

/**
 * Displays a two-column view, with a repository list, and a list of file excerpts for the selected tree item.
 */
export class FileLocationsTree extends React.PureComponent<Props, State> {
    public state: State = { loading: false }

    private componentUpdates = new Subject<Props>()
    private locationsUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Manually requested refreshes.
        const refreshRequests = new Subject<void>()

        // Changes to the query callback function.
        const queryFuncChanges = this.componentUpdates.pipe(
            map(({ query }) => query),
            distinctUntilChanged()
        )

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
                            )
                        )

                        return merge(
                            result,
                            of({ loading: true }).pipe(
                                delay(50),
                                takeUntil(result)
                            ) // delay loading spinner to reduce jitter
                        ).pipe(startWith<PartialStateUpdate>({ locationsOrError: undefined, loading: false })) // clear old data immediately
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
            return <FileLocationsError pluralNoun={this.props.pluralNoun} error={this.state.locationsOrError} />
        }
        if (!this.state.loading && this.state.locationsOrError && this.state.locationsOrError.length === 0) {
            return <FileLocationsNotFound pluralNoun={this.props.pluralNoun} />
        }

        // Don't show tree until we have at least one item.
        if (this.state.locationsOrError === undefined) {
            return null
        } else if (this.state.loading && this.state.locationsOrError.length === 0) {
            return <LoadingSpinner className="icon-inline p-2" />
        }

        // Locations grouped by repository.
        const locationsByRepo = new Map<string, Location[]>()

        // Repositories with >0 locations, in order (to avoid jitter as more results stream in).
        const orderedRepos: string[] = []

        if (this.state.locationsOrError) {
            for (const loc of this.state.locationsOrError) {
                const { repoPath } = parseRepoURI(loc.uri)
                if (!locationsByRepo.has(repoPath)) {
                    locationsByRepo.set(repoPath, [])
                    orderedRepos.push(repoPath)
                }
                locationsByRepo.get(repoPath)!.push(loc)
            }
        }

        const selectedRepo: string | undefined =
            this.state.selectedRepo && orderedRepos.includes(this.state.selectedRepo)
                ? this.state.selectedRepo
                : orderedRepos[0]

        return (
            <div className={`file-locations-tree ${this.props.className}`}>
                <Resizable
                    className="file-locations-tree__resizable"
                    handlePosition="right"
                    storageKey="file-locations-tree-resizable"
                    defaultSize={200 /* px */}
                    element={
                        <div className="list-group list-group-flush file-locations-tree__list">
                            {orderedRepos.map((repo, i) => (
                                <Link
                                    key={i}
                                    className={`list-group-item file-locations-tree__item ${
                                        selectedRepo === repo ? 'active' : ''
                                    }`}
                                    to={this.props.location}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    onClick={e => this.onSelectTree(e, repo)}
                                >
                                    <span className="file-locations-tree__item-name" title={repo}>
                                        <RepositoryIcon className="icon-inline file-locations-tree__item-icon" />
                                        <span className="file-locations-tree__item-name-text">
                                            <RepoLink to={null} repoPath={repo} />
                                        </span>
                                    </span>
                                    <span className="badge badge-secondary badge-pill file-locations-tree__item-badge">
                                        {locationsByRepo.get(repo)!.length}
                                    </span>
                                </Link>
                            ))}
                            {this.state.loading && <LoadingSpinner className="icon-inline p-2" />}
                        </div>
                    }
                />
                <FileLocations
                    className="file-locations-tree__content"
                    // tslint:disable-next-line:jsx-no-lambda
                    query={() =>
                        of({
                            loading: false,
                            locations: selectedRepo ? locationsByRepo.get(selectedRepo)! : [],
                        })
                    }
                    updates={this.locationsUpdates}
                    inputRepo={this.props.inputRepo}
                    inputRevision={this.props.inputRevision}
                    onSelect={this.props.onSelectLocation}
                    icon={RepositoryIcon}
                    pluralNoun={this.props.pluralNoun}
                    isLightTheme={this.props.isLightTheme}
                />
            </div>
        )
    }

    private onSelectTree = (e: React.MouseEvent<HTMLElement>, repo: string): void => {
        e.preventDefault()

        this.setState({ selectedRepo: repo })

        if (this.props.onSelectTree) {
            this.props.onSelectTree()
        }
    }
}
