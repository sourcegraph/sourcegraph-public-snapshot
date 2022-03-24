import { useEffect } from 'react'

import * as Monaco from 'monaco-editor'

export function focusLastPositionInMonacoEditor(editor: Monaco.editor.IStandaloneCodeEditor | undefined): void {
    if (!editor) {
        return
    }
    // setTimeout executes the editor focus in a separate run-loop which prevents adding a newline at the start of the input,
    // if Enter key was used to show the editor.
    setTimeout(() => {
        const lines = editor.getValue().split('\n')
        editor.setPosition({ column: lines[lines.length - 1].length + 2, lineNumber: lines.length })
        editor.focus()
    }, 0)
}

interface UseFocusMonacoEditorOnMountProps {
    editor: Monaco.editor.IStandaloneCodeEditor | undefined
    isEditing: boolean | undefined
}

export const useFocusMonacoEditorOnMount = ({ editor, isEditing }: UseFocusMonacoEditorOnMountProps): void => {
    useEffect(() => {
        if (isEditing) {
            focusLastPositionInMonacoEditor(editor)
        }
    }, [editor, isEditing])
}
