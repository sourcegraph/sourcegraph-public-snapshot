import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { GitCommitNode } from '../../../../repo/commits/GitCommitNode'
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
 * A list of commits in a changeset.
 */
export const ChangesetCommitsList: React.FunctionComponent<Props> = ({
    thread,
    xchangeset,
    threadSettings,
    ...props
}) => (
    <div className="changeset-commits-list">
        {xchangeset.repositoryComparisons.map((c, i) => (
            <ol key={i} className="list-group mb-4">
                {c.commits.nodes.map((commit, i) => (
                    <li key={i} className="list-group-item p-0">
                        <GitCommitNode repoName={c.baseRepository.name} node={commit} />
                    </li>
                ))}
            </ol>
        ))}
    </div>
)
