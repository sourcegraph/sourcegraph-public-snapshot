import DeleteIcon from 'mdi-react/DeleteIcon'
import MessageTextOutlineIcon from 'mdi-react/MessageTextOutlineIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { deleteSavedSearch, fetchSavedQueries } from '../../search/backend'
import { OrgAreaPageProps } from '../area/OrgArea'

interface NodeProps extends RouteComponentProps {
    savedSearch: GQL.ISavedSearch
    onDelete: () => void
}

class SavedSearchNode extends React.PureComponent<NodeProps> {
    constructor(props: NodeProps) {
        super(props)
    }

    private subscriptions = new Subscription()
    private delete = new Subject<GQL.ISavedSearch>()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.delete
                .pipe(
                    switchMap(search =>
                        deleteSavedSearch(search.id).pipe(
                            mapTo(void 0),
                            catchError(err => {
                                console.error(err)
                                return []
                            })
                        )
                    )
                )
                .subscribe(() => this.props.onDelete())
        )
    }
    public render(): JSX.Element | null {
        return (
            <div className="org-saved-searches-list-page__row list-group-item">
                <div className="d-flex">
                    <MessageTextOutlineIcon className="org-saved-searches-list-page__row--icon icon-inline" />
                    <div>{this.props.savedSearch.description}</div>
                </div>
                <div>
                    <Link
                        className="btn btn-secondary btn-sm e2e-edit-external-service-button"
                        to={`${this.props.match.path}/${this.props.savedSearch.id}`}
                        data-tooltip="Saved search settings"
                    >
                        <SettingsIcon className="icon-inline" /> Settings
                    </Link>{' '}
                    <button
                        className="btn btn-sm btn-danger e2e-delete-external-service-button"
                        onClick={this.onDelete}
                        // disabled={this.state.loading}
                        data-tooltip="Delete saved search"
                    >
                        <DeleteIcon className="icon-inline" />
                    </button>
                </div>
            </div>
        )
    }

    private onDelete = () => {
        if (!window.confirm(`Delete the external service ${this.props.savedSearch.description}?`)) {
            return
        }
        this.delete.next(this.props.savedSearch)
    }
}

interface State {
    savedSearches: GQL.ISavedSearch[]
}

interface Props extends RouteComponentProps<{}>, OrgAreaPageProps {}

export class OrgSavedSearchesListPage extends React.Component<Props, State> {
    public subscriptions = new Subscription()
    private refreshRequests = new Subject<void>()

    constructor(props: Props) {
        super(props)
        this.state = {
            savedSearches: [],
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.refreshRequests
                .pipe(
                    startWith(void 0),
                    switchMap(fetchSavedQueries),
                    map(savedSearches => ({ savedSearches }))
                )
                .subscribe(newState => this.setState(newState as State), err => console.error(err))
        )
    }

    public render(): JSX.Element | null {
        return (
            <div className="org-saved-searches-list-page">
                <div className="org-saved-searches-list-page__title">
                    <div>
                        <h2>Saved searches</h2>
                        <div>Manage notifications and alerts for specific search queries</div>
                    </div>
                    <div>
                        <Link to={`${this.props.match.path}/add`} className="btn btn-primary">
                            <PlusIcon className="icon-inline" /> Add saved search
                        </Link>
                    </div>
                </div>
                <div className="list-group list-group-flush">
                    {this.state.savedSearches.length > 0 &&
                        this.state.savedSearches
                            .filter(search => search.orgID === this.props.org.id)
                            .map(search => (
                                <SavedSearchNode {...this.props} savedSearch={search} onDelete={this.onDelete} />
                            ))}
                </div>
            </div>
        )
    }

    private onDelete = () => {
        this.refreshRequests.next()
    }
}
