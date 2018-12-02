import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, endWith, map, startWith, switchMap } from 'rxjs/operators'
import { Location } from '../../api/protocol/plainTypes'
import { FetchFileCtx } from '../../components/CodeExcerpt'
import { RepositoryIcon } from '../../components/icons' // TODO: Switch to mdi icon
import { RepoLink } from '../../components/RepoLink'
import { Resizable } from '../../components/Resizable'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { asError } from '../../util/errors'
import { parseRepoURI } from '../../util/url'
import { FileLocations, FileLocationsError, FileLocationsNotFound } from './FileLocations'

interface Props {
    /**
     * The function called to query for file locations.
     */
    query: () => Observable<Location[]>

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

    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
}

const LOADING: 'loading' = 'loading'

interface State {
    /**
     * Locations (inside files identified by LSP-style git:// URIs) to display, loading, or an error if they failed
     * to load.
     */
    locationsOrError: typeof LOADING | Location[] | ErrorLike

    locationsComplete: boolean

    selectedRepo?: string
}

/**
 * Displays a two-column view, with a repository list, and a list of file excerpts for the selected tree item.
 */
export class FileLocationsTree extends React.PureComponent<Props, State> {
    public state: State = { locationsOrError: LOADING, locationsComplete: false }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Changes to the query callback function.
        const queryFuncChanges = this.componentUpdates.pipe(
            map(({ query }) => query),
            distinctUntilChanged()
        )

        this.subscriptions.add(
            queryFuncChanges
                .pipe(
                    switchMap(query =>
                        query().pipe(
                            catchError(error => [asError(error) as ErrorLike]),
                            map(result => ({ locationsOrError: result, locationsComplete: false })),
                            startWith<Pick<State, 'locationsOrError' | 'locationsComplete'>>({
                                locationsOrError: LOADING,
                                locationsComplete: false,
                            }),
                            endWith({ locationsComplete: true })
                        )
                    )
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
        if (
            this.state.locationsOrError === LOADING ||
            (!this.state.locationsComplete && this.state.locationsOrError.length === 0)
        ) {
            return <LoadingSpinner className="icon-inline m-1" />
        }
        if (this.state.locationsOrError.length === 0) {
            return <FileLocationsNotFound pluralNoun={this.props.pluralNoun} />
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
                            {!this.state.locationsComplete && <LoadingSpinner className="icon-inline m-2" />}
                        </div>
                    }
                />
                <FileLocations
                    className="file-locations-tree__content"
                    // tslint:disable-next-line:jsx-no-lambda
                    query={() => of(selectedRepo ? locationsByRepo.get(selectedRepo)! : [])}
                    onSelect={this.props.onSelectLocation}
                    icon={RepositoryIcon}
                    pluralNoun={this.props.pluralNoun}
                    isLightTheme={this.props.isLightTheme}
                    fetchHighlightedFileLines={this.props.fetchHighlightedFileLines}
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
