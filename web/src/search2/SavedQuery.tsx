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
import { createSavedQuery, deleteSavedQuery, fetchSearchResultCount } from './backend'
import { buildSearchURLQuery } from './index'
import { QueryButton } from './QueryButton'
import { SavedQueryUpdateForm } from './SavedQueryUpdateForm'

interface Props {
    savedQuery: GQL.ISavedQuery
    onDidUpdate?: () => void
    onDidDuplicate?: () => void
    onDidDelete?: () => void
    isLightTheme: boolean
}

interface State {
    editing?: boolean
    loading: boolean
    error?: Error
    approximateResultCount?: string
    refreshedAt: number
}

export class SavedQuery extends React.PureComponent<Props, State> {
    public state: State = { editing: false, loading: true, refreshedAt: 0 }

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
                    switchMap(savedQuery => fetchSearchResultCount(savedQuery.query)),
                    map(results => ({
                        refreshedAt: Date.now(),
                        approximateResultCount: results.approximateResultCount,
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
                            props.savedQuery.query.scopeQuery
                        )
                    )
                )
                .subscribe(newSavedQuery => props.onDidDuplicate && props.onDidDuplicate(), err => console.error(err))
        )

        this.subscriptions.add(
            this.deleteRequested
                .pipe(
                    withLatestFrom(propsChanges),
                    switchMap(([, props]) => deleteSavedQuery(props.savedQuery.subject, props.savedQuery.index))
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
                <div className="saved-query__row">
                    <h2 className="saved-query__description">{this.props.savedQuery.description}</h2>
                    <h2 className="saved-query__result-count">
                        {this.state.loading ? (
                            <Loader className="icon-inline" />
                        ) : (
                            <Link
                                to={'/search?' + buildSearchURLQuery(this.props.savedQuery.query)}
                                title="Number of results for this query"
                                onMouseUp={this.logEvent}
                            >
                                {this.state.approximateResultCount}
                            </Link>
                        )}
                    </h2>
                </div>
                <div className="saved-query__row">
                    <div className="saved-query__query">
                        <QueryButton query={this.props.savedQuery.query} onMouseUp={this.logEvent} />
                    </div>
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
                {this.state.editing && (
                    <div className="saved-query__row">
                        <SavedQueryUpdateForm
                            savedQuery={this.props.savedQuery}
                            onDidUpdate={this.onDidUpdateSavedQuery}
                            onDidCancel={this.toggleEditing}
                            isLightTheme={this.props.isLightTheme}
                        />
                    </div>
                )}
            </div>
        )
    }

    private toggleEditing = () => {
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

    private duplicate = () => this.duplicateRequested.next()

    private confirmDelete = () => {
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
