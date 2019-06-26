import H from 'history'
import React from 'react'
import { RepositoryIcon } from '../../../../../../shared/src/components/icons'
import { RepoLink } from '../../../../../../shared/src/components/RepoLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { GitCommitNode } from '../../../../repo/commits/GitCommitNode'
import { DiffStat } from '../../../../repo/compare/DiffStat'
import { RepositoryCompareCommitsPage } from '../../../../repo/compare/RepositoryCompareCommitsPage'
import { ThreadSettings } from '../../../threads/settings'
interface Props extends ExtensionsControllerProps, PlatformContextProps {
    thread: GQL.IDiscussionThread
    xchangeset: GQL.IChangeset
    threadSettings: ThreadSettings

    location: H.Location
    history: H.History
    isLightTheme: boolean
}

/**
 * A list of affected repositories in a changeset.
 */
export const ChangesetRepositoriesList: React.FunctionComponent<Props> = ({ xchangeset }) => (
    <div className="changeset-repositories-list">
        <ul className="list-group mb-4">
            {xchangeset.repositoryComparisons.map((c, i) => (
                <li key={i} className="list-group-item d-flex justify-content-between">
                    <RepoLink
                        key={c.baseRepository.id}
                        repoName={c.baseRepository.name}
                        to={c.baseRepository.url}
                        icon={RepositoryIcon}
                    />
                    <DiffStat {...c.fileDiffs.diffStat} />
                </li>
            ))}
        </ul>
    </div>
)
