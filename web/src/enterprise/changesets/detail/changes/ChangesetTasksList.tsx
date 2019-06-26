import H from 'history'
import React, { useEffect } from 'react'
import { TextModel } from '../../../../../../shared/src/api/client/services/modelService'
import { WorkspaceRootWithMetadata } from '../../../../../../shared/src/api/client/services/workspaceService'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { getModeFromPath } from '../../../../../../shared/src/languages'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { makeRepoURI } from '../../../../../../shared/src/util/url'
import { TasksList } from '../../../tasks/list/TasksList'
import { ThreadSettings } from '../../../threads/settings'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    thread: GQL.IDiscussionThread
    xchangeset: GQL.IChangeset
    threadSettings: ThreadSettings

    className?: string
    itemClassName?: string
    location: H.Location
    history: H.History
    isLightTheme: boolean
}

/**
 * A list of tasks that apply to a changeset.
 */
export const ChangesetTasksList: React.FunctionComponent<Props> = ({
    xchangeset,
    extensionsController,
    className = '',
    itemClassName = '',
    ...props
}) => {
    const makeURI = (repo: GQL.IRepository, revspec: GQL.IGitRevSpecExpr, filePath?: string) =>
        makeRepoURI({
            repoName: repo.name,
            rev: revspec.object ? revspec.object.oid : revspec.expr,
            filePath,
        })

    useEffect(() => {
        extensionsController.services.workspace.roots.next(
            xchangeset.repositoryComparisons.map<WorkspaceRootWithMetadata>(c => ({
                uri: makeURI(c.headRepository, c.range.headRevSpec),
                baseUri: makeURI(c.baseRepository, c.range.baseRevSpec),
                inputRevision: c.range.headRevSpec.expr,
            }))
        )
        return () => extensionsController.services.workspace.roots.next([])
    }, [extensionsController.services.workspace.roots, xchangeset.repositoryComparisons])

    useEffect(() => {
        const models = xchangeset.repositoryComparisons.flatMap(c =>
            c.fileDiffs.nodes.map<TextModel>(f => ({
                uri: makeURI(c.headRepository, c.range.headRevSpec, f.newFile!.path),
                languageId: getModeFromPath(f.newFile!.path),
                text: f.newFile!.content,
            }))
        )
        for (const model of models) {
            if (!extensionsController.services.model.hasModel(model.uri)) {
                extensionsController.services.model.addModel(model)
            }
        }
        return () => {
            for (const model of models) {
                extensionsController.services.model.removeModel(model.uri)
            }
        }
    }, [
        extensionsController.services.model,
        extensionsController.services.workspace.roots,
        xchangeset.repositoryComparisons,
    ])

    return (
        <div className={`changeset-tasks-list ${className}`}>
            <TasksList {...props} extensionsController={extensionsController} itemClassName={itemClassName} />
        </div>
    )
}
