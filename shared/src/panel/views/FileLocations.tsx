import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, publishReplay, refCount, startWith, switchMap } from 'rxjs/operators'
import { Location } from '../../api/protocol/plainTypes'
import { FetchFileCtx } from '../../components/CodeExcerpt'
import { FileMatch, IFileMatch, ILineMatch } from '../../components/FileMatch'
import { VirtualList } from '../../components/VirtualList'
import { asError } from '../../util/errors'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { propertyIsDefined } from '../../util/types'
import { parseRepoURI, toPrettyBlobURL } from '../../util/url'

export const FileLocationsError: React.FunctionComponent<{ pluralNoun: string; error: ErrorLike }> = ({
    pluralNoun,
    error,
}) => (
    <div className="file-locations__error alert alert-danger m-2">
        <AlertCircleIcon className="icon-inline" /> Error getting {pluralNoun}: {upperFirst(error.message)}
    </div>
)

export const FileLocationsNotFound: React.FunctionComponent<{ pluralNoun: string }> = ({ pluralNoun }) => (
    <div className="file-locations__not-found m-2">
        <MapSearchIcon className="icon-inline" /> No {pluralNoun} found
    </div>
)

interface Props {
    /**
     * The function called to query for file locations.
     */
    query: () => Observable<Location[]>

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

    /** Called when a location is selected. */
    onSelect?: () => void

    /** The plural noun described by the locations, such as "references" or "implementations". */
    pluralNoun: string

    className?: string

    isLightTheme: boolean

    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
}

const LOADING: 'loading' = 'loading'

interface State {
    /**
     * Locations (inside files identified by LSP-style git:// URIs) to display, loading, or an error if they failed
     * to load.
     */
    locationsOrError: typeof LOADING | Location[] | ErrorLike

    itemsToShow: number
}

/**
 * Displays a flat list of file excerpts. For a tree view, use FileLocationsTree.
 */
export class FileLocations extends React.PureComponent<Props, State> {
    public state: State = {
        locationsOrError: LOADING,
        itemsToShow: 3,
    }

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
                    switchMap(query => query().pipe(catchError(error => [asError(error) as ErrorLike]))),
                    startWith(LOADING),
                    map(result => ({ locationsOrError: result }))
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
        if (this.state.locationsOrError === LOADING) {
            return <LoadingSpinner className="icon-inline m-1" />
        }
        if (this.state.locationsOrError.length === 0) {
            return <FileLocationsNotFound pluralNoun={this.props.pluralNoun} />
        }

        // Locations by fully qualified URI, like git://github.com/gorilla/mux?rev#mux.go
        const locationsByURI = new Map<string, Location[]>()

        // URIs with >0 locations, in order (to avoid jitter as more results stream in).
        const orderedURIs: { uri: string; repo: string }[] = []

        if (this.state.locationsOrError) {
            for (const loc of this.state.locationsOrError) {
                if (!locationsByURI.has(loc.uri)) {
                    locationsByURI.set(loc.uri, [])

                    const { repoPath } = parseRepoURI(loc.uri)
                    orderedURIs.push({ uri: loc.uri, repo: repoPath })
                }
                locationsByURI.get(loc.uri)!.push(loc)
            }
        }

        return (
            <div className={`file-locations ${this.props.className || ''}`}>
                <VirtualList
                    itemsToShow={this.state.itemsToShow}
                    onShowMoreItems={this.onShowMoreItems}
                    items={orderedURIs.map(({ uri, repo }, i) => (
                        <FileMatch
                            key={i}
                            expanded={true}
                            result={refsToFileMatch(
                                uri,
                                repo === this.props.inputRepo ? this.props.inputRevision : undefined,
                                locationsByURI.get(uri)!
                            )}
                            icon={this.props.icon}
                            onSelect={this.onSelect}
                            showAllMatches={true}
                            isLightTheme={this.props.isLightTheme}
                            fetchHighlightedFileLines={this.props.fetchHighlightedFileLines}
                        />
                    ))}
                />
            </div>
        )
    }

    private onShowMoreItems = (): void => {
        this.setState(state => ({ itemsToShow: state.itemsToShow + 3 }))
    }

    private onSelect = (): void => {
        if (this.props.onSelect) {
            this.props.onSelect()
        }
    }
}

function refsToFileMatch(uri: string, rev: string | undefined, refs: Location[]): IFileMatch {
    const p = parseRepoURI(uri)
    return {
        file: {
            path: p.filePath || '',
            url: toPrettyBlobURL({ repoPath: p.repoPath, filePath: p.filePath!, rev: rev || p.commitID || '' }),
            commit: {
                oid: p.commitID || p.rev || rev,
            },
        },
        repository: {
            name: p.repoPath,
            // This is the only usage of toRepoURL, and it is arguably simpler than getting the value from the
            // GraphQL API. We will be removing these old-style git: URIs eventually, so it's not worth fixing this
            // deprecated usage.
            //
            // tslint:disable-next-line deprecation
            url: toRepoURL(p.repoPath),
        },
        limitHit: false,
        lineMatches: refs.filter(propertyIsDefined('range')).map(
            (ref): ILineMatch => ({
                preview: '',
                limitHit: false,
                lineNumber: ref.range.start.line,
                offsetAndLengths: [[ref.range.start.character, ref.range.end.character - ref.range.start.character]],
            })
        ),
    }
}

/**
 * Returns the URL path for the given repository name.
 *
 * @deprecated Obtain the repository's URL from the GraphQL Repository.url field instead.
 */
function toRepoURL(repoName: string): string {
    return `/${repoName}`
}
