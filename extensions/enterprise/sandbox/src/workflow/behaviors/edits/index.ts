import * as sourcegraph from 'sourcegraph'
import { Unsubscribable } from 'rxjs'
import { SerializedWorkspaceEdit, WorkspaceEdit } from '../../../../../../../shared/src/api/types/workspaceEdit'
import * as GQL from '../../../../../../../shared/src/graphql/schema'

export interface EditsBehaviorCommandContext {
    title: string
    body?: string
    headBranch: string
}

export interface Changeset {
    title: string
    body?: string
    baseBranch: string
    baseRepository: string
    headBranch: string
    headRepository: string
    edit: sourcegraph.WorkspaceEdit
    sideEffects?: GQL.ISideEffectInput[]
}

export type EditsBehaviorCommand<C extends EditsBehaviorCommandContext> = (
    edit: sourcegraph.WorkspaceEdit,
    context: C
) => Promise<(Omit<Changeset, 'edit'> & { edit: sourcegraph.WorkspaceEdit })[]>

export const registerEditsBehaviorCommand = <C extends EditsBehaviorCommandContext>(
    id: string,
    fn: EditsBehaviorCommand<C>
): Unsubscribable =>
    sourcegraph.commands.registerCommand(id, async (serializedEdit: SerializedWorkspaceEdit, context: any) => {
        const changesets = await fn(WorkspaceEdit.fromJSON(serializedEdit), context)
        return changesets.map(({ edit, ...changeset }) => ({ ...changeset, edit: (edit as WorkspaceEdit).toJSON() }))
    })
