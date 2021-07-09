import classNames from 'classnames'
import * as Monaco from 'monaco-editor'
import React, { useState, useCallback, useRef, useEffect } from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { MonacoEditor } from '@sourcegraph/web/src/components/MonacoEditor'

import blockStyles from './SearchNotebookBlock.module.scss'
import styles from './SearchNotebookMarkdownBlock.module.scss'
import { useBlockFocusHandlers } from './useBlockFocusHandlers'
import { useBlockShortcutHandlers } from './useBlockShortcutHandlers'
import { MONACO_BLOCK_INPUT_OPTIONS, useMonacoBlockInput } from './useMonacoBlockInput'

import { BlockProps, MarkdownBlock } from '.'

interface SearchNotebookMarkdownBlockProps extends BlockProps, Omit<MarkdownBlock, 'type'>, ThemeProps {
    isMacPlatform: boolean
}

// TODO: Use React.memo
export const SearchNotebookMarkdownBlock: React.FunctionComponent<SearchNotebookMarkdownBlockProps> = ({
    id,
    input,
    output,
    isSelected,
    isLightTheme,
    isMacPlatform,
    onRunBlock,
    onBlockInputChange,
    onSelectBlock,
    onMoveBlockSelection,
    onDeleteBlock,
    onMoveBlock,
    onDuplicateBlock,
}) => {
    const [isEditing, setIsEditing] = useState(false)
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()
    const blockElement = useRef<HTMLDivElement>(null)

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

    const onSelect = useCallback(() => onSelectBlock(id), [id, onSelectBlock])
    const onEnterBlock = useCallback(() => setIsEditing(true), [setIsEditing])

    const { onBlur } = useBlockFocusHandlers({ blockElement: blockElement.current, onSelectBlock, isSelected })
    const { onKeyDown } = useBlockShortcutHandlers({
        id,
        isMacPlatform,
        onMoveBlockSelection,
        onEnterBlock,
        onDeleteBlock,
        onRunBlock: runBlock,
        onMoveBlock,
        onDuplicateBlock,
    })

    useEffect(() => {
        if (isEditing) {
            editor?.focus()
        } else {
            blockElement.current?.focus()
        }
    }, [isEditing, editor])

    if (!isEditing) {
        return (
            // eslint-disable-next-line jsx-a11y/no-static-element-interactions
            <div
                className={classNames(blockStyles.block, isSelected && blockStyles.selected, styles.outputWrapper)}
                onClick={onSelect}
                onDoubleClick={onDoubleClick}
                onFocus={onSelect}
                onBlur={onBlur}
                onKeyDown={onKeyDown}
                // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                tabIndex={0}
                // eslint-disable-next-line jsx-a11y/aria-role
                role="notebook-block"
                aria-label="Notebook block"
                data-block-id={id}
                ref={blockElement}
            >
                <div className={styles.output}>
                    <Markdown dangerousInnerHTML={output ?? ''} />
                </div>
            </div>
        )
    }

    return (
        // eslint-disable-next-line jsx-a11y/no-static-element-interactions
        <div
            className={classNames(blockStyles.block, styles.input, isInputFocused && blockStyles.selected)}
            onClick={onSelect}
            onFocus={onSelect}
            onBlur={onBlur}
            onKeyDown={onKeyDown}
            // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
            tabIndex={0}
            // eslint-disable-next-line jsx-a11y/aria-role
            role="notebook-block"
            aria-label="Notebook block"
            data-block-id={id}
            ref={blockElement}
        >
            {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions */}
            <div className={blockStyles.monacoWrapper} onKeyDown={event => event.stopPropagation()}>
                <MonacoEditor
                    language="markdown"
                    value={input}
                    height={150}
                    isLightTheme={isLightTheme}
                    editorWillMount={() => {}}
                    onEditorCreated={setEditor}
                    options={MONACO_BLOCK_INPUT_OPTIONS}
                    border={false}
                />
            </div>
        </div>
    )
}
