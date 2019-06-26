import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { GitCommitNode } from '../../../../repo/commits/GitCommitNode'
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
        <ul className="list-group mb-4">
            {xchangeset.commits.map((commit, i) => (
                <li key={i} className="list-group-item p-0">
                    <GitCommitNode
                        repoName={commit.repository.name}
                        node={commit}
                        compact={true}
                        showRepository={true}
                    />
                </li>
            ))}
        </ul>
    </div>
)
