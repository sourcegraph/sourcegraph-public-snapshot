import GearIcon from '@sourcegraph/icons/lib/Gear'
import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { fetchAllRepositories } from './backend'

export const RepositoryNode: React.SFC<{ node: GQL.IRepository }> = ({ node: repo }) => (
    <li key={repo.id} className="site-admin-detail-list__item site-admin-repositories-page__repo">
        <div className="site-admin-detail-list__header site-admin-repositories-page__repo-header">
            <Link to={`/${repo.uri}`} className="site-admin-detail-list__name">
                {repo.uri}
            </Link>
        </div>
        <ul className="site-admin-detail-list__info site-admin-repositories-page__repo-info">
            <li>ID: {repo.id}</li>
            {repo.createdAt && <li>Created: {format(repo.createdAt, 'YYYY-MM-DD')}</li>}
        </ul>
    </li>
)

interface Props extends RouteComponentProps<any> {}

/**
 * A page displaying the repositories on this site.
 */
export class SiteAdminRepositoriesPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminRepos')
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-detail-list site-admin-repositories-page">
                <PageTitle title="Repositories" />
                <h2>Repositories</h2>
                <div className="site-admin-page__actions">
                    <Link
                        to="/site-admin/configuration"
                        className="btn btn-primary btn-sm site-admin-page__actions-btn"
                    >
                        <GearIcon className="icon-inline" /> Configure repositories
                    </Link>
                </div>
                <FilteredConnection
                    className="site-admin-page__filtered-connection"
                    noun="repository"
                    pluralNoun="repositories"
                    queryConnection={fetchAllRepositories}
                    nodeComponent={RepositoryNode}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }
}
