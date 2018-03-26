import CloseIcon from '@sourcegraph/icons/lib/Close'
import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as H from 'history'
import isEqual from 'lodash/isEqual'
import omit from 'lodash/omit'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { bufferTime } from 'rxjs/operators/bufferTime'
import { concat } from 'rxjs/operators/concat'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { scan } from 'rxjs/operators/scan'
import { skip } from 'rxjs/operators/skip'
import { startWith } from 'rxjs/operators/startWith'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Location } from 'vscode-languageserver-types'
import { fetchReferences } from '../backend/lsp'
import { FileLocationsPanelContent } from '../panel/fileLocations/FileLocationsPanel'
import { AbsoluteRepoFilePosition, PositionSpec, RangeSpec, RepoFile, RepoFilePosition } from '../repo'
import { eventLogger } from '../tracking/eventLogger'
import { parseHash, toPrettyBlobURL } from '../util/url'
import { fetchExternalReferences } from './backend'

interface Props extends AbsoluteRepoFilePosition {
    location: H.Location
    history: H.History
    isLightTheme: boolean
}

/** The references' subject (what the references refer to). */
interface ReferencesStateSubject {
    repoPath: string
    commitID: string
    filePath: string
    line: number
    character: number
}

interface ReferencesState extends ReferencesStateSubject {
    references: Location[]
    loadingLocal: boolean
    loadingExternal: boolean
}

interface State extends ReferencesState {
    group?: 'local' | 'external'
}

function initialReferencesState(props: Props): ReferencesState {
    return {
        ...referencesStateSubject(props),
        references: [],
        loadingLocal: true,
        loadingExternal: true,
    }
}

function referencesStateSubject(props: Props): ReferencesStateSubject {
    const parsedHash = parseHash(props.location.hash)
    return {
        repoPath: props.repoPath,
        commitID: props.commitID,
        filePath: props.filePath,
        line: parsedHash.line || 1,
        character: parsedHash.character || 1,
    }
}

function referencesStateSubjectIsEqual(
    a: ReferencesStateSubject,
    b: ReferencesStateSubject & { line?: number; character?: number }
): boolean {
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

export class ReferencesWidget extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private locationsUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        const parsedHash = parseHash(props.location.hash)
        this.state = {
            group: parsedHash.modalMode || 'local',
            ...initialReferencesState(props),
        }
    }

    public componentDidMount(): void {
        const componentUpdates = this.componentUpdates.pipe(startWith(this.props))

        // Update references when subject changes.
        this.subscriptions.add(
            componentUpdates
                .pipe(
                    distinctUntilChanged<Props>((a, b) =>
                        referencesStateSubjectIsEqual(referencesStateSubject(a), referencesStateSubject(b))
                    ),
                    skip(1)
                )
                .subscribe(() => this.locationsUpdates.next())
        )
    }

    public componentWillReceiveProps(nextProps: Props): void {
        const parsedHash = parseHash(nextProps.location.hash)
        if (parsedHash.modalMode && parsedHash.modalMode !== this.state.group) {
            this.setState({ group: parsedHash.modalMode })
        }
        if (!isEqual(omit(this.props, 'rev'), omit(nextProps, 'rev'))) {
            this.componentUpdates.next(nextProps)
        }
    }

    public getRefsGroupFromUrl(urlStr: string): 'local' | 'external' {
        if (urlStr.indexOf('$references:local') !== -1) {
            return 'local'
        }
        if (urlStr.indexOf('$references:external') !== -1) {
            return 'external'
        }
        return 'local'
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const ctx: RepoFilePosition = this.props

        return (
            <div className="references-widget">
                <div className="references-widget__title-bar">
                    <h5>
                        <Link
                            className={
                                'references-widget__title-bar-group' +
                                (this.state.group === 'local' ? ' references-widget__title-bar-group--active' : '')
                            }
                            to={toPrettyBlobURL({ ...ctx, referencesMode: 'local' })}
                            onClick={this.onLocalRefsButtonClick}
                        >
                            This repository
                        </Link>
                    </h5>
                    <h5>
                        <Link
                            className={
                                'references-widget__title-bar-group' +
                                (this.state.group === 'external' ? ' references-widget__title-bar-group--active' : '')
                            }
                            to={toPrettyBlobURL({ ...ctx, referencesMode: 'external' })}
                            onClick={this.onShowExternalRefsButtonClick}
                        >
                            Other repositories
                        </Link>
                    </h5>
                    <span className="references-widget__close-icon" onClick={this.onDismiss} data-tooltip="Close">
                        <CloseIcon className="icon-inline" />
                    </span>
                </div>
                <div className="references-widget__groups">
                    {this.state.group === 'local' && (
                        <FileLocationsPanelContent
                            query={this.queryReferencesLocal}
                            updates={this.locationsUpdates}
                            inputRevision={this.props.rev}
                            icon={RepoIcon}
                            onSelect={this.logLocalSelection}
                            isLightTheme={this.props.isLightTheme}
                        />
                    )}
                    {this.state.group === 'external' && (
                        <FileLocationsPanelContent
                            query={this.queryReferencesExternal}
                            updates={this.locationsUpdates}
                            icon={GlobeIcon}
                            onSelect={this.logExternalSelection}
                            isLightTheme={this.props.isLightTheme}
                        />
                    )}
                </div>
            </div>
        )
    }

    private onDismiss = (): void => {
        this.props.history.push(
            // Cast because we want this to have a type with a full absolute position/range but
            // with referencesMode undefined, because the purpose of this call is to remove
            // referencesMode from the URL.
            toPrettyBlobURL({ ...this.props, referencesMode: undefined } as RepoFile &
                Partial<PositionSpec> & { referencesMode: undefined } & Partial<RangeSpec>)
        )
    }
    private onLocalRefsButtonClick = () => eventLogger.log('ShowLocalRefsButtonClicked')
    private onShowExternalRefsButtonClick = () => eventLogger.log('ShowExternalRefsButtonClicked')
    private logLocalSelection = () => eventLogger.log('GoToLocalRefClicked')
    private logExternalSelection = () => eventLogger.log('GoToExternalRefClicked')

    private queryReferencesLocal = (): Observable<{ loading: boolean; locations: Location[] }> =>
        fetchReferences(this.props).pipe(map(c => ({ loading: false, locations: c })))

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
}
