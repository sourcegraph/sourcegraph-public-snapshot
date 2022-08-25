import React, { useContext } from 'react'

import { EditorView } from '@codemirror/view'

export interface CodeMirrorContext {
    setCodeMirrorBlobEditor: (editor: EditorView | undefined) => void
    triggerSearchIfAvailable: () => boolean
}

export const CodeMirrorContext = React.createContext<CodeMirrorContext>({
    setCodeMirrorBlobEditor: () => undefined,
    triggerSearchIfAvailable: () => false,
})

export const useCodeMirrorContext = (): CodeMirrorContext => useContext(CodeMirrorContext)
