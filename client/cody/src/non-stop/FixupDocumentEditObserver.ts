import * as vscode from 'vscode'

import { FixupFileCollection, FixupTextChanged } from './roles'

/**
 * Observes text document changes and updates the regions with active fixups.
 * Notifies the fixup controller when text being edited by a fixup changes.
 * Fixups must track ranges of interest within documents that are being worked
 * on. Ranges of interest include the region of text we sent to the LLM, and the
 * and the decorations indicating where edits will appear.
 */
export class FixupDocumentEditObserver {
    constructor(private readonly provider_: FixupFileCollection & FixupTextChanged) {}

    public textDocumentChanged(event: vscode.TextDocumentChangeEvent): void {
        const file = this.provider_.maybeFileForUri(event.document.uri)
        if (!file) {
            return
        }
        const tasks = this.provider_.tasksForFile(file)
        // Notify which tasks have changed text
        for (const task of tasks) {
            for (const edit of event.contentChanges) {
                if (
                    edit.range.end.isBeforeOrEqual(task.selectionRange.start) ||
                    edit.range.start.isAfterOrEqual(task.selectionRange.end)
                ) {
                    continue
                }
                this.provider_.textDidChange(task)
                break
            }
        }
        // TODO: Update the selection ranges which were edited, see
        // ./tracked-range or consider using simpler line-based
        // ranges. Until this is implemented adding/deleting lines
        // before fixups will cause conflicts.
        //
        // Need to clarify whether multi-range edits are progressive in the
        // sense that edit 2 refers to ranges updated after edit 1, or if
        // all the edits refer to ranges in the original document.
        // TODO: Create new code lenses with updated ranges for each task with .set()
    }
}
