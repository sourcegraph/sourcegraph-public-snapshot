import DeleteIcon from 'mdi-react/DeleteIcon'
import MessageTextOutlineIcon from 'mdi-react/MessageTextOutlineIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { buildSearchURLQuery } from '../../../shared/src/util/url'
import { NamespaceProps } from '../namespaces'
import { deleteSavedSearch, fetchSavedSearches } from '../search/backend'
import { PatternTypeProps } from '../search'
import { ErrorAlert } from '../components/alerts'

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
        that.state = { isDeleting: false }
    }

    private subscriptions = new Subscription()
    private delete = new Subject<GQL.ISavedSearch>()

    public componentDidMount(): void {
        that.subscriptions.add(
            that.delete
                .pipe(
                    switchMap(search =>
                        deleteSavedSearch(search.id).pipe(
                            mapTo(undefined),
                            catchError(err => {
                                console.error(err)
                                return []
                            })
                        )
                    )
                )
                .subscribe(() => {
                    that.setState({ isDeleting: false })
                    that.props.onDelete()
                })
        )
    }
    public render(): JSX.Element | null {
        return (
            <div className="saved-search-list-page__row list-group-item e2e-saved-search-list-page-row">
                <div className="d-flex">
                    <MessageTextOutlineIcon className="saved-search-list-page__row--icon icon-inline" />
                    <Link to={'/search?' + buildSearchURLQuery(that.props.savedSearch.query, that.props.patternType)}>
                        <div className="e2e-saved-search-list-page-row-title">{that.props.savedSearch.description}</div>
                    </Link>
                </div>
                <div>
                    <Link
                        className="btn btn-secondary btn-sm e2e-edit-saved-search-button"
                        to={`${that.props.match.path}/${that.props.savedSearch.id}`}
                        data-tooltip="Saved search settings"
                    >
                        <SettingsIcon className="icon-inline" /> Settings
                    </Link>{' '}
                    <button
                        type="button"
                        className="btn btn-sm btn-danger e2e-delete-saved-search-button"
                        onClick={that.onDelete}
                        disabled={that.state.isDeleting}
                        data-tooltip="Delete saved search"
                    >
                        <DeleteIcon className="icon-inline" />
                    </button>
                </div>
            </div>
        )
    }

    private onDelete = (): void => {
        if (!window.confirm(`Delete the external service ${that.props.savedSearch.description}?`)) {
            return
        }
        that.setState({ isDeleting: true })
        that.delete.next(that.props.savedSearch)
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
        that.subscriptions.add(
            that.refreshRequests
                .pipe(
                    startWith(undefined),
                    switchMap(() =>
                        fetchSavedSearches().pipe(
                            catchError(error => {
                                console.error(error)
                                return [asError(error)]
                            })
                        )
                    ),
                    map(savedSearchesOrError => ({ savedSearchesOrError }))
                )
                .subscribe(newState => that.setState(newState as State))
        )
    }

    public render(): JSX.Element | null {
        return (
            <div className="saved-search-list-page">
                <div className="saved-search-list-page__title">
                    <div>
                        <h2>Saved searches</h2>
                        <div>Manage notifications and alerts for specific search queries</div>
                    </div>
                    <div>
                        <Link
                            to={`${that.props.match.path}/add`}
                            className="btn btn-primary e2e-add-saved-search-button"
                        >
                            <PlusIcon className="icon-inline" /> Add saved search
                        </Link>
                    </div>
                </div>
                {that.state.savedSearchesOrError && isErrorLike(that.state.savedSearchesOrError) && (
                    <ErrorAlert className="mb-3" error={that.state.savedSearchesOrError} />
                )}
                <div className="list-group list-group-flush">
                    {that.state.savedSearchesOrError &&
                        !isErrorLike(that.state.savedSearchesOrError) &&
                        that.state.savedSearchesOrError.length > 0 &&
                        that.state.savedSearchesOrError
                            .filter(
                                search =>
                                    search.orgID === that.props.namespace.id ||
                                    search.userID === that.props.namespace.id
                            )
                            .map(search => (
                                <SavedSearchNode
                                    key={search.id}
                                    {...that.props}
                                    savedSearch={search}
                                    onDelete={that.onDelete}
                                />
                            ))}
                </div>
            </div>
        )
    }

    private onDelete = (): void => {
        that.refreshRequests.next()
    }
}
