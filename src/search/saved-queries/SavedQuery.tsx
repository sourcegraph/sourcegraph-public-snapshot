import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import PencilIcon from 'mdi-react/PencilIcon'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { startWith, switchMap, withLatestFrom } from 'rxjs/operators'
import * as GQL from '../../backend/graphqlschema'
import { eventLogger } from '../../tracking/eventLogger'
import { createSavedQuery, deleteSavedQuery } from '../backend'
import { SavedQueryRow } from './SavedQueryRow'
import { SavedQueryUpdateForm } from './SavedQueryUpdateForm'

interface Props {
    user: GQL.IUser | null
    savedQuery: GQL.ISavedQuery
    onDidUpdate?: () => void
    onDidDuplicate?: () => void
    onDidDelete?: () => void
    isLightTheme: boolean
}

interface State {
    isEditing: boolean
    isSaving: boolean
    loading: boolean
    error?: Error
    approximateResultCount?: string
    sparkline?: number[]
    refreshedAt: number
    redirect: boolean
}

export class SavedQuery extends React.PureComponent<Props, State> {
    public state: State = { isEditing: false, isSaving: false, loading: true, refreshedAt: 0, redirect: false }

    private componentUpdates = new Subject<Props>()
    private refreshRequested = new Subject<GQL.ISavedQuery>()
    private duplicateRequested = new Subject<void>()
    private deleteRequested = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const propsChanges = this.componentUpdates.pipe(startWith(this.props))

        this.subscriptions.add(
            this.duplicateRequested
                .pipe(
                    withLatestFrom(propsChanges),
                    switchMap(([, props]) =>
                        createSavedQuery(
                            props.savedQuery.subject,
                            duplicate(props.savedQuery.description),
                            props.savedQuery.query,
                            props.savedQuery.showOnHomepage,
                            props.savedQuery.notify,
                            props.savedQuery.notifySlack
                        )
                    )
                )
                .subscribe(
                    newSavedQuery => {
                        if (this.props.onDidDuplicate) {
                            this.props.onDidDuplicate()
                        }
                    },
                    err => {
                        console.error(err)
                    }
                )
        )

        this.subscriptions.add(
            this.deleteRequested
                .pipe(
                    withLatestFrom(propsChanges),
                    switchMap(([, props]) => deleteSavedQuery(props.savedQuery.subject, props.savedQuery.id))
                )
                .subscribe(
                    () => {
                        if (this.props.onDidDelete) {
                            this.props.onDidDelete()
                        }
                    },
                    err => {
                        console.error(err)
                    }
                )
        )
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        return (
            <SavedQueryRow
                query={this.props.savedQuery.query}
                description={this.props.savedQuery.description}
                className={this.state.isEditing ? 'editing' : ''}
                eventName="SavedQueryClick"
                isLightTheme={this.props.isLightTheme}
                actions={
                    <div className="saved-query-row__actions">
                        {!this.state.isEditing && (
                            <button className="btn btn-icon action" onClick={this.toggleEditing}>
                                <PencilIcon className="icon-inline" />
                                Edit
                            </button>
                        )}
                        {!this.state.isEditing && (
                            <button className="btn btn-icon action" onClick={this.duplicate}>
                                <ContentCopyIcon className="icon-inline" />
                                Duplicate
                            </button>
                        )}
                        <button className="btn btn-icon action" onClick={this.confirmDelete}>
                            <DeleteIcon className="icon-inline" />
                            Delete
                        </button>
                    </div>
                }
                form={
                    this.state.isEditing && (
                        <div className="saved-query-row__row">
                            <SavedQueryUpdateForm
                                user={this.props.user}
                                savedQuery={this.props.savedQuery}
                                onDidUpdate={this.onDidUpdateSavedQuery}
                                onDidCancel={this.toggleEditing}
                            />
                        </div>
                    )
                }
            />
        )
    }

    private toggleEditing = (e?: React.MouseEvent<HTMLElement>) => {
        if (e) {
            e.stopPropagation()
            e.preventDefault()
        }
        eventLogger.log('SavedQueryToggleEditing', { queries: { editing: !this.state.isEditing } })
        this.setState(state => ({ isEditing: !state.isEditing }))
    }

    private onDidUpdateSavedQuery = () => {
        eventLogger.log('SavedQueryUpdated')
        this.setState({ isEditing: false, approximateResultCount: undefined, loading: true }, () => {
            this.refreshRequested.next()
            if (this.props.onDidUpdate) {
                this.props.onDidUpdate()
            }
        })
    }

    private duplicate = (e: React.MouseEvent<HTMLElement>) => {
        e.stopPropagation()
        e.preventDefault()
        this.duplicateRequested.next()
    }

    private confirmDelete = (e: React.MouseEvent<HTMLElement>) => {
        e.stopPropagation()
        e.preventDefault()
        if (window.confirm('Really delete this saved query?')) {
            eventLogger.log('SavedQueryDeleted')
            this.deleteRequested.next()
        } else {
            eventLogger.log('SavedQueryDeletedCanceled')
        }
    }
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
