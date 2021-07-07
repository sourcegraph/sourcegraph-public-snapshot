import * as Monaco from 'monaco-editor'
import React, { useState, useCallback, useEffect } from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { MonacoEditor } from '@sourcegraph/web/src/components/MonacoEditor'

import styles from './SearchNotebookMarkdownBlock.module.scss'

import { BlockProps, MarkdownBlock } from '.'

interface SearchNotebookMarkdownBlockProps extends BlockProps, Omit<MarkdownBlock, 'type'>, ThemeProps {}

const options: Monaco.editor.IStandaloneEditorConstructionOptions = {
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
}

// TODO: Use React.memo
export const SearchNotebookMarkdownBlock: React.FunctionComponent<SearchNotebookMarkdownBlockProps> = ({
    id,
    input,
    output,
    isLightTheme,
    onRunBlock,
    onBlockInputChange,
}) => {
    const [isEditing, setIsEditing] = useState(false)

    const onDoubleClick = useCallback(() => {
        if (!isEditing) {
            setIsEditing(true)
        }
    }, [isEditing, setIsEditing])

    const [, setMonacoInstance] = useState<typeof Monaco>()
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.addAction({
            id: 'render-on-cmd-enter',
            label: 'Render markdown',
            keybindings: [Monaco.KeyMod.CtrlCmd | Monaco.KeyCode.Enter],
            run: () => {
                onRunBlock(id)
                setIsEditing(false)
            },
        })
        return () => disposable.dispose()
    }, [editor, id, setIsEditing, onRunBlock])

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.onDidChangeModelContent(() => {
            onBlockInputChange(id, editor.getValue())
        })
        return () => disposable.dispose()
    }, [editor, id, onBlockInputChange])

    if (!isEditing) {
        return (
            <div className={styles.outputWrapper}>
                <div className={styles.output} onDoubleClick={onDoubleClick}>
                    <Markdown dangerousInnerHTML={output ?? ''} />
                </div>
            </div>
        )
    }
    return (
        <div className={styles.input}>
            <div className={styles.monacoWrapper}>
                <MonacoEditor
                    language="markdown"
                    value={input}
                    height={150}
                    isLightTheme={isLightTheme}
                    editorWillMount={setMonacoInstance}
                    onEditorCreated={setEditor}
                    options={options}
                    border={false}
                />
            </div>
        </div>
    )
}
