import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import { highlight } from 'highlight.js'
import * as H from 'history'
import marked from 'marked'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { bufferTime } from 'rxjs/operators/bufferTime'
import { catchError } from 'rxjs/operators/catchError'
import { concat } from 'rxjs/operators/concat'
import { delay } from 'rxjs/operators/delay'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { scan } from 'rxjs/operators/scan'
import { skip } from 'rxjs/operators/skip'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Hover, Location, MarkedString } from 'vscode-languageserver-types'
import { ServerCapabilities } from 'vscode-languageserver/lib/main'
import {
    fetchDefinition,
    fetchHover,
    fetchReferences,
    fetchServerCapabilities,
    firstMarkedString,
    isEmptyHover,
    queryImplementation,
} from '../../../backend/lsp'
import { PanelItemPortal } from '../../../panel/PanelItemPortal'
import { PanelTitlePortal } from '../../../panel/PanelTitlePortal'
import { eventLogger } from '../../../tracking/eventLogger'
import { asError, ErrorLike, isErrorLike } from '../../../util/errors'
import { parseHash } from '../../../util/url'
import { AbsoluteRepoFilePosition } from '../../index'
import { RepoHeaderActionPortal } from '../../RepoHeaderActionPortal'
import { RepoRevSidebarCommits } from '../../RepoRevSidebarCommits'
import { ToggleBlobPanel } from '../actions/ToggleBlobPanel'
import { fetchExternalReferences } from '../references/backend'
import { FileLocations } from './FileLocations'
import { FileLocationsTree } from './FileLocationsTree'

interface Props extends AbsoluteRepoFilePosition {
    location: H.Location
    history: H.History
    repoID: GQLID
    isLightTheme: boolean
}

/** The subject (what the contextual information refers to). */
interface ContextSubject {
    repoPath: string
    commitID: string
    filePath: string
    line: number
    character: number
}

type BlobPanelTabID = 'def' | 'references' | 'references:external' | 'impl'

function toSubject(props: Props): ContextSubject {
    const parsedHash = parseHash(props.location.hash)
    return {
        repoPath: props.repoPath,
        commitID: props.commitID,
        filePath: props.filePath,
        line: parsedHash.line || 1,
        character: parsedHash.character || 1,
    }
}

function subjectIsEqual(a: ContextSubject, b: ContextSubject & { line?: number; character?: number }): boolean {
    return (
        a &&
        b &&
        a.repoPath === b.repoPath &&
        a.commitID === b.commitID &&
        a.filePath === b.filePath &&
        a.line === b.line &&
        a.character === b.character
    )
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The LSP server capabilities information. */
    serverCapabilitiesOrError?: ServerCapabilities | ErrorLike

    /** The hover information for the subject. */
    hoverOrError?: Hover | ErrorLike | typeof LOADING
}

/**
 * A panel on the blob page that displays contextual information.
 */
export class BlobPanel extends React.PureComponent<Props, State> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
    private locationsUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const componentUpdates = this.componentUpdates.pipe(startWith(this.props))

        // Changes to the context subject, including upon the initial mount.
        const subjectChanges = componentUpdates.pipe(
            distinctUntilChanged<Props>((a, b) => subjectIsEqual(toSubject(a), toSubject(b)))
        )

        // Update server capabilities.
        this.subscriptions.add(
            subjectChanges
                .pipe(
                    // This remains the same for all positions/ranges in the file.
                    distinctUntilChanged(
                        (a, b) => a.repoPath === b.repoPath && a.commitID === b.commitID && a.filePath === b.filePath
                    ),
                    switchMap(subject => {
                        type PartialStateUpdate = Pick<State, 'serverCapabilitiesOrError'>
                        return fetchServerCapabilities(subject).pipe(
                            catchError(error => [asError(error)]),
                            map(c => ({ serverCapabilitiesOrError: c } as PartialStateUpdate)),
                            startWith<PartialStateUpdate>({ serverCapabilitiesOrError: undefined })
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )

        // Update hover.
        this.subscriptions.add(
            subjectChanges
                .pipe(
                    switchMap(subject => {
                        type PartialStateUpdate = Pick<State, 'hoverOrError'>
                        const result = fetchHover(subject).pipe(
                            catchError(error => [asError(error)]),
                            map(c => ({ hoverOrError: c } as PartialStateUpdate))
                        )
                        return merge(
                            result,
                            of({ hoverOrError: LOADING }).pipe(delay(150), takeUntil(result)) // delay loading spinner to reduce jitter
                        ).pipe(
                            startWith<PartialStateUpdate>({ hoverOrError: undefined }) // clear old data immediately)
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )

        // Update references when subject changes after the initial mount.
        this.subscriptions.add(subjectChanges.pipe(skip(1)).subscribe(() => this.locationsUpdates.next()))
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let titleRendered: React.ReactFragment | undefined
        let extraRendered: React.ReactFragment | undefined
        if (this.state.hoverOrError === LOADING) {
            titleRendered = <Loader className="icon-inline" />
        } else if (this.state.hoverOrError === undefined) {
            // Don't show loading indicator yet (to reduce UI jitter).
            titleRendered = undefined
        } else if (
            this.state.hoverOrError &&
            !isErrorLike(this.state.hoverOrError) &&
            !isEmptyHover(this.state.hoverOrError)
        ) {
            // Hover with one or more MarkedStrings.
            titleRendered = renderMarkedString(firstMarkedString(this.state.hoverOrError)!)

            if (Array.isArray(this.state.hoverOrError.contents) && this.state.hoverOrError.contents.length >= 2) {
                extraRendered = this.state.hoverOrError.contents.slice(1).map((s, i) => (
                    <div key={i} className="blob-panel__extra-item px-2 pt-1">
                        {renderMarkedString(s)}
                    </div>
                ))
            }
        } else {
            // Error or no hover information.
            //
            // Don't bother showing the error, if any; if it occurs on the panel contents fetches, it will be
            // displayed.
        }

        const isValidToken =
            this.state.hoverOrError && this.state.hoverOrError !== LOADING && !isErrorLike(this.state.hoverOrError)

        return (
            <>
                <RepoHeaderActionPortal
                    position="right"
                    priority={20}
                    element={
                        <ToggleBlobPanel
                            key="toggle-blob-panel"
                            location={this.props.location}
                            history={this.props.history}
                        />
                    }
                />

                {titleRendered && <PanelTitlePortal>{titleRendered}</PanelTitlePortal>}
                {extraRendered && (
                    <PanelItemPortal
                        id="info"
                        label="Info"
                        priority={1}
                        element={<div className="mt-2">{extraRendered}</div>}
                    />
                )}
                {isValidToken && (
                    <PanelItemPortal
                        id="def"
                        label="Definition"
                        priority={0}
                        element={
                            <FileLocations
                                className="panel__tabs-content"
                                query={this.queryDefinition}
                                updates={this.locationsUpdates}
                                inputRepo={this.props.repoPath}
                                inputRevision={this.props.rev}
                                // tslint:disable-next-line:jsx-no-lambda
                                onSelect={() => this.onSelectLocation('def')}
                                icon={RepoIcon}
                                pluralNoun="definitions"
                                isLightTheme={this.props.isLightTheme}
                            />
                        }
                    />
                )}
                {isValidToken && (
                    <PanelItemPortal
                        id="references"
                        label="References"
                        priority={-1}
                        element={
                            <FileLocations
                                className="panel__tabs-content"
                                query={this.queryReferencesLocal}
                                updates={this.locationsUpdates}
                                inputRepo={this.props.repoPath}
                                inputRevision={this.props.rev}
                                // tslint:disable-next-line:jsx-no-lambda
                                onSelect={() => this.onSelectLocation('references')}
                                icon={RepoIcon}
                                pluralNoun="local references"
                                isLightTheme={this.props.isLightTheme}
                            />
                        }
                    />
                )}
                {isValidToken && (
                    <PanelItemPortal
                        id="references:external"
                        label="External references"
                        priority={-2}
                        element={
                            <FileLocationsTree
                                className="panel__tabs-content"
                                query={this.queryReferencesExternal}
                                updates={this.locationsUpdates}
                                // tslint:disable-next-line:jsx-no-lambda
                                onSelectLocation={() => this.onSelectLocation('references:external')}
                                icon={RepoIcon}
                                pluralNoun="external references"
                                isLightTheme={this.props.isLightTheme}
                                location={this.props.location}
                            />
                        }
                    />
                )}
                {isValidToken && (
                    <PanelItemPortal
                        id="impl"
                        label="Implementation"
                        priority={-3}
                        hidden={
                            !this.state.serverCapabilitiesOrError ||
                            isErrorLike(this.state.serverCapabilitiesOrError) ||
                            // TODO(sqs): implementationProvider is in vscode-languageserver-node repo tag v4.0.0 but is
                            // not published to npm yet.
                            !(this.state.serverCapabilitiesOrError as any).implementationProvider
                        }
                        element={
                            <FileLocations
                                className="panel__tabs-content"
                                query={this.queryImplementation}
                                updates={this.locationsUpdates}
                                inputRepo={this.props.repoPath}
                                inputRevision={this.props.rev}
                                // tslint:disable-next-line:jsx-no-lambda
                                onSelect={() => this.onSelectLocation('impl')}
                                icon={RepoIcon}
                                pluralNoun="implementations"
                                isLightTheme={this.props.isLightTheme}
                            />
                        }
                    />
                )}
                <PanelItemPortal
                    id="history"
                    label="File history"
                    priority={-4}
                    element={
                        <RepoRevSidebarCommits
                            key="commits"
                            repoID={this.props.repoID}
                            rev={this.props.rev}
                            filePath={this.props.filePath}
                            history={this.props.history}
                            location={this.props.location}
                        />
                    }
                />
            </>
        )
    }

    private onSelectLocation = (tab: BlobPanelTabID): void => eventLogger.log('BlobPanelLocationSelected', { tab })

    private queryDefinition = (): Observable<{ loading: boolean; locations: Location[] }> =>
        fetchDefinition(this.props).pipe(map(c => ({ loading: false, locations: Array.isArray(c) ? c : [c] })))

    private queryReferencesLocal = (): Observable<{ loading: boolean; locations: Location[] }> =>
        fetchReferences({ ...(this.props as AbsoluteRepoFilePosition), includeDeclaration: false }).pipe(
            map(c => ({ loading: false, locations: c }))
        )

    private queryReferencesExternal = (): Observable<{ loading: boolean; locations: Location[] }> =>
        fetchExternalReferences(this.props).pipe(
            map(c => ({ loading: true, locations: c })),
            concat([{ loading: false, locations: [] }]),
            bufferTime(500), // reduce UI jitter
            scan<{ loading: boolean; locations: Location[] }[], { loading: boolean; locations: Location[] }>(
                (cur, locs) => ({
                    loading: cur.loading && locs.every(({ loading }) => loading),
                    locations: cur.locations.concat(...locs.map(({ locations }) => locations)),
                }),
                { loading: true, locations: [] }
            )
        )

    private queryImplementation = (): Observable<{ loading: boolean; locations: Location[] }> =>
        queryImplementation(this.props).pipe(map(c => ({ loading: false, locations: c })))
}

function renderMarkedString(markedString: MarkedString): React.ReactFragment {
    const value = typeof markedString === 'string' ? markedString : markedString.value
    const language = typeof markedString === 'string' ? 'markdown' : markedString.language
    try {
        if (language === 'markdown') {
            return (
                <div
                    dangerouslySetInnerHTML={{
                        __html: marked(value, { gfm: true, breaks: true, sanitize: true }),
                    }}
                />
            )
        }
        return <code className="hljs" dangerouslySetInnerHTML={{ __html: highlight(language, value).value }} />
    } catch (e) {
        // Ignore rendering or syntax highlighting errors; plain text will be rendered.
    }
    return value
}
