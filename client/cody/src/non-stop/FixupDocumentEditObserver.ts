import * as vscode from 'vscode'

import { Edit, Position, Range } from './diff'
import { FixupFileCollection, FixupTextChanged } from './roles'
import { TextChange, updateRangeMultipleChanges } from './tracked-range'

// This does some thunking to manage the two range types: diff ranges, and
// text change ranges.
function updateDiffRange(range: Range, changes: TextChange[]): Range {
    return toDiffRange(updateRangeMultipleChanges(toVsCodeRange(range), changes))
}

function toDiffRange(range: vscode.Range): Range {
    return {
        start: toDiffPosition(range.start),
        end: toDiffPosition(range.end),
    }
}

function toDiffPosition(position: vscode.Position): Position {
    return { line: position.line, character: position.character }
}

function toVsCodeRange(range: Range): vscode.Range {
    return new vscode.Range(toVsCodePosition(range.start), toVsCodePosition(range.end))
}

function toVsCodePosition(position: Position): vscode.Position {
    return new vscode.Position(position.line, position.character)
}

// Updates the ranges in a diff.
function updateRanges(ranges: Range[], changes: TextChange[]): void {
    for (let i = 0; i < ranges.length; i++) {
        ranges[i] = updateDiffRange(ranges[i], changes)
    }
}

// Updates the range in an edit.
function updateEdits(edits: Edit[], changes: TextChange[]): void {
    for (const [i, edit] of edits.entries()) {
        edits[i].range = updateDiffRange(edit.range, changes)
    }
}

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
            const changes = new Array<TextChange>(...event.contentChanges)
            const updatedRange = updateRangeMultipleChanges(task.selectionRange, changes)
            if (task.diff) {
                updateRanges(task.diff.conflicts, changes)
                updateEdits(task.diff.edits, changes)
                updateRanges(task.diff.highlights, changes)
                // Note, we may not notify the decorator of range changes here
                // if the gross range has not changed. That is OK because
                // VScode moves decorations and we can reproduce them lazily.
            }
            if (!updatedRange.isEqual(task.selectionRange)) {
                task.selectionRange = updatedRange
                this.provider_.rangeDidChange(task)
            }
        }
    }
}
