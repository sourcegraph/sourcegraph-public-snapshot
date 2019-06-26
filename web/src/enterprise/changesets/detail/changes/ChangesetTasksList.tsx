import H from 'history'
import React, { useEffect } from 'react'
import { WorkspaceRootWithMetadata } from '../../../../../../shared/src/api/client/services/workspaceService'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { makeRepoURI } from '../../../../../../shared/src/util/url'
import { TasksList } from '../../../tasks/list/TasksList'
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
 * A list of tasks that apply to a changeset.
 */
export const ChangesetTasksList: React.FunctionComponent<Props> = ({ xchangeset, extensionsController, ...props }) => {
    useEffect(() => {
        extensionsController.services.workspace.roots.next(
            xchangeset.repositoryComparisons.map<WorkspaceRootWithMetadata>(c => ({
                uri: makeRepoURI({
                    repoName: c.baseRepository.name,
                    rev: c.range.headRevSpec.object ? c.range.headRevSpec.object.oid : c.range.headRevSpec.expr,
                }),
                inputRevision: c.range.headRevSpec.expr,
            }))
        )
        return () => extensionsController.services.workspace.roots.next([])
    }, [extensionsController.services.workspace.roots, xchangeset.repositoryComparisons])

    return (
        <div className="changeset-tasks-list">
            <TasksList {...props} extensionsController={extensionsController} />
        </div>
    )
}
