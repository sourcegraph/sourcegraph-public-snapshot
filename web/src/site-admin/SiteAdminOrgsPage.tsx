import AddIcon from '@sourcegraph/icons/lib/Add'
import DeleteIcon from '@sourcegraph/icons/lib/Delete'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import UserIcon from '@sourcegraph/icons/lib/User'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { orgURL } from '../org'
import { eventLogger } from '../tracking/eventLogger'
import { deleteOrganization, fetchAllOrgs } from './backend'

interface OrgNodeProps {
    /**
     * The org to display in this list item.
     */
    node: GQL.IOrg

    /**
     * Called when the org is updated by an action in this list item.
     */
    onDidUpdate?: () => void
}

interface OrgNodeState {
    loading: boolean
    errorDescription?: string
}

class OrgNode extends React.PureComponent<OrgNodeProps, OrgNodeState> {
    public state: OrgNodeState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        return (
            <li className="site-admin-detail-list__item">
                <div className="site-admin-detail-list__header">
                    <Link
                        to={`${orgURL(this.props.node.name)}/settings/members`}
                        className="site-admin-detail-list__name"
                    >
                        {this.props.node.name}
                    </Link>
                    <br />
                    <span className="site-admin-detail-list__display-name">{this.props.node.displayName}</span>
                </div>
                <div className="site-admin-detail-list__actions">
                    <button
                        className="btn btn-outline-danger site-admin-detail-list__action"
                        onClick={this.deleteOrg}
                        disabled={this.state.loading}
                        data-tooltip="Delete organization"
                    >
                        <DeleteIcon className="icon-inline" />
                    </button>
                    <Link
                        to={`${orgURL(this.props.node.name)}/settings/members`}
                        className="btn btn-secondary site-admin-detail-list__action"
                        data-tooltip="Organization members"
                    >
                        <UserIcon className="icon-inline" />{' '}
                        {this.props.node.memberships && this.props.node.memberships.length}
                    </Link>
                    <Link
                        to={`${orgURL(this.props.node.name)}/settings`}
                        className="btn btn-secondary site-admin-detail-list__action"
                        data-tooltip="Organization settings"
                    >
                        <GearIcon className="icon-inline" />
                    </Link>
                    {this.state.errorDescription && (
                        <p className="site-admin-detail-list__error">{this.state.errorDescription}</p>
                    )}
                </div>
            </li>
        )
    }

    private deleteOrg = () => {
        if (!window.confirm(`Really delete the organization ${this.props.node.name}?`)) {
            return
        }

        this.setState({
            errorDescription: undefined,
            loading: true,
        })

        deleteOrganization(this.props.node.id)
            .toPromise()
            .then(
                () => {
                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                err => this.setState({ loading: false, errorDescription: err.message })
            )
    }
}

interface Props extends RouteComponentProps<any> {}

export interface State {
    orgs?: GQL.IOrg[]
    totalCount?: number
}

class FilteredOrgConnection extends FilteredConnection<GQL.IOrg, Pick<OrgNodeProps, 'onDidUpdate'>> {}

/**
 * A page displaying the orgs on this site.
 */
export class SiteAdminOrgsPage extends React.Component<Props, State> {
    public state: State = {}

    private orgUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminOrgs')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<OrgNodeProps, 'onDidUpdate'> = {
            onDidUpdate: this.onDidUpdateOrg,
        }

        return (
            <div className="site-admin-detail-list site-admin-orgs-page">
                <PageTitle title="Organizations - Admin" />
                <h2>Organizations</h2>
                <div className="site-admin-page__actions">
                    <Link to="/organizations/new" className="btn btn-primary btn-sm site-admin-page__actions-btn">
                        <AddIcon className="icon-inline" /> Create organization
                    </Link>
                </div>
                <FilteredOrgConnection
                    className="site-admin-page__filtered-connection"
                    noun="organization"
                    pluralNoun="organizations"
                    queryConnection={fetchAllOrgs}
                    nodeComponent={OrgNode}
                    nodeComponentProps={nodeProps}
                    updates={this.orgUpdates}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private onDidUpdateOrg = () => this.orgUpdates.next()
}
