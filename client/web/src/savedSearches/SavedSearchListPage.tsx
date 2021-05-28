import DeleteIcon from 'mdi-react/DeleteIcon'
import MessageTextOutlineIcon from 'mdi-react/MessageTextOutlineIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../components/alerts'
import { NamespaceProps } from '../namespaces'
import { PatternTypeProps } from '../search'
import { deleteSavedSearch, fetchSavedSearches } from '../search/backend'
import { eventLogger } from '../tracking/eventLogger'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

interface NodeProps extends RouteComponentProps, Omit<PatternTypeProps, 'setPatternType'> {
    savedSearch: GQL.ISavedSearch
    onDelete: () => void
}

interface NodeState {
    isDeleting: boolean
}

class SavedSearchNode extends React.PureComponent<NodeProps, NodeState> {
    constructor(props: NodeProps) {
        super(props)
        this.state = { isDeleting: false }
    }

    private subscriptions = new Subscription()
    private delete = new Subject<GQL.ISavedSearch>()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.delete
                .pipe(
                    switchMap(search =>
                        deleteSavedSearch(search.id).pipe(
                            mapTo(undefined),
                            catchError(error => {
                                console.error(error)
                                return []
                            })
                        )
                    )
                )
                .subscribe(() => {
                    eventLogger.log('SavedSearchDeleted')
                    this.setState({ isDeleting: false })
                    this.props.onDelete()
                })
        )
    }
    public render(): JSX.Element | null {
        return (
            <div className="saved-search-list-page__row list-group-item test-saved-search-list-page-row">
                <div className="d-flex">
                    <MessageTextOutlineIcon className="saved-search-list-page__row--icon icon-inline" />
                    <Link
                        to={
                            '/search?' +
                            buildSearchURLQuery(this.props.savedSearch.query, this.props.patternType, false)
                        }
                    >
                        <div className="test-saved-search-list-page-row-title">
                            {this.props.savedSearch.description}
                        </div>
                    </Link>
                </div>
                <div>
                    <Link
                        className="btn btn-secondary btn-sm test-edit-saved-search-button"
                        to={`${this.props.match.path}/${this.props.savedSearch.id}`}
                        data-tooltip="Saved search settings"
                    >
                        <SettingsIcon className="icon-inline" /> Settings
                    </Link>{' '}
                    <button
                        type="button"
                        className="btn btn-sm btn-danger test-delete-saved-search-button"
                        onClick={this.onDelete}
                        disabled={this.state.isDeleting}
                        data-tooltip="Delete saved search"
                    >
                        <DeleteIcon className="icon-inline" />
                    </button>
                </div>
            </div>
        )
    }

    private onDelete = (): void => {
        if (!window.confirm(`Delete the saved search ${this.props.savedSearch.description}?`)) {
            return
        }
        this.setState({ isDeleting: true })
        this.delete.next(this.props.savedSearch)
    }
}

interface State {
    savedSearchesOrError?: GQL.ISavedSearch[] | ErrorLike
}

interface Props extends RouteComponentProps<{}>, NamespaceProps, Omit<PatternTypeProps, 'setPatternType'> {}

export class SavedSearchListPage extends React.Component<Props, State> {
    public subscriptions = new Subscription()
    private refreshRequests = new Subject<void>()

    public state: State = {}

    public componentDidMount(): void {
        this.subscriptions.add(
            this.refreshRequests
                .pipe(
                    startWith(undefined),
                    switchMap(() =>
                        fetchSavedSearches().pipe(
                            catchError(error => {
                                return [asError(error)]
                            })
                        )
                    ),
                    map(savedSearchesOrError => ({ savedSearchesOrError }))
                )
                .subscribe(newState => this.setState(newState as State))
        )
        eventLogger.logViewEvent('SavedSearchListPage')
    }

    public render(): JSX.Element | null {
        let body: JSX.Element
        if (this.state.savedSearchesOrError === undefined) {
            body = <LoadingSpinner />
        } else if (isErrorLike(this.state.savedSearchesOrError)) {
            body = <ErrorAlert className="mb-3" error={this.state.savedSearchesOrError} />
        } else {
            const namespaceSavedSearches = this.state.savedSearchesOrError.filter(
                search => this.props.namespace.id === search.namespace.id
            )
            if (namespaceSavedSearches.length === 0) {
                body = <Container className="text-center text-muted">You haven't created a saved search yet.</Container>
            } else {
                body = (
                    <Container>
                        <div className="list-group list-group-flush">
                            {namespaceSavedSearches.map(search => (
                                <SavedSearchNode
                                    key={search.id}
                                    {...this.props}
                                    savedSearch={search}
                                    onDelete={this.onDelete}
                                />
                            ))}
                        </div>
                    </Container>
                )
            }
        }
        return (
            <div className="saved-search-list-page">
                <PageHeader
                    path={[{ text: 'Saved searches' }]}
                    headingElement="h2"
                    description="Manage notifications and alerts for specific search queries."
                    actions={
                        <Link
                            to={`${this.props.match.path}/add`}
                            className="btn btn-primary test-add-saved-search-button"
                        >
                            <PlusIcon className="icon-inline" /> Add saved search
                        </Link>
                    }
                    className="mb-3"
                />
                {body}
            </div>
        )
    }

    private onDelete = (): void => {
        this.refreshRequests.next()
    }
}
