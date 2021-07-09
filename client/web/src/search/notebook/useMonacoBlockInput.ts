import * as Monaco from 'monaco-editor'
import { useState, useEffect } from 'react'

import { BlockProps } from '.'

export const MONACO_BLOCK_INPUT_OPTIONS: Monaco.editor.IStandaloneEditorConstructionOptions = {
    readOnly: false,
    lineNumbers: 'off',
    lineHeight: 16,
    // Match the query input's height for suggestion items line height.
    suggestLineHeight: 34,
    minimap: {
        enabled: false,
    },
    scrollbar: {
        vertical: 'hidden',
        horizontal: 'hidden',
    },
    glyphMargin: false,
    hover: { delay: 150 },
    lineDecorationsWidth: 0,
    lineNumbersMinChars: 0,
    overviewRulerBorder: false,
    folding: false,
    rulers: [],
    overviewRulerLanes: 0,
    wordBasedSuggestions: false,
    quickSuggestions: false,
    fixedOverflowWidgets: true,
    contextmenu: false,
    links: false,
    // Match our monospace/code style from code.scss
    fontFamily: 'sfmono-regular, consolas, menlo, dejavu sans mono, monospace',
    // Display the cursor as a 1px line.
    cursorStyle: 'line',
    cursorWidth: 1,
    automaticLayout: true,
    renderLineHighlight: 'none',
    wordWrap: 'on',
}

type UseMonacoBlockEditorOptions = { editor: Monaco.editor.IStandaloneCodeEditor | undefined; id: string } & Pick<
    BlockProps,
    'onRunBlock' | 'onBlockInputChange' | 'onSelectBlock' | 'onMoveBlockSelection'
>

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
