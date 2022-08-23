import React, { useContext } from 'react'

import { EditorView } from '@codemirror/view'

export interface CodeMirrorContext {
    setCodeMirrorBlobEditor: (editor: EditorView | undefined) => void
    triggerSearchIfAvailable: () => void
}

export const CodeMirrorContext = React.createContext<CodeMirrorContext>({
    setCodeMirrorBlobEditor: () => undefined,
    triggerSearchIfAvailable: () => undefined,
})

export const useCodeMirrorContext = (): CodeMirrorContext => useContext(CodeMirrorContext)
