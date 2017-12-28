import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { fetchAllOrgs } from './backend'
import { SettingsInfo } from './util/SettingsInfo'

interface Props extends RouteComponentProps<any> {}

export interface State {
    orgs?: GQL.IOrg[]
}

/**
 * A page displaying the orgs on this site.
 */
export class SiteAdminOrgsPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminOrgs')

        this.subscriptions.add(fetchAllOrgs().subscribe(orgs => this.setState({ orgs })))
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
                            <li key={org.id} className="site-admin-detail-list__item">
                                <div className="site-admin-detail-list__header">
                                    <span className="site-admin-detail-list__name">{org.name}</span>
                                    <br />
                                    <span className="site-admin-detail-list__display-name">{org.displayName}</span>
                                </div>
                                <ul className="site-admin-detail-list__info">
                                    {org.id && <li>ID: {org.id}</li>}
                                    {org.createdAt && <li>Created: {format(org.createdAt, 'YYYY-MM-DD')}</li>}
                                    {org.members &&
                                        org.members.length > 0 && (
                                            <li>
                                                Members:{' '}
                                                <span title={org.members.map(m => m.user.username).join(', ')}>
                                                    {org.members.length} {pluralize('user', org.members.length)}
                                                </span>
                                            </li>
                                        )}
                                    {org.latestSettings && (
                                        <li>
                                            <SettingsInfo
                                                settings={org.latestSettings}
                                                filename={`org-settings-${org.id}.json`}
                                            />
                                        </li>
                                    )}
                                    {org.tags &&
                                        org.tags.length > 0 && (
                                            <li>Tags: {org.tags.map(tag => tag.name).join(', ')}</li>
                                        )}
                                </ul>
                            </li>
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
}
