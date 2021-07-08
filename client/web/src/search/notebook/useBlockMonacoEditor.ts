import * as Monaco from 'monaco-editor'
import { useState, useEffect } from 'react'

import { BlockProps } from '.'

type UseMonacoBlockEditorOptions = { editor: Monaco.editor.IStandaloneCodeEditor | undefined; id: string } & Omit<
    BlockProps,
    'isSelected'
>

function blurActiveElement(): void {
    if (document.activeElement instanceof HTMLElement) {
        document.activeElement.blur()
    }
}

export const useMonacoBlockInput = ({
    editor,
    id,
    onRunBlock,
    onBlockInputChange,
    onSelectBlock,
    onMoveBlockSelection,
}: UseMonacoBlockEditorOptions): {
    isInputFocused: boolean
} => {
    const [isInputFocused, setIsInputFocused] = useState(false)

    useEffect(() => {
        if (!editor) {
            return
        }
        const addRunBlockActionDisposable = editor.addAction({
            id: 'run-block-on-cmd-enter',
            label: 'Run block',
            keybindings: [Monaco.KeyMod.CtrlCmd | Monaco.KeyCode.Enter],
            run: () => onRunBlock(id),
        })
        const moveUpOnFirstLineDisposable = editor.addAction({
            id: 'move-up-on-first-line',
            label: 'Move up on first line',
            keybindings: [Monaco.KeyCode.UpArrow],
            run: editor => {
                const position = editor.getPosition()
                if (!position) {
                    return
                }
                if (position.lineNumber === 1) {
                    onMoveBlockSelection(id, 'up')
                    blurActiveElement()
                } else {
                    editor.setPosition({ lineNumber: position.lineNumber - 1, column: position.column })
                }
            },
        })
        const moveDownOnLastLineDisposable = editor.addAction({
            id: 'move-down-on-last-line',
            label: 'Move down on last line',
            keybindings: [Monaco.KeyCode.DownArrow],
            run: editor => {
                const position = editor.getPosition()
                if (!position) {
                    return
                }
                const lineCount = editor.getModel()?.getLineCount()
                if (!lineCount) {
                    return
                }
                if (position.lineNumber === lineCount) {
                    onMoveBlockSelection(id, 'down')
                    blurActiveElement()
                } else {
                    editor.setPosition({ lineNumber: position.lineNumber + 1, column: position.column })
                }
            },
        })
        return () => {
            addRunBlockActionDisposable.dispose()
            moveUpOnFirstLineDisposable.dispose()
            moveDownOnLastLineDisposable.dispose()
        }
    }, [editor, id, onRunBlock, onMoveBlockSelection])

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.onDidChangeModelContent(() => {
            onBlockInputChange(id, editor.getValue())
        })
        return () => disposable.dispose()
    }, [editor, id, onBlockInputChange])

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
