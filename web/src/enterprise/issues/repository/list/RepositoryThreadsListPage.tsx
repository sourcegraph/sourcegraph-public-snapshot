import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { ChangesetsList } from '../../list/ChangesetsList'
import { useChangesets } from '../../list/useChangesets'
import { RepositoryChangesetsAreaContext } from '../RepositoryChangesetsArea'

interface Props extends Pick<RepositoryChangesetsAreaContext, 'repo'>, ExtensionsControllerNotificationProps {
    newChangesetURL: string | null
}

/**
 * Lists a repository's changesets.
 */
export const RepositoryChangesetsListPage: React.FunctionComponent<Props> = ({ newChangesetURL, repo, ...props }) => {
    const changesets = useChangesets(repo)
    return (
        <div className="repository-changesets-list-page">
            {newChangesetURL && (
                <Link to={newChangesetURL} className="btn btn-primary mb-3">
                    New changeset
                </Link>
            )}
            <ChangesetsList {...props} changesets={changesets} />
        </div>
    )
}
