import Loader from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { pluralize } from '../util/strings'

/**
 * A global alert telling the site admin that some repositories are currently being
 * cloned.
 */
export const RepositoriesCloningAlert: React.SFC<{
    repositoriesCloning: GQL.IRepositoryConnection
}> = ({ repositoriesCloning: repositories }) => {
    // Only one line of text is shown, so don't bother actually rendering all of the repos if there are a lot.
    const showRepositories = repositories.nodes.slice(0, 15).map(({ uri }) => uri)

    return (
        <div className="alert alert-success site-alert repositories-cloning-alert">
            <Link className="site-alert__link repositories-cloning-alert__link" to="/site-admin/repositories">
                <Loader className="icon-inline site-alert__link-icon" />{' '}
                <span className="underline">
                    Cloning {repositories.totalCount}{' '}
                    {pluralize('repository', repositories.totalCount!, 'repositories')}...
                </span>
            </Link>
            <div className="repositories-cloning-alert__repositories" title={showRepositories.join('\n')}>
                {showRepositories.map(uri => (
                    <Link key={uri} title={uri} className="repositories-cloning-alert__repository" to={`/${uri}`}>
                        {uri}
                    </Link>
                ))}
            </div>
        </div>
    )
}
