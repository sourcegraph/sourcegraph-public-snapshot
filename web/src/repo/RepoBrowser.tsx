import GearIcon from '@sourcegraph/icons/lib/Gear'
import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { RepoBreadcrumb } from '../components/Breadcrumb'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { fetchAllRepositoriesAndPollIfAnyCloning } from '../site-admin/backend'
import { eventLogger } from '../tracking/eventLogger'

interface RepositoryNodeProps {
    node: GQL.IRepository
}

export const RepositoryNode: React.SFC<RepositoryNodeProps> = ({ node: repo }) => (
    <li key={repo.id} className="repo-browser__item">
        <div className="repo-browser__item-header">
            <Link to={`/${repo.uri}`} className="repo-browser__item-path">
                <RepoBreadcrumb repoPath={repo.uri} disableLinks={true} />
            </Link>
            {repo.mirrorInfo.cloneInProgress && (
                <span className="repo-browser__item-cloning">
                    <small>
                        <Loader className="icon-inline" /> Cloning
                    </small>
                </span>
            )}
        </div>
        <div className="repo-browser__item-spacer" />
        <div className="repo-browser__item-actions">
            {repo.viewerCanAdminister && (
                <Link
                    to={`/${repo.uri}/-/settings`}
                    className="btn btn-secondary btn-sm repo-browser__item-action"
                    data-tooltip="Repository settings"
                >
                    <GearIcon className="icon-inline" />
                </Link>
            )}
            <Link
                to={`/${repo.uri}`}
                className="btn btn-secondary btn-sm repo-browser__item-action"
                data-tooltip="Search and explore this repository"
            >
                <RepoIcon className="icon-inline" />&nbsp;View
            </Link>
        </div>
    </li>
)

interface RepoBrowserProps extends RouteComponentProps<any> {
    user: GQL.IUser | null
}

class FilteredRepositoryConnection extends FilteredConnection<GQL.IRepository> {}

export class RepoBrowser extends React.PureComponent<RepoBrowserProps> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('Browse')
    }

    public render(): JSX.Element | null {
        return (
            <div className="repo-browser">
                <PageTitle title="Repositories" />
                <div className="repo-browser__header">
                    <h2>Repositories</h2>
                    {this.props.user &&
                        this.props.user.siteAdmin && (
                            <div className="repo-browser__actions">
                                <Link
                                    to="/site-admin/configuration"
                                    title="Site admin only"
                                    className="btn btn-secondary btn-sm site-admin-page__actions-btn"
                                >
                                    <GearIcon className="icon-inline" /> Configure repositories
                                </Link>
                            </div>
                        )}
                </div>
                <FilteredRepositoryConnection
                    className="repo-browser__filtered-connection"
                    listClassName="repo-browser__items"
                    noun="repository"
                    pluralNoun="repositories"
                    queryConnection={fetchAllRepositoriesAndPollIfAnyCloning}
                    nodeComponent={RepositoryNode}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }
}
