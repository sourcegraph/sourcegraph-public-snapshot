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
        vertical: 'auto',
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
    renderLineHighlight: 'none',
    wordWrap: 'on',
}

interface UseMonacoBlockEditorOptions
    extends Pick<BlockProps, 'onRunBlock' | 'onBlockInputChange' | 'onSelectBlock' | 'onMoveBlockSelection'> {
    editor: Monaco.editor.IStandaloneCodeEditor | undefined
    id: string
    type: 'md' | 'query'
}

export const useMonacoBlockInput = ({
    editor,
    id,
    type,
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
        const disposables = [
            editor.addAction({
                id: 'run-block-on-cmd-enter',
                label: 'Run block',
                keybindings: [Monaco.KeyMod.CtrlCmd | Monaco.KeyCode.Enter],
                run: () => onRunBlock(id),
            }),
            editor.addAction({
                id: 'blur-on-esacpe',
                label: 'Blur on escape',
                keybindings: [Monaco.KeyCode.Escape],
                run: () => {
                    if (document.activeElement instanceof HTMLElement) {
                        document.activeElement.blur()
                    }
                },
            }),
        ]
        return () => {
            for (const disposable of disposables) {
                disposable.dispose()
            }
        }
    }, [editor, id, onRunBlock, onMoveBlockSelection])

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.onDidChangeModelContent(() => {
            onBlockInputChange(id, { type, input: editor.getValue() })
        })
        return () => disposable.dispose()
    }, [editor, id, type, onBlockInputChange])

    useEffect(() => {
        if (!editor) {
            setIsInputFocused(false)
            return
        }
        const disposables = [
            editor.onDidFocusEditorText(() => {
                setIsInputFocused(true)
                onSelectBlock(id)
            }),
            editor.onDidBlurEditorText(() => setIsInputFocused(false)),
            editor.onDidDispose(() => setIsInputFocused(false)),
        ]
        return () => {
            for (const disposable of disposables) {
                disposable.dispose()
            }
        }
    }, [editor, id, setIsInputFocused, onSelectBlock])

    return { isInputFocused }
}
