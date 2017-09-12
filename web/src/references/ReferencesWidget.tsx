import * as H from 'history'
import groupBy from 'lodash/groupBy'
import partition from 'lodash/partition'
import * as React from 'react'
import DownIcon from 'react-icons/lib/fa/angle-down'
import RightIcon from 'react-icons/lib/fa/angle-right'
import RepoIcon from 'react-icons/lib/go/repo'
import CloseIcon from 'react-icons/lib/md/close'
import GlobeIcon from 'react-icons/lib/md/language'
import { Link } from 'react-router-dom'
import 'rxjs/add/observable/fromPromise'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/switchMap'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { fetchReferences } from 'sourcegraph/backend/lsp'
import { CodeExcerpt } from 'sourcegraph/components/CodeExcerpt'
import { Reference } from 'sourcegraph/references'
import { fetchExternalReferences } from 'sourcegraph/references/backend'
import { AbsoluteRepoFilePosition, RepoFilePosition } from 'sourcegraph/repo'
import { events } from 'sourcegraph/tracking/events'
import * as url from 'sourcegraph/util/url'
import { parseHash } from 'sourcegraph/util/url'

interface ReferenceGroupProps {
    uri: string
    path: string
    refs: Reference[]
    isLocal: boolean
    localRev?: string
    hidden?: boolean
}

interface ReferenceGroupState {
    hidden?: boolean
}

export class ReferencesGroup extends React.Component<ReferenceGroupProps, ReferenceGroupState> {
    constructor(props: ReferenceGroupProps) {
        super(props)
        this.state = { hidden: props.hidden }
    }

    public render(): JSX.Element | null {
        const uriSplit = this.props.uri.split('/')
        const uriStr = uriSplit.length > 1 ? uriSplit.slice(1).join('/') : this.props.uri
        const pathSplit = this.props.path.split('/')
        const filePart = pathSplit.pop()

        let refs: JSX.Element | null = null
        if (!this.state.hidden) {
            refs = (
                <div className='references-group__list'>
                    {
                        this.props.refs
                            .sort((a, b) => {
                                if (a.range.start.line < b.range.start.line) {
                                    return -1
                                }
                                if (a.range.start.line === b.range.start.line) {
                                    if (a.range.start.character < b.range.start.character) {
                                        return -1
                                    }
                                    if (a.range.start.character === b.range.start.character) {
                                        return 0
                                    }
                                    return 1
                                }
                                return 1
                            })
                            .map((ref, i) => {
                                const uri = new URL(ref.uri)
                                const href = this.getRefURL(ref)
                                return (
                                    <Link
                                        to={href}
                                        key={i}
                                        className='references-group__reference'
                                        onClick={this.logEvent}
                                    >
                                        <CodeExcerpt
                                            repoPath={uri.hostname + uri.pathname}
                                            commitID={uri.search.substr('?'.length)}
                                            filePath={uri.hash.substr('#'.length)}
                                            position={{ line: ref.range.start.line, char: ref.range.start.character }}
                                            highlightLength={ref.range.end.character - ref.range.start.character}
                                            previewWindowExtraLines={1}
                                        />
                                    </Link>
                                )
                            })
                    }
                </div>
            )
        }

        return (
            <div className='references-group'>
                <div className='references-group__title' onClick={this.toggle}>
                    {this.props.isLocal ? <RepoIcon className='references-group__icon' /> : <GlobeIcon className='references-group__icon' />}
                    <div className='references-group__uri-path-part'>{uriStr}</div>
                    <div>{pathSplit.join('/')}{pathSplit.length > 0 ? '/' : ''}</div>
                    <div className='references-group__file-path-part'>{filePart}</div>
                    {this.state.hidden ? <RightIcon className='references-group__expand-icon' /> : <DownIcon className='references-group__expand-icon' />}
                </div>
                {refs}
            </div>
        )
    }

    private toggle = () => {
        this.setState({ hidden: !this.state.hidden })
    }

    private logEvent = (): void => {
        (this.props.isLocal ? events.GoToLocalRefClicked : events.GoToExternalRefClicked).log()
    }

    private getRefURL(ref: Reference): string {
        const uri = new URL(ref.uri)
        const rev = this.props.isLocal && this.props.localRev ?
            this.props.localRev :
            uri.search.substr('?'.length)
        return `/${uri.hostname + uri.pathname}${rev ? '@' + rev : ''}/-/blob/${uri.hash.substr('#'.length)}#L${ref.range.start.line + 1}`
    }
}

interface Props extends AbsoluteRepoFilePosition {
    location: H.Location
    history: H.History
}

interface State {
    group: 'local' | 'external'
    references: Reference[]
    loadingLocal: boolean
    loadingExternal: boolean
}

export class ReferencesWidget extends React.Component<Props, State> {
    public state: State = {
        group: 'local',
        references: [],
        loadingLocal: true,
        loadingExternal: true
    }
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        const parsedHash = parseHash(props.location.hash)
        this.state.group = parsedHash.modalMode ? parsedHash.modalMode : 'local'
        this.subscriptions.add(
            this.componentUpdates
                .switchMap(props => Observable.merge(
                    Observable.fromPromise(fetchReferences(props))
                        .map(refs => ({ references: this.state.references.concat(refs) } as State))
                        .catch(e => {
                            console.error(e)
                            return []
                        })
                        .concat([{ loadingLocal: false } as State]),
                    fetchExternalReferences(props)
                        .map(refs => ({ references: this.state.references.concat(refs) } as State))
                        .catch(e => {
                            console.error(e)
                            return []
                        })
                        .concat([{ loadingExternal: false } as State])
                ))
                .subscribe(state => this.setState(state))
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        const parsedHash = parseHash(nextProps.location.hash)
        let group = this.state.group
        if (parsedHash.modalMode) {
            group = parsedHash.modalMode
        }
        this.setState({ references: [], loadingLocal: true, loadingExternal: true, group })
        this.componentUpdates.next(nextProps)
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

    public isLoading(group: string): boolean {
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
            <div className='references-widget'>
                <div className='references-widget__title-bar'>
                    <Link
                        className={'references-widget__title-bar-group' + (this.state.group === 'local' ? ' references-widget__title-bar-group--active' : '')}
                        to={url.toPrettyBlobPositionURL({ ...ctx, referencesMode: 'local' })}
                        onClick={this.onLocalRefsButtonClick}>
                        This repository
                    </Link>
                    <div className='references-widget__badge'>{localRefCount}</div>
                    <Link className={'references-widget__title-bar-group' + (this.state.group === 'external' ? ' references-widget__title-bar-group--active' : '')}
                        to={url.toPrettyBlobPositionURL({ ...ctx, referencesMode: 'external' })}
                        onClick={this.onShowExternalRefsButtonClick}>
                        Other repositories
                    </Link>
                    <div className='references-widget__badge'>{externalRefCount}</div>
                    <CloseIcon className='references-widget__close-icon' onClick={this.onDismiss} />
                </div>
                {
                    isEmptyGroup() && <div className='references-widget__placeholder'>
                        {this.isLoading(this.state.group) ? 'Working...' : 'No results'}
                    </div>
                }
                <div className='references-widget__groups'>
                    {
                        this.state.group === 'local' &&
                            localRefs.sort().map((uri, i) => {
                                const parsed = new URL(uri)
                                return (
                                    <ReferencesGroup
                                        key={i}
                                        uri={parsed.hostname + parsed.pathname}
                                        path={parsed.hash.substr('#'.length)}
                                        isLocal={true}
                                        localRev={this.props.rev}
                                        refs={refsByUri[uri]} />
                                )
                            })
                    }
                    {
                        this.state.group === 'external' &&
                            externalRefs.map((uri, i) => { /* don't sort, to avoid jerky UI as new repo results come in */
                                const parsed = new URL(uri)
                                return (
                                    <ReferencesGroup
                                        key={i}
                                        uri={parsed.hostname + parsed.pathname}
                                        path={parsed.hash.substr('#'.length)}
                                        isLocal={false}
                                        refs={refsByUri[uri]} />
                                )
                            })
                    }
                </div>
            </div>
        )
    }

    private onDismiss = (): void => {
        this.props.history.push(url.toPrettyBlobPositionURL(this.props))
    }
    private onLocalRefsButtonClick = () => events.ShowLocalRefsButtonClicked.log()
    private onShowExternalRefsButtonClick = () => events.ShowExternalRefsButtonClicked.log()
}
