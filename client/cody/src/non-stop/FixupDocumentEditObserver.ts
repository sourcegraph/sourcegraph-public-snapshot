import * as vscode from 'vscode'

import { FixupFileCollection, FixupTextChanged } from './roles'
import { TextChange, updateRangeMultipleChanges } from './tracked-range'

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
        // Notify which tasks have changed text or the range edits apply to
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
            const updatedRange = updateRangeMultipleChanges(
                task.selectionRange,
                new Array<TextChange>(...event.contentChanges)
            )
            if (!updatedRange.isEqual(task.selectionRange)) {
                task.selectionRange = updatedRange
                this.provider_.rangeDidChange(task)
            }
        }
    }
}
