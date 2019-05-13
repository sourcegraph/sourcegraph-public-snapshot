import MessageTextOutlineIcon from 'mdi-react/MessageTextOutlineIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { map, startWith, switchMap } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { fetchSavedQueries } from '../../search/backend'
import { OrgAreaPageProps } from '../area/OrgArea'

interface State {
    savedSearches: GQL.ISavedSearch[]
}

interface Props extends RouteComponentProps<{}>, OrgAreaPageProps {}

export class OrgSavedSearchesListPage extends React.Component<Props, State> {
    public subscriptions = new Subscription()
    public refreshRequests = new Subject<void>()

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
                            Add saved search
                        </Link>
                    </div>
                </div>
                {this.state.savedSearches.length > 0 &&
                    this.state.savedSearches
                        .filter(search => search.orgID === this.props.org.id)
                        .map(search => (
                            <Link to={`${this.props.match.path}/${search.id}`}>
                                <div className="org-saved-searches-list-page__row">
                                    <MessageTextOutlineIcon className="org-saved-searches-list-page__row--icon icon-inline" />
                                    <div>{search.description}</div>
                                </div>
                            </Link>
                        ))}
            </div>
        )
    }
}
