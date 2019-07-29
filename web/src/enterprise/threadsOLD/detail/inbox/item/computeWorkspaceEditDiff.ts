import { WorkspaceEdit } from '../../../../../../../shared/src/api/types/workspaceEdit'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { computeDiffFromEdits } from '../../changes/computeDiff'

export async function computeWorkspaceEditDiff(
    extensionsController: ExtensionsControllerProps['extensionsController'],
    workspaceEdit: WorkspaceEdit
): Promise<{ diff: string }> {
    const fileDiffs = await computeDiffFromEdits(extensionsController, [workspaceEdit])
    return { diff: fileDiffs.map(d => d.hunks.map(h => h.body).join('\n')).join('\n\n') }
}
