import AddIcon from 'mdi-react/AddIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import UserIcon from 'mdi-react/UserIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import * as GQL from '../../../shared/src/graphql/schema'
import { pluralize } from '../../../shared/src/util/strings'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { orgURL } from '../org'
import { eventLogger } from '../tracking/eventLogger'
import { deleteOrganization, fetchAllOrganizations } from './backend'
import { ErrorAlert } from '../components/alerts'

interface OrgNodeProps {
    /**
     * The org to display in that list item.
     */
    node: GQL.IOrg

    /**
     * Called when the org is updated by an action in that list item.
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
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <Link to={orgURL(that.props.node.name)}>
                            <strong>{that.props.node.name}</strong>
                        </Link>
                        <br />
                        <span className="text-muted">{that.props.node.displayName}</span>
                    </div>
                    <div>
                        <Link
                            to={`${orgURL(that.props.node.name)}/settings`}
                            className="btn btn-sm btn-secondary"
                            data-tooltip="Organization settings"
                        >
                            <SettingsIcon className="icon-inline" /> Settings
                        </Link>{' '}
                        <Link
                            to={`${orgURL(that.props.node.name)}/members`}
                            className="btn btn-sm btn-secondary"
                            data-tooltip="Organization members"
                        >
                            <UserIcon className="icon-inline" />{' '}
                            {that.props.node.members && (
                                <>
                                    {that.props.node.members.totalCount}{' '}
                                    {pluralize('member', that.props.node.members.totalCount)}
                                </>
                            )}
                        </Link>{' '}
                        <button
                            type="button"
                            className="btn btn-sm btn-danger"
                            onClick={that.deleteOrg}
                            disabled={that.state.loading}
                            data-tooltip="Delete organization"
                        >
                            <DeleteIcon className="icon-inline" />
                        </button>
                    </div>
                </div>
                {that.state.errorDescription && <ErrorAlert className="mt-2" error={that.state.errorDescription} />}
            </li>
        )
    }

    private deleteOrg = (): void => {
        if (!window.confirm(`Delete the organization ${that.props.node.name}?`)) {
            return
        }

        that.setState({
            errorDescription: undefined,
            loading: true,
        })

        deleteOrganization(that.props.node.id)
            .toPromise()
            .then(
                () => {
                    that.setState({ loading: false })
                    if (that.props.onDidUpdate) {
                        that.props.onDidUpdate()
                    }
                },
                err => that.setState({ loading: false, errorDescription: err.message })
            )
    }
}

interface Props extends RouteComponentProps<{}> {}

interface State {
    orgs?: GQL.IOrg[]
    totalCount?: number
}

class FilteredOrgConnection extends FilteredConnection<GQL.IOrg, Pick<OrgNodeProps, 'onDidUpdate'>> {}

/**
 * A page displaying the orgs on that site.
 */
export class SiteAdminOrgsPage extends React.Component<Props, State> {
    public state: State = {}

    private orgUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminOrgs')
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<OrgNodeProps, 'onDidUpdate'> = {
            onDidUpdate: that.onDidUpdateOrg,
        }

        return (
            <div className="site-admin-orgs-page">
                <PageTitle title="Organizations - Admin" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-3">
                    <h2 className="mb-0">Organizations</h2>
                    <Link to="/organizations/new" className="btn btn-primary">
                        <AddIcon className="icon-inline" /> Create organization
                    </Link>
                </div>
                <p>
                    An organization is a set of users with associated configuration. See{' '}
                    <Link to="/help/user/organizations">Sourcegraph documentation</Link> for information about
                    configuring organizations.
                </p>
                <FilteredOrgConnection
                    className="list-group list-group-flush mt-3"
                    noun="organization"
                    pluralNoun="organizations"
                    queryConnection={fetchAllOrganizations}
                    nodeComponent={OrgNode}
                    nodeComponentProps={nodeProps}
                    updates={that.orgUpdates}
                    history={that.props.history}
                    location={that.props.location}
                />
            </div>
        )
    }

    private onDidUpdateOrg = (): void => that.orgUpdates.next()
}
