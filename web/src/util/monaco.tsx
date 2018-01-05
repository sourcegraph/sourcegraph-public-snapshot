import { applyEdits } from '@sqs/jsonc-parser/lib/format'
import { toMonacoEdits } from '../settings/MonacoSettingsEditor'
import { eventLogger } from '../tracking/eventLogger'

export function addEditorAction(
    inputEditor: monaco.editor.IStandaloneCodeEditor,
    model: monaco.editor.IModel,
    label: string,
    id: string,
    run: any
): void {
    inputEditor.addAction({
        label,
        id,
        run: editor => {
            eventLogger.log('SiteConfigurationActionExecuted', { id })
            editor.focus()
            editor.pushUndoStop()
            const { edits, selectText } = run(editor.getValue())
            const monacoEdits = toMonacoEdits(model, edits)
            let selection: monaco.Selection | undefined
            if (typeof selectText === 'string') {
                const afterText = applyEdits(editor.getValue(), edits)
                let offset = afterText.slice(edits[0].offset).indexOf(selectText)
                if (offset !== -1) {
                    offset += edits[0].offset
                    selection = monaco.Selection.fromPositions(
                        getPositionAt(afterText, offset),
                        getPositionAt(afterText, offset + selectText.length)
                    )
                }
            }
            if (!selection) {
                selection = monaco.Selection.fromPositions(
                    monacoEdits[0].range.getStartPosition(),
                    monacoEdits[monacoEdits.length - 1].range.getEndPosition()
                )
            }
            editor.executeEdits(id, monacoEdits, [selection])
            editor.revealPositionInCenter(selection.getStartPosition())
        },
    })
}

function getPositionAt(text: string, offset: number): monaco.Position {
    const lines = text.split('\n')
    let pos = 0
    for (const [i, line] of lines.entries()) {
        if (offset < pos + line.length + 1) {
            return new monaco.Position(i + 1, offset - pos + 1)
        }
        pos += line.length + 1
    }
    throw new Error(`offset ${offset} out of bounds in text of length ${text.length}`)
}
