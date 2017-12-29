import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { fetchAllOrgs } from './backend'
import { SettingsInfo } from './util/SettingsInfo'

interface OrgListItemProps {
    className: string

    /**
     * The org to display in this list item.
     */
    org: GQL.IOrg

    /**
     * Called when the org is updated by an action in this list item.
     */
    onDidUpdate?: () => void
}

interface OrgListItemState {
    loading: boolean
    errorDescription?: string
}

class OrgListItem extends React.PureComponent<OrgListItemProps, OrgListItemState> {
    public state: OrgListItemState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        return (
            <li className={this.props.className}>
                <div className="site-admin-detail-list__header">
                    <span className="site-admin-detail-list__name">{this.props.org.name}</span>
                    <br />
                    <span className="site-admin-detail-list__display-name">{this.props.org.displayName}</span>
                </div>
                <ul className="site-admin-detail-list__info">
                    {this.props.org.id && <li>ID: {this.props.org.id}</li>}
                    {this.props.org.createdAt && <li>Created: {format(this.props.org.createdAt, 'YYYY-MM-DD')}</li>}
                    {this.props.org.members &&
                        this.props.org.members.length > 0 && (
                            <li>
                                Members:{' '}
                                <span title={this.props.org.members.map(m => m.user.username).join(', ')}>
                                    {this.props.org.members.length} {pluralize('user', this.props.org.members.length)}
                                </span>
                            </li>
                        )}
                    {this.props.org.latestSettings && (
                        <li>
                            <SettingsInfo
                                settings={this.props.org.latestSettings}
                                filename={`this.props.org-settings-${this.props.org.id}.json`}
                            />
                        </li>
                    )}
                    {this.props.org.tags &&
                        this.props.org.tags.length > 0 && (
                            <li>Tags: {this.props.org.tags.map(tag => tag.name).join(', ')}</li>
                        )}
                </ul>
                <div className="site-admin-detail-list__actions">
                    {this.state.errorDescription && (
                        <p className="site-admin-detail-list__error">{this.state.errorDescription}</p>
                    )}
                </div>
            </li>
        )
    }
}

interface Props extends RouteComponentProps<any> {}

export interface State {
    orgs?: GQL.IOrg[]
}

/**
 * A page displaying the orgs on this site.
 */
export class SiteAdminOrgsPage extends React.Component<Props, State> {
    public state: State = {}

    private orgUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminOrgs')

        this.subscriptions.add(this.orgUpdates.pipe(mergeMap(fetchAllOrgs)).subscribe(orgs => this.setState({ orgs })))
        this.orgUpdates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-detail-list site-admin-orgs-page">
                <PageTitle title="Organizations - Admin" />
                <h2>Organizations</h2>
                <p>
                    See{' '}
                    <a href="https://about.sourcegraph.com/docs/server/config/organizations">
                        Sourcegraph documentation
                    </a>{' '}
                    for information about configuring organizations.
                </p>
                <p>
                    <Link to="/settings/orgs/new" className="btn btn-primary btn-sm">
                        Create new organization
                    </Link>
                </p>
                <ul className="site-admin-detail-list__list">
                    {this.state.orgs &&
                        this.state.orgs.map(org => (
                            <OrgListItem
                                key={org.id}
                                className="site-admin-detail-list__item"
                                org={org}
                                onDidUpdate={this.onDidUpdateOrg}
                            />
                        ))}
                </ul>
                {this.state.orgs && (
                    <p>
                        <small>
                            {this.state.orgs.length} {pluralize('organization', this.state.orgs.length)} total
                        </small>
                    </p>
                )}
            </div>
        )
    }

    private onDidUpdateOrg = () => this.orgUpdates.next()
}
