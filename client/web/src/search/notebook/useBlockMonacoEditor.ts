import * as Monaco from 'monaco-editor'
import { useState, useEffect } from 'react'

import { BlockProps } from '.'

type UseBlockMonacoEditorOptions = { editor: Monaco.editor.IStandaloneCodeEditor | undefined; id: string } & Omit<
    BlockProps,
    'isSelected'
>

export const useBlockMonacoInput = ({
    editor,
    id,
    onRunBlock,
    onBlockInputChange,
    onSelectBlock,
}: UseBlockMonacoEditorOptions): {
    isInputFocused: boolean
} => {
    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.addAction({
            id: 'run-on-cmd-enter',
            label: 'Run block',
            keybindings: [Monaco.KeyMod.CtrlCmd | Monaco.KeyCode.Enter],
            run: () => {
                onRunBlock(id)
            },
        })
        return () => disposable.dispose()
    }, [editor, id, onRunBlock])

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.onDidChangeModelContent(() => {
            onBlockInputChange(id, editor.getValue())
        })
        return () => disposable.dispose()
    }, [editor, id, onBlockInputChange])

    const [isInputFocused, setIsInputFocused] = useState(false)
    useEffect(() => {
        if (!editor) {
            setIsInputFocused(false)
            return
        }
        const onDidFocusEditorTextDisposable = editor.onDidFocusEditorText(() => {
            setIsInputFocused(true)
            onSelectBlock(id)
        })
        const onDidBlurEditorTextDisposable = editor.onDidBlurEditorText(() => setIsInputFocused(false))
        return () => {
            onDidFocusEditorTextDisposable.dispose()
            onDidBlurEditorTextDisposable.dispose()
        }
    }, [editor, id, setIsInputFocused, onSelectBlock])

    return { isInputFocused }
}
