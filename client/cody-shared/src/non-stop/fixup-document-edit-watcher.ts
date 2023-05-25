import * as vscode from 'vscode'

// A "live" range of lines in an input file. FixupDocumentEditWatcher can update
// FixupLineRanges as the document is edited.
export class FixupLineRange {
    // Creates a FixupLineRange over the specified lines.
    constructor(public start: number, public end: number) {}

    // If set, called when text is edited in the range. The range will be
    // updated with a new start and end line number before the callback is
    // invoked.
    public onRangeEdited: ((document: vscode.TextDocument, range: FixupLineRange) => void) | undefined
}

// Fixups must track ranges of interest within documents that are being worked
// on. Ranges of interest include the region of text we sent to the LLM, and the
// and the decorations indicating where edits will appear.
export class FixupDocumentEditWatcher implements vscode.Disposable {
    private subscription_: vscode.Disposable

    constructor() {
        this.subscription_ = vscode.workspace.onDidChangeTextDocument(this.textDocumentChanged.bind(this))
    }

    public dispose(): void {
        this.subscription_.dispose()
    }

    // TODO: register a range to track
    // TODO: de-register a range to track

    // TODO: watch documents get closed, renamed

    private textDocumentChanged(event: vscode.TextDocumentChangeEvent): void {
        console.log('NYI')
        // TODO: track the registered ranges
        /*
        if (this.batch) {
            let range: vscode.Range | null = this.batch.range
            for (const change of event.contentChanges) {
                range = updateRange<vscode.Range, vscode.Position>(range, change)
                if (!range) {
                    break
                }
            }
            if (range) {
                this.batch.range = range
                this.updateDebugDecorations()
            } else {
                // TODO: Handle the case where the place Cody is editing has been obliterated.
                this.batch = undefined
            }
        }

        let decorations = this.decorations.get(event.document.uri)
        if (decorations) {
            const decorationsToDelete: TrackedDecoration[] = []
            for (const decoration of decorations) {
                for (const change of event.contentChanges) {
                    const updatedRange = updateRange<vscode.Range, vscode.Position>(decoration.range, change)
                    if (updatedRange) {
                        decoration.range = updatedRange
                    } else {
                        decorationsToDelete.push(decoration)
                    }
                }
                if (decorationsToDelete) {
                    decorations = decorations.filter(decoration => !decorationsToDelete.includes(decoration))
                    this.decorations.set(event.document.uri, decorations)
                }
            }
        }
        */
    }
}
