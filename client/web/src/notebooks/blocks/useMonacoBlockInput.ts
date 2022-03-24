import { useEffect } from 'react'

import * as Monaco from 'monaco-editor'

import { BlockProps } from '..'

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

interface UseMonacoBlockEditorOptions extends Pick<BlockProps, 'onRunBlock'> {
    editor: Monaco.editor.IStandaloneCodeEditor | undefined
    id: string
    preventNewLine?: boolean
    tabMovesFocus?: boolean
    onInputChange: (value: string) => void
}

const REPLACE_NEW_LINE_REGEX = /[\n\r↵]/g

export const useMonacoBlockInput = ({
    editor,
    id,
    preventNewLine,
    tabMovesFocus = true,
    onRunBlock,
    onInputChange,
}: UseMonacoBlockEditorOptions): void => {
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

        if (preventNewLine) {
            disposables.push(
                editor.addAction({
                    id: 'preventEnter',
                    label: 'preventEnter',
                    keybindings: [Monaco.KeyCode.Enter],
                    run: () => {
                        editor.trigger('preventEnter', 'acceptSelectedSuggestion', [])
                    },
                })
            )
        }

        return () => {
            for (const disposable of disposables) {
                disposable.dispose()
            }
        }
    }, [editor, id, preventNewLine, onRunBlock])

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.onDidChangeModelContent(() => {
            const value = editor.getValue()
            onInputChange(preventNewLine ? value.replace(REPLACE_NEW_LINE_REGEX, '') : value)
        })
        return () => disposable.dispose()
    }, [editor, id, preventNewLine, onInputChange])

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposables = [
            editor.onDidFocusEditorText(() => {
                if (tabMovesFocus) {
                    editor.createContextKey('editorTabMovesFocus', true)
                }
            }),
        ]
        return () => {
            for (const disposable of disposables) {
                disposable.dispose()
            }
        }
    }, [editor, id, tabMovesFocus])
}
