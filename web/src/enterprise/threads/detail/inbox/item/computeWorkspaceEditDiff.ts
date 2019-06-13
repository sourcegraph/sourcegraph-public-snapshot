import { WorkspaceEdit } from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { computeDiff } from '../../changes/computeDiff'

export async function computeWorkspaceEditDiff(
    extensionsController: ExtensionsControllerProps['extensionsController'],
    workspaceEdit: WorkspaceEdit
): Promise<{ diff: string }> {
    const fileDiffs = await computeDiff(extensionsController, [{ edit: workspaceEdit }])
    return { diff: fileDiffs.map(d => d.hunks.map(h => h.body).join('\n')).join('\n\n') }
}
