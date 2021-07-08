import classNames from 'classnames'
import * as Monaco from 'monaco-editor'
import React, { useState, useCallback } from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { MonacoEditor } from '@sourcegraph/web/src/components/MonacoEditor'

import blockStyles from './SearchNotebookBlock.module.scss'
import styles from './SearchNotebookMarkdownBlock.module.scss'
import { useMonacoBlockInput } from './useBlockMonacoEditor'

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
    isSelected,
    isLightTheme,
    onRunBlock,
    onBlockInputChange,
    onSelectBlock,
    onMoveBlockSelection,
}) => {
    const [isEditing, setIsEditing] = useState(false)
    const [, setMonacoInstance] = useState<typeof Monaco>()
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()

    const runBlock = useCallback(
        (id: string) => {
            onRunBlock(id)
            setIsEditing(false)
        },
        [onRunBlock, setIsEditing]
    )

    const { isInputFocused } = useMonacoBlockInput({
        editor,
        id,
        onRunBlock: runBlock,
        onBlockInputChange,
        onSelectBlock,
        onMoveBlockSelection,
    })

    const onDoubleClick = useCallback(() => {
        if (!isEditing) {
            setIsEditing(true)
            onSelectBlock(id)
        }
    }, [id, isEditing, setIsEditing, onSelectBlock])

    const onSelect = useCallback(() => {
        onSelectBlock(id)
    }, [id, onSelectBlock])

    if (!isEditing) {
        return (
            <div
                className={classNames(blockStyles.block, isSelected && blockStyles.selected, styles.outputWrapper)}
                onClick={onSelect}
                onDoubleClick={onDoubleClick}
                role="presentation"
                data-block-id={id}
            >
                <div className={styles.output}>
                    <Markdown dangerousInnerHTML={output ?? ''} />
                </div>
            </div>
        )
    }
    return (
        <div
            className={classNames(
                blockStyles.block,
                styles.input,
                isSelected && !isInputFocused && blockStyles.selected,
                isSelected && isInputFocused && blockStyles.selectedNotFocused
            )}
            onClick={onSelect}
            role="presentation"
            data-block-id={id}
        >
            <div className={classNames(blockStyles.monacoWrapper, isInputFocused && blockStyles.selected)}>
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
