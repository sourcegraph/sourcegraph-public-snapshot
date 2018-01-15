import FolderIcon from '@sourcegraph/icons/lib/Folder'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import Loader from '@sourcegraph/icons/lib/Loader'
import SearchIcon from '@sourcegraph/icons/lib/Search'
import escapeRegexp from 'escape-string-regexp'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { RepoBreadcrumb } from '../components/Breadcrumb'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { buildSearchURLQuery } from '../search'
import { fetchAllRepositoriesAndPollIfAnyCloning } from '../site-admin/backend'
import { eventLogger } from '../tracking/eventLogger'

export const RepositoryNode: React.SFC<{ node: GQL.IRepository }> = ({ node: repo }) => (
    <li key={repo.id} className="repo-browser__item">
        <div className="repo-browser__item-header">
            <Link to={`/${repo.uri}`} className="repo-browser__item-path">
                <RepoBreadcrumb repoPath={repo.uri} disableLinks={true} />
            </Link>
            {repo.cloneInProgress && (
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
                data-tooltip="Explore files in repository"
            >
                <FolderIcon className="icon-inline" />
            </Link>
            <Link
                to={`/search?${buildSearchURLQuery({ query: `repo:^${escapeRegexp(repo.uri)}$ ` })}&focus`}
                className="btn btn-secondary btn-sm repo-browser__item-action"
                data-tooltip="Search in repository"
            >
                <SearchIcon className="icon-inline" />
                &nbsp;Search
            </Link>
        </div>
    </li>
)

interface Props extends RouteComponentProps<any> {
    user: GQL.IUser | null
}

export class RepoBrowser extends React.PureComponent<Props> {
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
                <FilteredConnection
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
