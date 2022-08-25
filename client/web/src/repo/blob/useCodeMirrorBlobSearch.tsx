import { useCallback, useRef, useState } from 'react'

import { openSearchPanel } from '@codemirror/search'
import { EditorView } from '@codemirror/view'
import { Shortcut } from '@slimsag/react-shortcuts'

import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'

import { useExperimentalFeatures } from '../../stores'

export interface UseCodeMirrorBlobSearch {
    searchShortcut: React.ReactNode
    isCodeMirrorAvailable: boolean
    // Returns true if blob search was triggered
    triggerSearchIfAvailable: () => boolean
    setCodeMirrorBlobEditor: (editor: EditorView | undefined) => void
}

/**
 * Returns everything you need to trigger search in CodeMirrorBlob:
 *
 * searchShortcut - to be rendered in Layout
 * triggerSeachIfAvailable - for manual triggering in any callback
 * isCodeMirrorAvailable - flag indicating if codeMirror is available
 * setCodeMirrorEditor - to be used by CodeMirrorBlob
 *
 * @returns
 */
export function useCodeMirrorBlobSearch(): UseCodeMirrorBlobSearch {
    const enableCodeMirror = useExperimentalFeatures(features => features.enableCodeMirrorFileView ?? false)

    const searchInFileShortcut = useKeyboardShortcut('searchCodeMirrorBlob')

    const [codeMirrorAvailable, setCodeMirrorAvailable] = useState(false)
    const codeMirrorEditorRef = useRef<EditorView | undefined>()
    const setCodeMirrorBlobEditor = (editor: EditorView | undefined): void => {
        codeMirrorEditorRef.current = editor
        setCodeMirrorAvailable(Boolean(editor))
    }

    const searchCallback = useCallback(() => {
        if (codeMirrorEditorRef.current) {
            openSearchPanel(codeMirrorEditorRef.current)
            return true
        }
        return false
    }, [])

    return {
        searchShortcut:
            (enableCodeMirror &&
                codeMirrorAvailable &&
                searchInFileShortcut?.keybindings.map(keybinding => (
                    <Shortcut key="shortcut-searchCodeMirrorBlob" {...keybinding} onMatch={searchCallback} />
                ))) ??
            null,
        isCodeMirrorAvailable: codeMirrorAvailable,
        triggerSearchIfAvailable: searchCallback,
        setCodeMirrorBlobEditor,
    }
}
