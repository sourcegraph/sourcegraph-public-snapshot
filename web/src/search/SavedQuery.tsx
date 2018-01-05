import CopyIcon from '@sourcegraph/icons/lib/Copy'
import DeleteIcon from '@sourcegraph/icons/lib/Delete'
import Loader from '@sourcegraph/icons/lib/Loader'
import PencilIcon from '@sourcegraph/icons/lib/Pencil'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { debounceTime } from 'rxjs/operators/debounceTime'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { eventLogger } from '../tracking/eventLogger'
import { createSavedQuery, deleteSavedQuery, fetchSearchResultStats } from './backend'
import { buildSearchURLQuery } from './index'
import { SavedQueryUpdateForm } from './SavedQueryUpdateForm'
import { Sparkline } from './Sparkline'

interface Props {
    savedQuery: GQL.ISavedQuery
    onDidUpdate?: () => void
    onDidDuplicate?: () => void
    onDidDelete?: () => void
    hideBottomBorder: boolean
}

interface State {
    editing?: boolean
    loading: boolean
    error?: Error
    approximateResultCount?: string
    sparkline?: number[]
    refreshedAt: number
    redirect: boolean
}

export class SavedQuery extends React.PureComponent<Props, State> {
    public state: State = { editing: false, loading: true, refreshedAt: 0, redirect: false }

    private componentUpdates = new Subject<Props>()
    private refreshRequested = new Subject<GQL.ISavedQuery>()
    private duplicateRequested = new Subject<void>()
    private deleteRequested = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        const propsChanges = this.componentUpdates.pipe(startWith(this.props))

        this.subscriptions.add(
            this.refreshRequested
                .pipe(
                    debounceTime(250),
                    withLatestFrom(propsChanges),
                    map(([v, props]) => v || props.savedQuery),
                    switchMap(savedQuery => fetchSearchResultStats(savedQuery.query)),
                    map(results => ({
                        refreshedAt: Date.now(),
                        approximateResultCount: results.approximateResultCount,
                        sparkline: results.sparkline,
                        loading: false,
                    }))
                )
                .subscribe(
                    newState => this.setState(newState as State),
                    err => {
                        this.setState({
                            refreshedAt: Date.now(),
                            approximateResultCount: '!',
                            loading: false,
                        })
                        console.error(err)
                    }
                )
        )
        this.refreshRequested.next(props.savedQuery)

        this.subscriptions.add(
            this.duplicateRequested
                .pipe(
                    withLatestFrom(propsChanges),
                    switchMap(([, props]) =>
                        createSavedQuery(
                            props.savedQuery.subject,
                            duplicate(props.savedQuery.description),
                            props.savedQuery.query.query,
                            props.savedQuery.showOnHomepage
                        )
                    )
                )
                .subscribe(newSavedQuery => props.onDidDuplicate && props.onDidDuplicate(), err => console.error(err))
        )

        this.subscriptions.add(
            this.deleteRequested
                .pipe(
                    withLatestFrom(propsChanges),
                    switchMap(([, props]) => deleteSavedQuery(props.savedQuery.subject, props.savedQuery.id))
                )
                .subscribe(() => props.onDidDelete && props.onDidDelete(), err => console.error(err))
        )
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className={`saved-query ${this.state.editing ? 'editing' : ''}`}>
                <Link onClick={this.logEvent} to={'/search?' + buildSearchURLQuery(this.props.savedQuery.query)}>
                    <div title={this.props.savedQuery.query.query} className={`saved-query__row`}>
                        <div className="saved-query__row-column">
                            <div className="saved-query__description">{this.props.savedQuery.description}</div>
                            <div className="saved-query__actions">
                                {!this.state.editing && (
                                    <button className="btn btn-icon action" onClick={this.toggleEditing}>
                                        <PencilIcon className="icon-inline" />
                                        Edit
                                    </button>
                                )}
                                {!this.state.editing && (
                                    <button className="btn btn-icon action" onClick={this.duplicate}>
                                        <CopyIcon className="icon-inline" />
                                        Duplicate
                                    </button>
                                )}
                                <button className="btn btn-icon action" onClick={this.confirmDelete}>
                                    <DeleteIcon className="icon-inline" />
                                    Delete
                                </button>
                            </div>
                        </div>
                        <div className="saved-query__results-container">
                            {!this.state.loading &&
                                this.state.sparkline && (
                                    <div title="Results found in the last 30 days" className="saved-query__sparkline">
                                        <Sparkline data={this.state.sparkline} width={200} height={40} />
                                    </div>
                                )}
                            {this.state.loading ? (
                                <Loader className="icon-inline" />
                            ) : (
                                <div className="saved-query__result-count">{this.state.approximateResultCount}</div>
                            )}
                        </div>
                    </div>
                </Link>
                {this.state.editing && (
                    <div className="saved-query__row">
                        <SavedQueryUpdateForm
                            savedQuery={this.props.savedQuery}
                            onDidUpdate={this.onDidUpdateSavedQuery}
                            onDidCancel={this.toggleEditing}
                        />
                    </div>
                )}
            </div>
        )
    }

    private toggleEditing = (e?: React.MouseEvent<HTMLElement>) => {
        if (e) {
            e.stopPropagation()
            e.preventDefault()
        }
        eventLogger.log('SavedQueryToggleEditing', { queries: { editing: !this.state.editing } })
        this.setState({ editing: !this.state.editing })
    }

    private onDidUpdateSavedQuery = () => {
        eventLogger.log('SavedQueryUpdated')
        this.setState({ editing: false, approximateResultCount: undefined, loading: true }, () => {
            this.refreshRequested.next()
            if (this.props.onDidUpdate) {
                this.props.onDidUpdate()
            }
        })
    }

    private duplicate = (e: React.MouseEvent<HTMLElement>) => {
        e.stopPropagation()
        this.duplicateRequested.next()
    }

    private confirmDelete = (e: React.MouseEvent<HTMLElement>) => {
        e.stopPropagation()
        if (window.confirm('Really delete this saved query?')) {
            eventLogger.log('SavedQueryDeleted')
            this.deleteRequested.next()
        } else {
            eventLogger.log('SavedQueryDeletedCanceled')
        }
    }

    private logEvent = () => eventLogger.log('SavedQueryClicked')
}

function duplicate(s: string): string {
    const m = s.match(/ \(copy(?: (\d+))?\)$/)
    if (m && m[1]) {
        return `${s.slice(0, m.index)} (copy ${parseInt(m[1], 10) + 1})`
    }
    if (m) {
        return `${s.slice(0, m.index)} (copy 2)`
    }
    return `${s} (copy)`
}
