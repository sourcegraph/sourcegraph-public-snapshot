import CloseIcon from '@sourcegraph/icons/lib/Close'
import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as H from 'history'
import groupBy from 'lodash/groupBy'
import isEqual from 'lodash/isEqual'
import omit from 'lodash/omit'
import partition from 'lodash/partition'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { merge } from 'rxjs/observable/merge'
import { bufferTime } from 'rxjs/operators/bufferTime'
import { catchError } from 'rxjs/operators/catchError'
import { concat } from 'rxjs/operators/concat'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { scan } from 'rxjs/operators/scan'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Location } from 'vscode-languageserver-types'
import { fetchReferences } from '../backend/lsp'
import { VirtualList } from '../components/VirtualList'
import { AbsoluteRepoFilePosition, PositionSpec, RangeSpec, RepoFile, RepoFilePosition } from '../repo'
import { FileMatch, IFileMatch, ILineMatch } from '../search/FileMatch'
import { eventLogger } from '../tracking/eventLogger'
import { parseHash, toPrettyBlobURL } from '../util/url'
import { fetchExternalReferences } from './backend'

interface Props extends AbsoluteRepoFilePosition {
    location: H.Location
    history: H.History
}

interface ReferencesState {
    references: Location[]
    loadingLocal: boolean
    loadingExternal: boolean
}

interface State extends ReferencesState {
    group?: 'local' | 'external'
    itemsToShow: number
}

export class ReferencesWidget extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        const parsedHash = parseHash(props.location.hash)
        this.state = {
            group: parsedHash.modalMode || 'local',
            itemsToShow: 3,
            references: [],
            loadingLocal: true,
            loadingExternal: true,
        }
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged(isEqual),
                    switchMap(props =>
                        merge(
                            fetchReferences(props).pipe(
                                map(references => ({ references } as State)),
                                catchError(e => {
                                    console.error(e)
                                    return []
                                }),
                                concat([{ loadingLocal: false } as State])
                            ),
                            fetchExternalReferences(props).pipe(
                                map(references => ({ references } as State)),
                                catchError(e => {
                                    console.error(e)
                                    return []
                                }),
                                concat([{ loadingExternal: false } as State])
                            )
                        )
                    ),
                    bufferTime(500),
                    filter(updates => updates.length > 0),
                    scan<ReferencesState[], ReferencesState>(
                        (currState, updates) => {
                            let newState = currState
                            for (const update of updates) {
                                if (update.references) {
                                    newState = {
                                        ...newState,
                                        references: newState.references.concat(update.references),
                                    }
                                } else {
                                    newState = { ...newState, ...update }
                                }
                            }
                            return newState
                        },
                        { references: [], loadingLocal: true, loadingExternal: true }
                    )
                )
                .subscribe(state => this.setState(state))
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        const parsedHash = parseHash(nextProps.location.hash)
        if (parsedHash.modalMode && parsedHash.modalMode !== this.state.group) {
            this.setState({ group: parsedHash.modalMode, itemsToShow: 3 })
        }
        if (isEqual(omit(this.props, 'rev'), omit(nextProps, 'rev'))) {
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

    public isLoading(group?: string): boolean {
        if (!group) {
            return this.state.loadingLocal
        }
        switch (group) {
            case 'local':
                return this.state.loadingLocal

            case 'external':
                return this.state.loadingExternal
        }
        return false
    }

    public render(): JSX.Element | null {
        const refs = this.state.references

        // References by fully qualified URI, like git://github.com/gorilla/mux?rev#mux.go
        const refsByUri = groupBy(refs, ref => ref.uri)

        const localPrefix = 'git://' + this.props.repoPath
        const [localRefs, externalRefs] = partition(Object.keys(refsByUri), uri => uri.startsWith(localPrefix))

        const localRefCount = localRefs.reduce((memo, uri) => memo + refsByUri[uri].length, 0)
        const externalRefCount = externalRefs.reduce((memo, uri) => memo + refsByUri[uri].length, 0)

        const isEmptyGroup = () => {
            switch (this.state.group) {
                case 'local':
                    return localRefs.length === 0

                case 'external':
                    return externalRefs.length === 0
            }
            return false
        }

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
                    <div
                        className={
                            'references-widget__badge' +
                            (this.state.group === 'local' ? ' references-widget__badge--active' : '')
                        }
                    >
                        {localRefCount}
                    </div>
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
                    <div
                        className={
                            'references-widget__badge' +
                            (this.state.group === 'external' ? ' references-widget__badge--active' : '')
                        }
                    >
                        {externalRefCount}
                    </div>
                    <span className="references-widget__close-icon" onClick={this.onDismiss} data-tooltip="Close">
                        <CloseIcon className="icon-inline" />
                    </span>
                </div>
                {isEmptyGroup() && (
                    <div className="references-widget__placeholder">
                        {this.isLoading(this.state.group) ? 'Working...' : 'No results'}
                    </div>
                )}
                <div className="references-widget__groups">
                    {this.state.group === 'local' &&
                        this.state.itemsToShow && (
                            <VirtualList
                                itemsToShow={this.state.itemsToShow}
                                onShowMoreItems={this.onShowMoreItems}
                                items={localRefs
                                    .sort()
                                    .map((uri, i) => (
                                        <FileMatch
                                            key={i}
                                            expanded={true}
                                            result={refsToFileMatch(uri, this.props.rev, refsByUri[uri])}
                                            icon={RepoIcon}
                                            onSelect={this.logLocalSelection}
                                            showAllMatches={true}
                                        />
                                    ))}
                            />
                        )}
                    {this.state.group === 'external' &&
                        this.state.itemsToShow && (
                            <VirtualList
                                itemsToShow={this.state.itemsToShow}
                                onShowMoreItems={this.onShowMoreItems}
                                items={externalRefs.map((uri, i) => (
                                    /* don't sort, to avoid jerky UI as new repo results come in */
                                    <FileMatch
                                        key={i}
                                        expanded={true}
                                        result={refsToFileMatch(uri, undefined, refsByUri[uri])}
                                        icon={GlobeIcon}
                                        onSelect={this.logExternalSelection}
                                        showAllMatches={true}
                                    />
                                ))}
                            />
                        )}
                </div>
            </div>
        )
    }

    private onShowMoreItems = (): void => {
        this.setState(state => ({ itemsToShow: state.itemsToShow + 3 }))
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
