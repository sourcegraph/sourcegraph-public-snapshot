import CloseIcon from '@sourcegraph/icons/lib/Close'
import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import { highlight } from 'highlight.js'
import * as H from 'history'
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
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { scan } from 'rxjs/operators/scan'
import { skip } from 'rxjs/operators/skip'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Hover, Location } from 'vscode-languageserver-types'
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
import { Spacer, Tab, TabsWithURLViewStatePersistence } from '../../../components/Tabs'
import { FileLocationsPanelContent } from '../../../panel/fileLocations/FileLocationsPanel'
import { eventLogger } from '../../../tracking/eventLogger'
import { getModeFromExtension, getPathExtension } from '../../../util'
import { asError, ErrorLike, isErrorLike } from '../../../util/errors'
import { parseHash } from '../../../util/url'
import { AbsoluteRepoFilePosition } from '../../index'
import { fetchExternalReferences } from '../references/backend'

interface Props extends AbsoluteRepoFilePosition {
    location: H.Location
    history: H.History
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
 *
 * NOTE: This is the new version, still feature-flagged off by default.
 */
export class BlobPanel2 extends React.PureComponent<Props, State> {
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
                        const result = fetchServerCapabilities(subject).pipe(
                            catchError(error => [asError(error)]),
                            map(c => ({ serverCapabilitiesOrError: c } as PartialStateUpdate)),
                            publishReplay<PartialStateUpdate>(),
                            refCount()
                        )
                        return merge(
                            of({ serverCapabilitiesOrError: undefined }), // clear old data immediately
                            result
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
                            map(c => ({ hoverOrError: c } as PartialStateUpdate)),
                            publishReplay<PartialStateUpdate>(),
                            refCount()
                        )

                        return merge(
                            of({ hoverOrError: undefined }), // clear old data immediately
                            result,
                            of({ hoverOrError: LOADING }).pipe(delay(150), takeUntil(result)) // delay loading spinner to reduce jitter
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
        const tabs: Tab<BlobPanelTabID>[] = [
            {
                id: 'def',
                label: 'Definition',
            },
            {
                id: 'references',
                label: 'References',
            },
            {
                id: 'references:external',
                label: 'External references',
            },
            {
                id: 'impl',
                label: 'Implementation',
                hidden:
                    !this.state.serverCapabilitiesOrError ||
                    isErrorLike(this.state.serverCapabilitiesOrError) ||
                    // TODO(sqs): implementationProvider is in vscode-languageserver-node repo tag v4.0.0 but is
                    // not published to npm yet.
                    !(this.state.serverCapabilitiesOrError as any).implementationProvider,
            },
        ]

        let title: React.ReactFragment | null
        let titleHTML: string | undefined
        if (this.state.hoverOrError === LOADING) {
            title = <Loader className="icon-inline" />
        } else if (this.state.hoverOrError === undefined) {
            // Don't show loading indicator yet (to reduce UI jitter).
            title = null
        } else if (
            this.state.hoverOrError &&
            !isErrorLike(this.state.hoverOrError) &&
            !isEmptyHover(this.state.hoverOrError)
        ) {
            // Hover with one or more MarkedStrings.
            const markedString = firstMarkedString(this.state.hoverOrError)!
            const value = typeof markedString === 'string' ? markedString : markedString.value
            title = value
            try {
                titleHTML = highlight(getModeFromExtension(getPathExtension(this.props.filePath)), value).value
            } catch (e) {
                // Ignore syntax highlighting errors; plain text will be rendered.
            }
        } else {
            // Error or no hover information.
            //
            // Don't bother showing the error, if any; if it occurs on the panel contents fetches, it will be
            // displayed.
            title = 'Context'
        }

        return (
            <div className="blob-panel2">
                <header className="blob-panel2__header">
                    {titleHTML ? (
                        <code
                            className="blob-panel2__header-title hljs"
                            dangerouslySetInnerHTML={titleHTML ? { __html: titleHTML } : undefined}
                        />
                    ) : (
                        <code className="blob-panel2__header-title">{title}</code>
                    )}
                    <button
                        onClick={this.onDismiss}
                        className="btn btn-icon blob-panel2__header-close-button"
                        data-tooltip="Close"
                    >
                        <CloseIcon />
                    </button>
                </header>
                <TabsWithURLViewStatePersistence
                    tabs={tabs}
                    tabBarEndFragment={<Spacer />}
                    className="blob-references-panel"
                    tabClassName="tab-bar__tab--h5like"
                    onSelectTab={this.onSelectTab}
                    location={this.props.location}
                >
                    <FileLocationsPanelContent
                        key="def"
                        className="blob-references-panel__content"
                        query={this.queryDefinition}
                        updates={this.locationsUpdates}
                        inputRepo={this.props.repoPath}
                        inputRevision={this.props.rev}
                        // tslint:disable-next-line:jsx-no-lambda
                        onSelect={() => this.onSelectLocation('def')}
                        icon={RepoIcon}
                        isLightTheme={this.props.isLightTheme}
                    />
                    <FileLocationsPanelContent
                        key="references"
                        className="blob-references-panel__content"
                        query={this.queryReferencesLocal}
                        updates={this.locationsUpdates}
                        inputRepo={this.props.repoPath}
                        inputRevision={this.props.rev}
                        // tslint:disable-next-line:jsx-no-lambda
                        onSelect={() => this.onSelectLocation('references')}
                        icon={RepoIcon}
                        isLightTheme={this.props.isLightTheme}
                    />
                    <FileLocationsPanelContent
                        key="references:external"
                        className="blob-references-panel__content"
                        query={this.queryReferencesExternal}
                        updates={this.locationsUpdates}
                        // tslint:disable-next-line:jsx-no-lambda
                        onSelect={() => this.onSelectLocation('references:external')}
                        icon={RepoIcon}
                        isLightTheme={this.props.isLightTheme}
                    />
                    <FileLocationsPanelContent
                        key="impl"
                        className="blob-references-panel__content"
                        query={this.queryImplementation}
                        updates={this.locationsUpdates}
                        inputRepo={this.props.repoPath}
                        inputRevision={this.props.rev}
                        // tslint:disable-next-line:jsx-no-lambda
                        onSelect={() => this.onSelectLocation('impl')}
                        icon={RepoIcon}
                        isLightTheme={this.props.isLightTheme}
                    />
                </TabsWithURLViewStatePersistence>
            </div>
        )
    }

    private onSelectTab = (tab: string): void => eventLogger.log('BlobPanelTabActivated', { tab })
    private onSelectLocation = (tab: BlobPanelTabID): void => eventLogger.log('BlobPanelLocationSelected', { tab })

    private onDismiss = (): void => {
        this.props.history.push(TabsWithURLViewStatePersistence.urlForTabID(this.props.location, null))
    }

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
